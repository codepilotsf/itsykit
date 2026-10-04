package appmanager

import (
	"context"
	"net/http"
	"os"
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
