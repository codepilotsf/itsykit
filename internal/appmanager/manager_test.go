package appmanager

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	first, err := Start(ctx, Spec{Name: "contact", Bundle: true, Capability: handler}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := Start(ctx, Spec{Name: "contact", Bundle: true, Capability: handler}, nil)
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

// TestFreezeExecutableIdentity proves a custom executable is launched only from
// a private frozen copy that matches the identity recorded at load, and that a
// file changed after validation is rejected.
func TestFreezeExecutableIdentity(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "app")
	body := []byte("#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(src, body, 0500); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	spec := Spec{Name: "demo", Executable: src, Hash: sum, Size: int64(len(body))}
	frozen, release, err := executableFor(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if frozen == src {
		t.Fatal("launched the source path instead of a frozen copy")
	}
	info, err := os.Stat(frozen)
	if err != nil || info.Mode().Perm() != 0500 {
		t.Fatalf("frozen mode %v %v", info, err)
	}
	got, err := os.ReadFile(frozen)
	if err != nil || !bytes.Equal(got, body) {
		t.Fatal("frozen bytes differ")
	}
	release()
	if _, err := os.Stat(frozen); !os.IsNotExist(err) {
		t.Fatal("frozen copy not cleaned up")
	}

	if err := os.Chmod(src, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("#!/bin/sh\nexit 1\n"), 0500); err != nil {
		t.Fatal(err)
	}
	if _, rel, err := executableFor(context.Background(), spec); err == nil {
		rel()
		t.Fatal("an executable changed after validation was accepted")
	}
}

func TestPrepareStorageIsPrivate(t *testing.T) {
	anchor := t.TempDir()
	runtimeRoot := filepath.Join(anchor, "flea", "runtime")
	storage := filepath.Join(runtimeRoot, "apps", "demo")
	if err := prepareStorage(Spec{Name: "demo", RuntimeRoot: runtimeRoot, Storage: storage}); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{runtimeRoot, filepath.Join(runtimeRoot, "apps"), storage} {
		info, err := os.Stat(dir)
		if err != nil || info.Mode().Perm() != 0700 {
			t.Fatalf("storage mode %v %v", dir, err)
		}
	}
}

// TestPrepareStorageRejectsSymlinkedAncestor proves a symlinked component inside
// the site cannot redirect app state outside the managed runtime.
func TestPrepareStorageRejectsSymlinkedAncestor(t *testing.T) {
	anchor := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(anchor, "flea")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	runtimeRoot := filepath.Join(anchor, "flea", "runtime")
	storage := filepath.Join(runtimeRoot, "apps", "demo")
	if err := prepareStorage(Spec{Name: "demo", RuntimeRoot: runtimeRoot, Storage: storage}); err == nil {
		t.Fatal("symlinked ancestor accepted")
	}
	if _, err := os.Stat(filepath.Join(outside, "runtime")); !os.IsNotExist(err) {
		t.Fatal("state was created through the symlink")
	}
}

// TestPrepareStorageRejectsPermissiveDirs proves an existing runtime or app
// directory that is accessible to other users is rejected rather than chmodded.
func TestPrepareStorageRejectsPermissiveDirs(t *testing.T) {
	anchor := t.TempDir()
	runtimeRoot := filepath.Join(anchor, "flea", "runtime")
	storage := filepath.Join(runtimeRoot, "apps", "demo")
	if err := os.MkdirAll(storage, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(runtimeRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := prepareStorage(Spec{Name: "demo", RuntimeRoot: runtimeRoot, Storage: storage}); err == nil {
		t.Fatal("permissive runtime directory accepted")
	}
	if info, err := os.Stat(runtimeRoot); err == nil && info.Mode().Perm() != 0755 {
		t.Fatal("existing user directory was chmodded")
	}
}

// TestPrepareStorageRejectsWritableContainer proves the administrator-owned
// container directory must not be writable by other users.
func TestPrepareStorageRejectsWritableContainer(t *testing.T) {
	anchor := t.TempDir()
	if err := os.Mkdir(filepath.Join(anchor, "flea"), 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(anchor, "flea"), 0777); err != nil {
		t.Fatal(err)
	}
	runtimeRoot := filepath.Join(anchor, "flea", "runtime")
	storage := filepath.Join(runtimeRoot, "apps", "demo")
	if err := prepareStorage(Spec{Name: "demo", RuntimeRoot: runtimeRoot, Storage: storage}); err == nil {
		t.Fatal("world-writable container accepted")
	}
}

// TestFreezeExecutableRejectsHardlink proves the frozen-copy identity check also
// refuses a hardlinked source, whose bytes could change through another name.
func TestFreezeExecutableRejectsHardlink(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "app")
	body := []byte("#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(src, body, 0500); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(src, filepath.Join(dir, "second")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	sum := sha256.Sum256(body)
	spec := Spec{Name: "demo", Executable: src, Hash: sum, Size: int64(len(body))}
	if frozen, release, err := executableFor(context.Background(), spec); err == nil {
		release()
		t.Fatalf("hardlinked executable frozen at %s", frozen)
	}
}

func TestRestartCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{}
	done := make(chan struct{})
	go func() {
		m.run(ctx, Spec{Name: "contact", Executable: "/nonexistent-flea-contact"}, "/nonexistent-flea-contact", nil)
		close(done)
	}()
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
	go func() {
		defer close(done)
		m.run(ctx, Spec{Name: "contact", Executable: executable}, executable, log.New(&logs, "", 0))
	}()
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
	go func() {
		defer close(done)
		m.run(ctx, Spec{Name: "contact", Executable: executable}, executable, log.New(logs, "", 0))
	}()
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
