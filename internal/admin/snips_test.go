package admin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"itsy/internal/editor"
	"itsy/internal/editstore"
	"itsy/internal/host"
	"itsy/internal/release"
	"itsy/internal/site"
	"itsy/internal/snips"
)

func richArchive(t *testing.T, include bool, css ...string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "itsy"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "site"), 0755); err != nil {
		t.Fatal(err)
	}
	store := snips.NewStore()
	store.Page("/").Title = "Local title"
	store.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Label: "title", Content: "Local seed"}
	data, _ := store.Bytes()
	if err := os.WriteFile(filepath.Join(root, "itsy/snips.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".itsy.config.json"), []byte(`{"app":{"site_name":"Test"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	page := `<p>No region</p>`
	if include {
		page = `<h1>{snip a7Km2pQ9xB}</h1>`
	}
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(page), 0644); err != nil {
		t.Fatal(err)
	}
	if len(css) > 0 {
		if err := os.MkdirAll(filepath.Join(root, "site", "css"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "site", "css", "theme.css"), []byte(css[0]), 0644); err != nil {
			t.Fatal(err)
		}
	}
	archive := filepath.Join(t.TempDir(), "site.tar.gz")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err = release.Pack(root, f); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return archive
}

func TestDeploymentPreservesAndArchivesOwnerSnips(t *testing.T) {
	m, o := fixture(t)
	o.Archive = richArchive(t, true)
	ctx := context.Background()
	if _, err := m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	oldPath := c.Content["example.test"]
	if filepath.Dir(oldPath) == c.Sites["example.test"] {
		t.Fatal("content stored inside release")
	}
	backend := &editstore.Store{Root: c.Sites["example.test"], ContentPath: oldPath, ConfigPath: m.ConfigPath, Publish: func(*site.Project) {}}
	doc, err := backend.Read()
	if err != nil {
		t.Fatal(err)
	}
	e := snips.Entry{Content: "Owner <em>edit</em>"}
	if err = backend.Save(editor.Update{Version: doc.Version, ID: "a7Km2pQ9xB", Content: e.Content}); err != nil {
		t.Fatal(err)
	}
	doc, err = backend.Read()
	if err != nil {
		t.Fatal(err)
	}
	if err = backend.Save(editor.Update{Version: doc.Version, Route: "/", Title: "Owner title"}); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	c, _, err = host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	current, err := snips.ReadFile(c.Content["example.test"])
	if err != nil || current.Pages["/"].Snips["a7Km2pQ9xB"].Content != e.Content || current.Pages["/"].Title != "Owner title" {
		t.Fatal(current, err)
	}
	previous := c.Content["example.test"]
	o.Archive = richArchive(t, false)
	originalVerify := m.Verify
	m.Verify = func(context.Context, host.Config, []string) error { return errors.New("failed health check") }
	if _, err = m.Apply(ctx, o); err == nil {
		t.Fatal("expected activation failure")
	}
	// Failed verification may leave a recovery journal if both verifications fail.
	m.Verify = originalVerify
	if _, err := os.Stat(m.pending()); err == nil {
		if err = m.Recover(ctx); err != nil {
			t.Fatal(err)
		}
	}
	c, _, err = host.Load(m.ConfigPath, nil)
	if err != nil || c.Content["example.test"] != previous {
		t.Fatal(c, err)
	}
	current, err = snips.ReadFile(previous)
	if err != nil || current.Pages["/"].Snips["a7Km2pQ9xB"].Content != e.Content {
		t.Fatal(current, err)
	}
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	c, _, err = host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	current, err = snips.ReadFile(c.Content["example.test"])
	if err != nil {
		t.Fatal(err)
	}
	active, _ := current.Index()
	if _, ok := active["a7Km2pQ9xB"]; ok || current.Archive["a7Km2pQ9xB"].Content != e.Content {
		t.Fatal(current)
	}
	o.Archive = richArchive(t, true)
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	c, _, _ = host.Load(m.ConfigPath, nil)
	current, err = snips.ReadFile(c.Content["example.test"])
	if err != nil || current.Pages["/"].Snips["a7Km2pQ9xB"].Content != e.Content {
		t.Fatal(current, err)
	}
}
