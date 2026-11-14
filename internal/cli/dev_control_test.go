package cli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"flea/internal/site"
)

func TestDevControlReplacementAndIsolation(t *testing.T) {
	root, other := t.TempDir(), t.TempDir()
	ctx, release, err := beginDevelopment(context.Background(), root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	otherCtx, otherRelease, err := beginDevelopment(context.Background(), other, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer otherRelease()
	go func() { <-ctx.Done(); release() }()
	var output bytes.Buffer
	nextCtx, nextRelease, err := beginDevelopment(context.Background(), root, &output)
	if err != nil {
		t.Fatal(err)
	}
	defer nextRelease()
	if ctx.Err() == nil || nextCtx.Err() != nil || otherCtx.Err() != nil {
		t.Fatal("replacement must stop only the previous server for this site")
	}
	if context.Cause(ctx) != errDevReplaced {
		t.Fatal("replacement must request draining, not ordinary forced shutdown")
	}
	if !strings.Contains(output.String(), "Stopping the previous dev server") {
		t.Fatal(output.String())
	}
	if unlock, err := site.LockDevelopment(root); err == nil {
		unlock()
		t.Fatal("replacement lost the site lock")
	}
}

func TestDevControlDoesNotInterruptPullOrReset(t *testing.T) {
	root := t.TempDir()
	unlock, err := site.LockDevelopment(root)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	_, release, err := beginDevelopment(context.Background(), root, io.Discard)
	if err == nil {
		release()
		t.Fatal("dev interrupted a non-dev operation")
	}
	if !strings.Contains(err.Error(), "pull, reset") {
		t.Fatal(err)
	}
}

func TestDevControlWaitsForPreviousCleanup(t *testing.T) {
	root := t.TempDir()
	oldCtx, oldRelease, err := beginDevelopment(context.Background(), root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer oldRelease()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	type result struct {
		release func()
		err     error
	}
	ready := make(chan result, 1)
	go func() {
		_, release, err := beginDevelopment(ctx, root, io.Discard)
		ready <- result{release, err}
	}()
	select {
	case <-oldCtx.Done():
	case <-ctx.Done():
		t.Fatal("old server never received stop request")
	}
	select {
	case r := <-ready:
		if r.release != nil {
			r.release()
		}
		t.Fatalf("replacement started before old requests/watcher finished: %v", r.err)
	case <-time.After(100 * time.Millisecond):
	}
	oldRelease()
	r := <-ready
	if r.err != nil {
		t.Fatal(r.err)
	}
	r.release()
}

func TestDevControlStaleSocketAndUnexpectedFile(t *testing.T) {
	root := t.TempDir()
	address, err := devControlPath(root)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: address, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	listener.Close()
	_, release, err := beginDevelopment(context.Background(), root, io.Discard)
	if err != nil {
		t.Fatal("stale socket blocked startup:", err)
	}
	release()
	if err := os.WriteFile(address, []byte("preserve me"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(address)
	if _, release, err := beginDevelopment(context.Background(), root, io.Discard); err == nil {
		release()
		t.Fatal("replaced unexpected file")
	}
	if data, _ := os.ReadFile(address); string(data) != "preserve me" {
		t.Fatal("file changed")
	}
	unlock, err := site.LockDevelopment(root)
	if err != nil {
		t.Fatal("failed startup leaked lock:", err)
	}
	unlock()
}

func TestDevControlRejectsInvalidRequest(t *testing.T) {
	root := t.TempDir()
	ctx, release, err := beginDevelopment(context.Background(), root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	address, _ := devControlPath(root)
	conn, err := net.Dial("unix", address)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	io.WriteString(conn, "nope\n")
	io.ReadAll(conn)
	if ctx.Err() != nil {
		t.Fatal("invalid request stopped dev")
	}
}

func TestDevExecutableReplacesServer(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "itsy")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "../../cmd/flea")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	type running struct {
		cmd     *exec.Cmd
		done    chan error
		address string
	}
	start := func(args ...string) running {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		t.Cleanup(cancel)
		cmd := exec.CommandContext(ctx, binary, append([]string{"dev"}, args...)...)
		cmd.Dir = root
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		addresses := make(chan string, 1)
		go func() {
			scanner := bufio.NewScanner(pipe)
			for scanner.Scan() {
				if strings.HasPrefix(scanner.Text(), "Website  ") {
					addresses <- strings.TrimPrefix(scanner.Text(), "Website  ")
				}
			}
			done <- cmd.Wait()
		}()
		t.Cleanup(func() { _ = cmd.Process.Signal(os.Interrupt) })
		select {
		case address := <-addresses:
			return running{cmd, done, address}
		case err := <-done:
			t.Fatalf("dev exited before listening: %v", err)
		case <-ctx.Done():
			t.Fatal("dev did not start")
		}
		return running{}
	}
	first := start("--port", "0")
	parsed, _ := url.Parse(first.address)
	args := []string{"--port", parsed.Port()}
	if _, err := devLANAddresses(); err == nil {
		args = append(args, "-l")
	}
	second := start(args...)
	select {
	case err := <-first.done:
		if err != nil {
			t.Fatal("previous dev did not exit cleanly:", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("previous dev is still running")
	}
	response, err := (&http.Client{Timeout: 3 * time.Second}).Get(second.address)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal(response.Status)
	}
	// A killed process leaves a socket behind but the OS must release the lock.
	second.cmd.Process.Kill()
	<-second.done
	third := start("--port", parsed.Port())
	third.cmd.Process.Signal(os.Interrupt)
	if err := <-third.done; err != nil {
		t.Fatal(err)
	}
}
