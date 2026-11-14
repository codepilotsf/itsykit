package admin

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/backup"
	"itsy/internal/host"
	"itsy/internal/media"
	"itsy/internal/release"
	"itsy/internal/site"
	"itsy/internal/snips"
)

func blogArchive(t *testing.T, title string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range map[string]string{"itsy.config.json": `{"app":{"site_name":"Blog"},"apps":{"blog":{"mount":"/blog"}}}`, "site/index.html": "Home", "site/blog/_index.html": `{#each blog.data.posts as post}{post.title}{/each}`, "site/blog/_post.html": `{blog.data.post.body}`} {
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var raw bytes.Buffer
	if err := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	photo, processed, err := media.Process(raw.Bytes(), "Cover", "A cover")
	if err != nil {
		t.Fatal(err)
	}
	s := snips.NewStore()
	s.Media = map[string]media.Photo{"Bbcdef1234": photo}
	s.Blog = []snips.BlogPost{{ID: "Abcdef1234", Title: title, Slug: "post", Status: "published", PublishedAt: "2026-10-13T12:00:00Z", Body: "<p>Body</p>", Cover: "Bbcdef1234"}}
	if err := os.MkdirAll(filepath.Join(root, "itsy", "media"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "itsy", "media", photo.File), processed, 0644); err != nil {
		t.Fatal(err)
	}
	if err := site.SaveContent(root, s); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "blog.tar.gz")
	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err = release.Pack(root, f); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return archive
}
func TestBlogDeployPreservationBackupAndRestore(t *testing.T) {
	m, o := fixture(t)
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = blogArchive(t, "Source seed")
	ctx := context.Background()
	if _, err := m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	config, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	live, err := snips.ReadFile(config.Content["example.test"])
	if err != nil {
		t.Fatal(err)
	}
	live.Blog[0].Title = "Live admin edit"
	live.Blog[0].Status = "draft"
	if err = snips.WriteFile(config.Content["example.test"], live); err != nil {
		t.Fatal(err)
	}
	o.Archive = blogArchive(t, "Changed source")
	o.Commit = strings.Repeat("b", 40)
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	snapshot, err := m.Snapshot(o.Name, o.Branch)
	if err != nil || snapshot.Snips.Blog[0].Title != "Live admin edit" || snapshot.Snips.Blog[0].Status != "draft" {
		t.Fatal(snapshot, err)
	}
	var output bytes.Buffer
	if err = m.Backup(o.Name, o.Branch, &output); err != nil {
		t.Fatal(err)
	}
	extracted := t.TempDir()
	if _, err = backup.Extract(bytes.NewReader(output.Bytes()), extracted); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(extracted)
	if err != nil || p.Content.Blog[0].Title != "Live admin edit" || len(p.Content.Media) != 1 {
		t.Fatal(p, err)
	}
	archive := filepath.Join(t.TempDir(), "blog.flea-backup")
	if err = os.WriteFile(archive, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	config, _, err = host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	live, err = snips.ReadFile(config.Content["example.test"])
	if err != nil {
		t.Fatal(err)
	}
	live.Blog[0].Title = "Later edit"
	if err = snips.WriteFile(config.Content["example.test"], live); err != nil {
		t.Fatal(err)
	}
	if _, err = m.Restore(ctx, o.Name, o.Branch, archive, snapshot.Version()); err == nil {
		t.Fatal("stale restore accepted")
	}
	current, err := m.Snapshot(o.Name, o.Branch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Restore(ctx, o.Name, o.Branch, archive, current.Version()); err != nil {
		t.Fatal(err)
	}
	restored, err := m.Snapshot(o.Name, o.Branch)
	if err != nil || restored.Snips.Blog[0].Title != "Live admin edit" {
		t.Fatal(restored, err)
	}
}
