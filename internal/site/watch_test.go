package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestDirectoryWatchLifecycle(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Watch(ctx, root, nil) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("watch did not stop")
		}
	})
	wait := func(path, want string, status int) {
		t.Helper()
		deadline := time.Now().Add(4 * time.Second)
		for time.Now().Before(deadline) {
			got := request(s, "GET", path)
			if got.Code == status && strings.Contains(got.Body.String(), want) {
				return
			}
			time.Sleep(15 * time.Millisecond)
		}
		t.Fatalf("watch did not update %s to %d / %q", path, status, want)
	}
	// A complete subtree created before its watch is installed must be discovered.
	put(t, root, "nested/deep/index.html", "new subtree")
	wait("/nested/deep", "new subtree", 200)
	put(t, root, "nested/deep/index.html", "nested edit")
	wait("/nested/deep", "nested edit", 200)
	// Editors commonly save to an ignored temporary file and rename over the page.
	put(t, root, "nested/deep/.save", "atomic replacement")
	if err := os.Rename(filepath.Join(root, "nested/deep/.save"), filepath.Join(root, "nested/deep/index.html")); err != nil {
		t.Fatal(err)
	}
	wait("/nested/deep", "atomic replacement", 200)
	if err := os.Rename(filepath.Join(root, "nested"), filepath.Join(root, "renamed")); err != nil {
		t.Fatal(err)
	}
	wait("/renamed/deep", "atomic replacement", 200)
	wait("/nested/deep", "", 404)
	put(t, root, "renamed/deep/index.html", "after rename")
	wait("/renamed/deep", "after rename", 200)
	if err := os.RemoveAll(filepath.Join(root, "renamed")); err != nil {
		t.Fatal(err)
	}
	wait("/renamed/deep", "", 404)
	put(t, root, "renamed/deep/index.html", "recreated")
	wait("/renamed/deep", "recreated", 200)
	// Extraction-generated writes must settle instead of creating a reload loop.
	put(t, root, "rich.html", `<h1 flea="Heading">Extract me</h1>`)
	wait("/rich", "Extract me", 200)
	time.Sleep(400 * time.Millisecond)
	s.mu.RLock()
	revision := s.revision
	s.mu.RUnlock()
	put(t, root, ".git/ignored", "ignored")
	put(t, root, "node_modules/ignored/index.html", "ignored")
	time.Sleep(400 * time.Millisecond)
	s.mu.RLock()
	final := s.revision
	s.mu.RUnlock()
	if final != revision {
		t.Fatalf("ignored or self-generated events triggered reloads: %d -> %d", revision, final)
	}
}

func TestWatchRegistrationExcludesPrivateMetadataAndLinks(t *testing.T) {
	root := fixture(t)
	put(t, root, "node_modules/pkg/index.html", "ignored")
	put(t, root, ".git/objects/object", "ignored")
	put(t, root, "nested/page.html", "included")
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err := syncWatches(w, root); err != nil {
		t.Fatal(err)
	}
	for _, name := range w.WatchList() {
		if strings.Contains(name, "node_modules") || strings.Contains(name, ".git") || strings.HasSuffix(name, "linked") {
			t.Fatal("excluded directory watched", name)
		}
	}
}

func TestWatchMissingRootFails(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	if err := s.Watch(context.Background(), filepath.Join(root, "absent"), nil); err == nil {
		t.Fatal("missing root silently accepted")
	}
}
