package admin

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/backup"
	"flea/internal/host"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
)

func TestQABackupLiveOverlaysRestoreStaleAndRecovery(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = richArchive(t, true)
	if _, e := m.Apply(ctx, o); e != nil {
		t.Fatal(e)
	}
	live, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	css := ":root {--brand-1: #123456;}"
	o.Sync = &syncstate.Plan{Expected: live.Version(), WithTheme: true, Theme: &css}
	if _, e = m.Apply(ctx, o); e != nil {
		t.Fatal(e)
	}
	c, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	content, e := snips.ReadFile(c.Content["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	content.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "Saved owner text"}
	if e = snips.WriteFile(c.Content["example.test"], content); e != nil {
		t.Fatal(e)
	}
	var buf bytes.Buffer
	if e = m.Backup(o.Name, o.Branch, &buf); e != nil {
		t.Fatal(e)
	}
	extracted := t.TempDir()
	if _, e = backup.Extract(bytes.NewReader(buf.Bytes()), extracted); e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(extracted)
	if e != nil {
		t.Fatal(e)
	}
	if p.Content.Page("/").Snips["a7Km2pQ9xB"].Content != "Saved owner text" || p.ThemeCSS() == nil || *p.ThemeCSS() != css {
		t.Fatal("backup used stale release seeds")
	}
	file := filepath.Join(t.TempDir(), "site.flea-backup")
	if e = os.WriteFile(file, buf.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	reviewed, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	content.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "Newer live text"}
	if e = snips.WriteFile(c.Content["example.test"], content); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Restore(ctx, o.Name, o.Branch, file, reviewed.Version()); e == nil {
		t.Fatal("accepted stale restore")
	}
	live, e = m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Restore(ctx, o.Name, "staging", file, live.Version()); e == nil {
		t.Fatal("accepted cross-branch restore")
	}
	before, e := os.ReadFile(m.ConfigPath)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	m.Verify = func(context.Context, host.Config, []string) error {
		calls++
		if calls == 1 {
			return errors.New("probe failure")
		}
		return nil
	}
	if _, e = m.Restore(ctx, o.Name, o.Branch, file, live.Version()); e == nil {
		t.Fatal("accepted failed probe")
	}
	after, e := os.ReadFile(m.ConfigPath)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("failed restore changed live config", e)
	}
	current, e := m.Snapshot(o.Name, o.Branch)
	if e != nil || current.Version() != live.Version() {
		t.Fatal("failed restore changed live content", e)
	}
	if _, e = m.Restore(ctx, o.Name, o.Branch, file, live.Version()); e != nil {
		t.Fatal(e)
	}
	restored, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if restored.Snips.Page("/").Snips["a7Km2pQ9xB"].Content != "Saved owner text" || restored.Theme == nil || *restored.Theme != css {
		t.Fatal("restore did not restore backup overlays")
	}
}
