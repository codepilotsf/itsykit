package admin

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/backup"
	"flea/internal/editor"
	"flea/internal/editstore"
	"flea/internal/host"
	"flea/internal/media"
	"flea/internal/release"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/video"
)

const typedMediaID = "Abcdefghij"

func typedMediaArchive(t *testing.T, kind string) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"site", "flea/media"} {
		if e := os.MkdirAll(filepath.Join(root, dir), 0755); e != nil {
			t.Fatal(e)
		}
	}
	store := snips.NewStore()
	entry := snips.Entry{Type: kind}
	var data []byte
	var filename, markup string
	if kind == "image" {
		var raw bytes.Buffer
		if e := png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 4, 2))); e != nil {
			t.Fatal(e)
		}
		photo, processed, e := media.Process(raw.Bytes(), "Seed", "Seed alt")
		if e != nil {
			t.Fatal(e)
		}
		entry.Image = &photo
		data = processed
		filename = photo.File
		markup = `<img src="{snip ` + typedMediaID + `.src}" alt="{snip ` + typedMediaID + `.alt}">`
	} else {
		var e error
		data, e = os.ReadFile("../video/testdata/silent.mp4")
		if e != nil {
			t.Fatal(e)
		}
		metadata, e := video.Inspect(bytes.NewReader(data), int64(len(data)))
		if e != nil {
			t.Fatal(e)
		}
		entry.Video = &video.Video{Metadata: metadata, Name: "Seed"}
		filename = metadata.File
		markup = `<video src="{snip ` + typedMediaID + `.src}" poster="{snip ` + typedMediaID + `.poster}" controls></video>`
	}
	store.Page("/").Snips[typedMediaID] = entry
	for name, body := range map[string][]byte{".flea.config.json": []byte(`{"app":{"site_name":"Test"}}`), "site/index.html": []byte(markup), "flea/media/" + filename: data} {
		if e := os.WriteFile(filepath.Join(root, name), body, 0644); e != nil {
			t.Fatal(e)
		}
	}
	if e := site.SaveContent(root, store); e != nil {
		t.Fatal(e)
	}
	archive := filepath.Join(t.TempDir(), "site.tar.gz")
	f, e := os.Create(archive)
	if e != nil {
		t.Fatal(e)
	}
	if e = release.Pack(root, f); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	return archive
}

func TestQALivePhotosSurviveDeployBackupAndRestore(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = typedMediaArchive(t, "image")
	if _, e := m.Apply(ctx, o); e != nil {
		t.Fatal(e)
	}
	c, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	store := editstore.Store{Root: c.Sites["example.test"], ContentPath: c.Content["example.test"], ConfigPath: m.ConfigPath, Publish: func(*site.Project) {}}
	d, e := store.Read()
	if e != nil {
		t.Fatal(e)
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var raw bytes.Buffer
	if e = png.Encode(&raw, img); e != nil {
		t.Fatal(e)
	}
	if e = store.Save(editor.Update{Page: &editor.PageEdit{Version: d.Version, Route: "/", Media: []editor.MediaEdit{{ID: typedMediaID, Data: raw.Bytes(), Name: "Live photo", Alt: "Owner upload"}}}}); e != nil {
		t.Fatal(e)
	}
	live, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if len(live.Snips.Page("/").Snips) != 1 {
		t.Fatal("missing uploaded photo")
	}
	id := typedMediaID
	photo := *live.Snips.Page("/").Snips[id].Image
	if _, e = m.Apply(ctx, o); e != nil {
		t.Fatal("normal code deployment lost live photo", e)
	}
	after, e := m.Snapshot(o.Name, o.Branch)
	if e != nil || *after.Snips.Page("/").Snips[id].Image != photo {
		t.Fatal("deploy changed photo", e)
	}
	var b bytes.Buffer
	if e = m.Backup(o.Name, o.Branch, &b); e != nil {
		t.Fatal(e)
	}
	stage := t.TempDir()
	if _, e = backup.Extract(bytes.NewReader(b.Bytes()), stage); e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(stage)
	if e != nil || *p.Content.Page("/").Snips[id].Image != photo {
		t.Fatal("backup omitted photo", e)
	}
	if _, e = site.ReadPhoto(filepath.Join(stage, "flea", "media"), photo.File); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(t.TempDir(), "photo.backup")
	if e = os.WriteFile(file, b.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Restore(ctx, o.Name, o.Branch, file, after.Version()); e != nil {
		t.Fatal(e)
	}
	restored, e := m.Snapshot(o.Name, o.Branch)
	if e != nil || *restored.Snips.Page("/").Snips[id].Image != photo {
		t.Fatal("restore lost photo", e)
	}
}
