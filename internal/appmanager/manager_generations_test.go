package appmanager

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"flea/internal/appapi"
)

// This file adds coverage for the case the failed-launch tests miss: a child
// that authenticates and is published as a usable Client before it exits. The
// supervisor must still replace the withdrawn Client and exhaust its five
// attempts.

// TestAppHelperChild is re-executed by readyContactScript as a real managed
// child. It reads its credentials from inherited fd 4, serves the authenticated
// Flea App Protocol handshake on inherited fd 3, records its exact PID and a
// nonce as ready, then exits on the matching trigger or parent-liveness EOF.
func TestAppHelperChild(t *testing.T) {
	if os.Getenv("FLEA_APP_HELPER") != "1" {
		t.Skip("not the app manager helper subprocess")
	}
	config := os.NewFile(4, "app-credentials")
	var launch struct{ Name, Token, Storage, CapabilityURL, CapabilityToken string }
	decoder := json.NewDecoder(io.LimitReader(config, 4097))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&launch)
	var extra any
	end := decoder.Decode(&extra)
	config.Close()
	if err != nil || end != io.EOF || len(launch.Token) != 64 {
		t.Fatalf("invalid app launch configuration: %v (tail %v)", err, end)
	}
	listenerFile := os.NewFile(3, "app-listener")
	listener, err := net.FileListener(listenerFile)
	listenerFile.Close()
	if err != nil {
		t.Fatalf("app listener is unavailable: %v", err)
	}
	defer listener.Close()
	server := &http.Server{ReadHeaderTimeout: 2 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "Bearer " + launch.Token
		if len(r.Header.Values("Authorization")) != 1 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(want)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/_flea/handshake" && r.URL.RawPath == "" && r.URL.RawQuery == "" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(appapi.Identity{Name: "contact", Version: "ready-helper", Protocol: appapi.Version})
			return
		}
		http.NotFound(w, r)
	})}
	go server.Serve(listener)

	readyPath := os.Getenv("FLEA_APP_READY")
	if readyPath == "" {
		t.Fatal("missing readiness path")
	}
	nonceBytes := make([]byte, 8)
	if _, err := rand.Read(nonceBytes); err != nil {
		t.Fatal(err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	marker, err := os.OpenFile(readyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintf(marker, "%d %s\n", os.Getpid(), nonce); err != nil {
		t.Fatal(err)
	}
	marker.Close()

	alive := os.NewFile(5, "app-parent")
	parentGone := make(chan struct{})
	go func() { _, _ = io.Copy(io.Discard, alive); close(parentGone) }()
	trigger := filepath.Join(filepath.Dir(readyPath), fmt.Sprintf("exit-%d-%s", os.Getpid(), nonce))
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(trigger); err == nil {
			return
		}
		select {
		case <-parentGone:
			return
		case <-ticker.C:
		}
	}
}

type readyChild struct {
	pid   int
	nonce string
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// readyContactScript writes a real executable in its own directory. It is a
// wrapper because the manager fixes the child's environment and arguments: the
// shell installs the helper marker and readiness path, then execs this test
// binary, so the managed PID is exactly the recording helper process.
func readyContactScript(t *testing.T) (string, string) {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	executable := filepath.Join(dir, "contact")
	readyPath := filepath.Join(dir, "ready")
	script := "#!/bin/sh\n" +
		"FLEA_APP_HELPER=1 FLEA_APP_READY=" + shellQuote(readyPath) +
		" exec " + shellQuote(binary) + " -test.run='^TestAppHelperChild$'\n"
	if err := os.WriteFile(executable, []byte(script), 0o500); err != nil {
		t.Fatal(err)
	}
	return executable, readyPath
}

func readReady(t *testing.T, path string) []readyChild {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	text := string(data)
	if !strings.HasSuffix(text, "\n") {
		// A child writes one newline-terminated record; ignore a trailing
		// partial write so waiting counts only complete readiness records.
		text = text[:strings.LastIndex(text, "\n")+1]
	}
	var children []readyChild
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		var child readyChild
		if _, err := fmt.Sscanf(line, "%d %s", &child.pid, &child.nonce); err != nil {
			t.Fatalf("malformed readiness record %q: %v", line, err)
		}
		children = append(children, child)
	}
	return children
}

func waitFor(t *testing.T, timeout time.Duration, what string, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if ready() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for %s", timeout, what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestRunReplacesReadyClientsOnChildExit pins the recovery path that the
// failed-launch tests never reach. Each child authenticates, is published as a
// usable Client, then exits; the supervisor must withdraw the stale Client,
// launch a distinct replacement, and after five generations exhaust and leave
// its routes unavailable. Because bundle state is shared, this test is
// deliberately not parallel.
func TestRunReplacesReadyClientsOnChildExit(t *testing.T) {
	executable, readyPath := readyContactScript(t)
	dir := filepath.Dir(readyPath)
	var logs lockedBuffer
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.run(ctx, Spec{Name: "contact", Executable: executable}, executable, log.New(&logs, "", 0))
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("supervisor did not stop after cancellation")
		}
	}()

	var (
		clients  []*appapi.Client
		children []readyChild
	)
	for generation := 0; generation < 5; generation++ {
		waitFor(t, 15*time.Second, fmt.Sprintf("ready client %d", generation), func() bool {
			client := m.Client()
			return client != nil && (len(clients) == 0 || client != clients[len(clients)-1])
		})
		client := m.Client()
		waitFor(t, 5*time.Second, fmt.Sprintf("readiness record %d", generation), func() bool {
			return len(readReady(t, readyPath)) >= generation+1
		})
		ready := readReady(t, readyPath)
		if len(ready) != generation+1 {
			t.Fatalf("generation %d: readiness records = %d, want %d", generation, len(ready), generation+1)
		}
		check, checkCancel := context.WithTimeout(ctx, 3*time.Second)
		if err := client.Handshake(check); err != nil {
			checkCancel()
			t.Fatalf("generation %d: published client is not usable: %v", generation, err)
		}
		checkCancel()
		clients = append(clients, client)
		children = append(children, ready[generation])

		trigger := filepath.Join(dir, fmt.Sprintf("exit-%d-%s", ready[generation].pid, ready[generation].nonce))
		if err := os.WriteFile(trigger, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		waitFor(t, 8*time.Second, fmt.Sprintf("withdrawal after generation %d", generation), func() bool {
			return m.Client() == nil
		})
		stale, staleCancel := context.WithTimeout(ctx, 2*time.Second)
		if err := client.Handshake(stale); err == nil {
			staleCancel()
			t.Fatalf("generation %d: withdrawn client still accepted a handshake", generation)
		}
		staleCancel()
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("supervisor did not exhaust its five attempts")
	}
	if m.Client() != nil {
		t.Fatal("an unavailable app published a client after exhaustion")
	}
	if got := len(readReady(t, readyPath)); got != 5 {
		t.Fatalf("ready generations after exhaustion = %d, want 5", got)
	}
	if !strings.Contains(logs.String(), "could not stay running") {
		t.Fatalf("missing exhaustion notice: %q", logs.String())
	}
	seen := map[int]bool{}
	for generation, child := range children {
		if seen[child.pid] {
			t.Errorf("generation %d reused child pid %d", generation, child.pid)
		}
		seen[child.pid] = true
		if err := syscall.Kill(child.pid, 0); !errors.Is(err, syscall.ESRCH) {
			t.Errorf("generation %d: expected child process %d to be gone, got %v", generation, child.pid, err)
		}
	}
	for generation := 1; generation < len(clients); generation++ {
		if clients[generation] == clients[generation-1] {
			t.Errorf("generation %d reused the previously withdrawn client", generation)
		}
	}
}
