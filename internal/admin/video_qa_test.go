package admin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/backup"
	"itsy/internal/editor"
	"itsy/internal/editstore"
	"itsy/internal/host"
	"itsy/internal/site"
)

func TestQALiveVideoDeployBackupRestore(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = typedMediaArchive(t, "video")
	if _, e := m.Apply(ctx, o); e != nil {
		t.Fatal(e)
	}
	c, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	s := editstore.Store{Root: c.Sites["example.test"], ContentPath: c.Content["example.test"], ConfigPath: m.ConfigPath, Publish: func(*site.Project) {}}
	d, e := s.Read()
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile("../video/testdata/valid.mp4")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Save(editor.Update{Page: &editor.PageEdit{Version: d.Version, Route: "/", Media: []editor.MediaEdit{{ID: typedMediaID, Name: "Live clip", Data: data}}}}); e != nil {
		t.Fatal(e)
	}
	live, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	id := typedMediaID
	if live.Snips.Page("/").Snips[id].Video == nil {
		t.Fatal("video missing")
	}
	v := *live.Snips.Page("/").Snips[id].Video
	if _, e = m.Apply(ctx, o); e != nil {
		t.Fatal("deploy lost live video", e)
	}
	current, e := m.Snapshot(o.Name, o.Branch)
	if e != nil || *current.Snips.Page("/").Snips[id].Video != v {
		t.Fatal("deploy replaced video", e)
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
	if e != nil || *p.Content.Page("/").Snips[id].Video != v {
		t.Fatal("backup lost video", e)
	}
	file := filepath.Join(t.TempDir(), "video.backup")
	if e = os.WriteFile(file, b.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Restore(ctx, o.Name, o.Branch, file, current.Version()); e != nil {
		t.Fatal("restore failed", e)
	}
	restored, e := m.Snapshot(o.Name, o.Branch)
	if e != nil || *restored.Snips.Page("/").Snips[id].Video != v {
		t.Fatal("restore lost video", e)
	}
}
