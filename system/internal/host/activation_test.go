package host

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A child test process receives real descriptors at 3/4, as it would from
// systemd. The parent retains the listeners through child shutdown and startup.
func TestActivationProcess(t *testing.T) {
	if os.Getenv("ITSY_ACTIVATION_TEST") != "1" {
		return
	}
	if os.Getenv("ITSY_SYSTEMD_TEST") != "1" && os.Getenv("ITSY_BAD_PID") != "1" {
		_ = os.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid()))
	}
	if delay, err := time.ParseDuration(os.Getenv("ITSY_START_DELAY")); err == nil {
		time.Sleep(delay)
	}
	httpPort, _ := strconv.Atoi(os.Getenv("ITSY_HTTP_PORT"))
	httpsPort, _ := strconv.Atoi(os.Getenv("ITSY_HTTPS_PORT"))
	l, err := listenInherited(httpPort, httpsPort)
	if os.Getenv("ITSY_EXPECT_ERROR") == "1" {
		if err == nil {
			l.close()
			t.Fatal("accepted invalid socket activation")
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer l.close()
	for _, name := range []string{"LISTEN_PID", "LISTEN_FDS", "LISTEN_FDNAMES"} {
		if os.Getenv(name) != "" {
			t.Fatal("activation environment leaked", name)
		}
	}
	c, h, err := Load(os.Getenv("ITSY_CONFIG"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if os.Getenv("ITSY_SYSTEMD_TEST") != "1" {
		go func() {
			_, _ = io.Copy(io.Discard, os.Stdin)
			cancel()
		}()
	}
	if err := serveWithListeners(ctx, c, h, log.New(os.Stdout, "", 0), nil, l); err != nil {
		t.Fatal(err)
	}
}

func activationFiles(t *testing.T, address string) (net.Listener, *os.File) {
	t.Helper()
	l, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	f, err := l.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return l, f
}

func activationCommand(files []*os.File, env ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestActivationProcess$")
	cmd.ExtraFiles = files
	cmd.Env = append(os.Environ(), "ITSY_ACTIVATION_TEST=1", "LISTEN_FDS=2", "LISTEN_PID=0", "LISTEN_FDNAMES=web:web")
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

func startActivated(t *testing.T, files []*os.File, env ...string) func() {
	t.Helper()
	cmd := activationCommand(files, env...)
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		_ = in.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			_ = cmd.Process.Kill()
			<-done
			t.Error("activated server shutdown timed out")
		}
	}
	t.Cleanup(stop)
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(out)
		if scanner.Scan() {
			ready <- strings.Contains(scanner.Text(), "ItsyKit server listening")
		} else {
			ready <- false
		}
		_, _ = io.Copy(io.Discard, out)
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("activated server did not start")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("activated server startup timed out")
	}
	return stop
}

func TestActivatedConnectionsSurviveRestart(t *testing.T) {
	c, _, pool := tlsFixture(t)
	c.Sites["plain.test"] = c.Sites["example.test"]
	config := filepath.Join(t.TempDir(), "server.json")
	data, _ := json.Marshal(c)
	if err := os.WriteFile(config, data, 0600); err != nil {
		t.Fatal(err)
	}
	httpListener, httpFile := activationFiles(t, ":0")
	httpsListener, httpsFile := activationFiles(t, ":0")
	env := []string{"ITSY_CONFIG=" + config, fmt.Sprint("ITSY_HTTP_PORT=", httpListener.Addr().(*net.TCPAddr).Port), fmt.Sprint("ITSY_HTTPS_PORT=", httpsListener.Addr().(*net.TCPAddr).Port)}
	files := []*os.File{httpsFile, httpFile} // Socket order must not matter.
	stop := startActivated(t, files, env...)
	stop()

	// Both TCP handshakes complete while no server process is running. HTTP
	// request bytes and the TLS ClientHello wait in the kernel until restart.
	httpConn, err := net.DialTimeout("tcp", httpListener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal("HTTP port closed during restart", err)
	}
	defer httpConn.Close()
	_ = httpConn.SetDeadline(time.Now().Add(10 * time.Second))
	_, _ = fmt.Fprint(httpConn, "GET / HTTP/1.1\r\nHost: plain.test\r\nConnection: close\r\n\r\n")
	httpsConn, err := net.DialTimeout("tcp", httpsListener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal("HTTPS port closed during restart", err)
	}
	defer httpsConn.Close()
	_ = httpsConn.SetDeadline(time.Now().Add(10 * time.Second))
	secure := tls.Client(httpsConn, &tls.Config{RootCAs: pool, ServerName: "example.test", MinVersion: tls.VersionTLS12})
	handshake := make(chan error, 1)
	go func() { handshake <- secure.Handshake() }()
	select {
	case err := <-handshake:
		t.Fatal("TLS did not wait for restart", err)
	case <-time.After(50 * time.Millisecond):
	}
	startActivated(t, files, env...)
	if err := <-handshake; err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprint(secure, "GET / HTTP/1.1\r\nHost: example.test\r\nConnection: close\r\n\r\n")
	for _, conn := range []net.Conn{httpConn, secure} {
		response, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != 200 || !strings.Contains(string(body), "Host test") {
			t.Fatalf("queued request failed: %v, %d, %s", err, response.StatusCode, body)
		}
	}
}

func TestActivationRejectsInvalidHandoff(t *testing.T) {
	wildcard, file := activationFiles(t, ":0")
	_, loopbackFile := activationFiles(t, "127.0.0.1:0")
	regular, err := os.CreateTemp(t.TempDir(), "not-a-socket")
	if err != nil {
		t.Fatal(err)
	}
	defer regular.Close()
	for _, tc := range []struct {
		name  string
		files []*os.File
		env   []string
	}{
		{"missing descriptors", nil, nil},
		{"wrong pid", []*os.File{file, file}, []string{"ITSY_BAD_PID=1"}},
		{"wrong count", []*os.File{file, file}, []string{"LISTEN_FDS=1"}},
		{"duplicate port", []*os.File{file, file}, nil},
		{"loopback", []*os.File{file, loopbackFile}, nil},
		{"regular file", []*os.File{regular, file}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := []string{"ITSY_EXPECT_ERROR=1", fmt.Sprint("ITSY_HTTP_PORT=", wildcard.Addr().(*net.TCPAddr).Port), "ITSY_HTTPS_PORT=443"}
			env = append(env, tc.env...)
			if out, err := activationCommand(tc.files, env...).CombinedOutput(); err != nil {
				t.Fatalf("%v: %s", err, out)
			}
		})
	}
}

func TestActivatedHTTPOnlyRejectsHTTPS(t *testing.T) {
	config := testConfig(t, ":80")
	httpListener, httpFile := activationFiles(t, ":0")
	httpsListener, httpsFile := activationFiles(t, ":0")
	startActivated(t, []*os.File{httpFile, httpsFile}, "ITSY_CONFIG="+config,
		fmt.Sprint("ITSY_HTTP_PORT=", httpListener.Addr().(*net.TCPAddr).Port),
		fmt.Sprint("ITSY_HTTPS_PORT=", httpsListener.Addr().(*net.TCPAddr).Port))
	dialer := &net.Dialer{Timeout: time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", httpsListener.Addr().String(), &tls.Config{ServerName: "example.test"})
	if err == nil {
		conn.Close()
		t.Fatal("HTTPS accepted without configuration")
	}
	if strings.Contains(err.Error(), "timeout") {
		t.Fatal("unconfigured HTTPS was left queued", err)
	}
}
