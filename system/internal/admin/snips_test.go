package admin

import (
	"os"
	"path/filepath"
	"testing"

	"itsy/internal/release"
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
	store.Blog = []snips.BlogPost{{ID: "a7Km2pQ9xB", Title: "Local title", Slug: "post", Body: "Local seed", Status: "draft"}}
	data, _ := store.Bytes()
	if err := os.WriteFile(filepath.Join(root, "itsy/snips.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "itsy.config.json"), []byte(`{"app":{"site_name":"Test"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	page := `<p>No region</p>`
	if include {
		page = `<h1>Ordinary HTML</h1>`
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
