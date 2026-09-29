package admin

import (
	"context"
	"flea/internal/host"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncVersionAndOwnerPreservation(t *testing.T) {
	m, o := fixture(t)
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = richArchive(t, true)
	if _, e := m.Apply(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	before, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Snapshot(o.Name, "staging"); e == nil {
		t.Fatal("cross-branch export")
	}
	c, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	current, e := snips.ReadFile(c.Content["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	entry := current.Page("/").Snips["a7Km2pQ9xB"]
	entry.Content = "Owner edit"
	current.Page("/").Snips["a7Km2pQ9xB"] = entry
	if e = snips.WriteFile(c.Content["example.test"], current); e != nil {
		t.Fatal(e)
	}
	o.Sync = &syncstate.Plan{Expected: before.Version(), Snips: before.Snips}
	if _, e = m.Apply(context.Background(), o); e == nil {
		t.Fatal("accepted stale reviewed plan")
	}
	fresh, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if fresh.Snips.Page("/").Snips["a7Km2pQ9xB"].Content != "Owner edit" {
		t.Fatal("lost concurrent edit")
	}
	css := ":root {--brand-1: #123456;}"
	o.Sync = &syncstate.Plan{Expected: fresh.Version(), WithTheme: true, Theme: &css}
	if _, e = m.Apply(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	o.Sync = nil
	if _, e = m.Apply(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	fresh, e = m.Snapshot(o.Name, o.Branch)
	if e != nil || fresh.Theme == nil || *fresh.Theme != css {
		t.Fatal(fresh.Theme, e)
	}
	if fresh.Snips.Page("/").Snips["a7Km2pQ9xB"].Content != "Owner edit" {
		t.Fatal("normal deploy replaced owner content")
	}
	c, _, e = host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(c.Themes["example.test"]); e != nil {
		t.Fatal(e)
	}
	if filepath.Dir(c.Themes["example.test"]) != filepath.Dir(c.Content["example.test"]) {
		t.Fatal("theme outside content directory")
	}
	if _, e = site.Load(c.Sites["example.test"]); e != nil {
		t.Fatal(e)
	}
}
