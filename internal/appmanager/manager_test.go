package appmanager

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"flea/internal/appbundle"
)

func TestManagedLifecycle(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing bundle")
		}
		t.Skip("requires scripts/build.sh")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "not admitted", 403) })
	first, err := Start(ctx, handler, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := Start(ctx, handler, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	deadline := time.Now().Add(8 * time.Second)
	for first.Client() == nil || second.Client() == nil {
		if time.Now().After(deadline) {
			t.Fatal("handshake timed out")
		}
		time.Sleep(20 * time.Millisecond)
	}
	bundles.Lock()
	refs, path := bundles.refs, bundles.path
	bundles.Unlock()
	if refs != 2 {
		t.Fatal("executable not shared", refs)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0500 {
		t.Fatal("private executable", err)
	}
	first.Close()
	if first.Client() != nil || second.Client() == nil {
		t.Fatal("closing one app affected another")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("executable removed while in use")
	}
	second.Close()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("executable not cleaned up", err)
	}
	bundles.Lock()
	refs = bundles.refs
	bundles.Unlock()
	if refs != 0 {
		t.Fatal("bundle ref leak")
	}
}

func TestRestartCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{}
	done := make(chan struct{})
	go func() { m.run(ctx, "/nonexistent-flea-contact", http.NotFoundHandler(), nil); close(done) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("backoff ignored cancellation")
	}
}

// failingContactScript writes a real executable that records one byte per launch
// and exits immediately, so every managed attempt fails at the first handshake.
// It returns the executable path and a function reporting the launch count.
func failingContactScript(t *testing.T) (string, func() int) {
	t.Helper()
	dir := t.TempDir()
	// The manager runs children with this directory as their working directory,
	// so the script can append to a relative marker without shell quoting.
	script := filepath.Join(dir, "contact")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf x >> launches\nexit 0\n"), 0500); err != nil {
		t.Fatal(err)
	}
	launches := filepath.Join(dir, "launches")
	return script, func() int {
		data, err := os.ReadFile(launches)
		if err != nil {
			if os.IsNotExist(err) {
				return 0
			}
			t.Fatal(err)
		}
		return len(data)
	}
}

type lockedBuffer struct {
	mu         sync.Mutex
	buf        bytes.Buffer
	restarting chan struct{}
	once       sync.Once
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.restarting != nil && bytes.Contains(p, []byte("Restarting its private service")) {
		b.once.Do(func() { close(b.restarting) })
	}
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestRunExhaustsProcessAttempts pins the bounded restart policy: a child that
// keeps failing is launched exactly five times, after which the manager stops and
// leaves no client published so its routes stay unavailable but the site survives.
func TestRunExhaustsProcessAttempts(t *testing.T) {
	executable, launches := failingContactScript(t)
	var logs lockedBuffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := &Manager{}
	done := make(chan struct{})
	go func() { defer close(done); m.run(ctx, executable, http.NotFoundHandler(), log.New(&logs, "", 0)) }()
	select {
	case <-done:
	case <-time.After(35 * time.Second):
		t.Fatal("managed run did not exhaust its attempts")
	}
	if got := launches(); got != 5 {
		t.Fatalf("process attempts = %d, want exactly 5", got)
	}
	if m.Client() != nil {
		t.Fatal("an unavailable app published a client after exhaustion")
	}
	if !strings.Contains(logs.String(), "could not stay running") {
		t.Fatalf("missing exhaustion notice: %q", logs.String())
	}
}

// TestRunBackoffCancelsPromptly proves the retry delay is cancellable rather
// than a fixed sleep: cancelling during the first backoff must return well before
// the one second delay elapses.
func TestRunBackoffCancelsPromptly(t *testing.T) {
	executable, launches := failingContactScript(t)
	logs := &lockedBuffer{restarting: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := &Manager{}
	done := make(chan struct{})
	go func() { defer close(done); m.run(ctx, executable, http.NotFoundHandler(), log.New(logs, "", 0)) }()
	select {
	case <-logs.restarting:
	case <-time.After(5 * time.Second):
		t.Fatal("manager never entered retry backoff")
	}
	start := time.Now()
	cancel()
	select {
	case <-done:
	case <-time.After(900 * time.Millisecond):
		t.Fatal("backoff ignored cancellation")
	}
	if elapsed := time.Since(start); elapsed > 900*time.Millisecond {
		t.Fatalf("cancellation took %s", elapsed)
	}
	if m.Client() != nil {
		t.Fatal("an unavailable app published a client")
	}
	if got := launches(); got != 1 {
		t.Fatalf("cancellation allowed another launch: got %d", got)
	}
}
