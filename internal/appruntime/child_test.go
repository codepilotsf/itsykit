package appruntime

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"flea/internal/appbundle"
)

func TestChildAuthenticationAndParentExit(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing bundle")
		}
		t.Skip("requires scripts/build.sh")
	}
	executable := filepath.Join(t.TempDir(), "contact")
	if err := appbundle.Extract(context.Background(), executable); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	inherited, err := listener.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	defer inherited.Close()
	configRead, configWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer configRead.Close()
	defer configWrite.Close()
	aliveRead, aliveWrite, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer aliveRead.Close()
	defer aliveWrite.Close()
	token, capToken := strings.Repeat("a", 64), strings.Repeat("b", 64)
	callback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != CapabilityPath || r.Header.Get("Authorization") != "Bearer "+capToken {
			t.Error("incorrect capability auth")
			http.Error(w, "unauthorized", 401)
			return
		}
		if r.Header.Get("Cookie") != "" {
			t.Error("cookie leaked")
		}
		w.WriteHeader(204)
	}))
	defer callback.Close()
	cmd := exec.Command(executable)
	cmd.Env = []string{"LANG=C"}
	cmd.ExtraFiles = []*os.File{inherited, configRead, aliveRead}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	defer func() { cmd.Process.Kill() }()
	configRead.Close()
	aliveRead.Close()
	inherited.Close()
	listener.Close()
	if err := json.NewEncoder(configWrite).Encode(launchConfig{Token: token, CapabilityURL: callback.URL, CapabilityToken: capToken}); err != nil {
		t.Fatal(err)
	}
	configWrite.Close()
	client := &http.Client{Timeout: time.Second}
	defer client.CloseIdleConnections()
	call := func(path, auth string, duplicate bool) (int, string, error) {
		r, _ := http.NewRequest("GET", "http://"+listener.Addr().String()+path, nil)
		if auth != "" {
			r.Header.Set("Authorization", auth)
		}
		if duplicate {
			r.Header.Add("Authorization", auth)
		}
		response, err := client.Do(r)
		if err != nil {
			return 0, "", err
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		return response.StatusCode, string(body), err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		code, _, err := call("/_flea/handshake", "Bearer "+token, false)
		if err == nil && code == 200 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("handshake unavailable", code, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	for _, auth := range []string{"", "Bearer " + capToken, "Bearer wrong"} {
		if code, _, err := call("/_flea/handshake", auth, false); err != nil || code != 401 {
			t.Fatal("unauthorized handshake", code, err)
		}
	}
	if code, _, err := call("/_flea/handshake", "Bearer "+token, true); err != nil || code != 401 {
		t.Fatal("duplicate auth", code, err)
	}
	r, _ := http.NewRequest("POST", "http://"+listener.Addr().String()+"/", strings.NewReader(`{"name":"Visitor","email":"visitor@example.test","message":"Hello"}`))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	response, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal(response.StatusCode)
	}
	// Closing the only parent writer simulates parent death, without a signal.
	aliveWrite.Close()
	select {
	case err := <-exited:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("child survived parent liveness EOF")
	}
}
