package admin

import (
	"context"
	"itsy/internal/host"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
	"strings"
	"testing"
)

func TestSyncVersionAndOwnerPreservation(t *testing.T) {
	m, o := fixture(t)
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = richArchive(t, true, ":root {--brand-1: #111111;}")
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
	entry := current.Blog[0]
	entry.Body = "Owner edit"
	current.Blog[0] = entry
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
	if fresh.Snips.Blog[0].Body != "Owner edit" {
		t.Fatal("lost concurrent edit")
	}
	css := ":root {--brand-1: #123456;}"
	o.Archive = richArchive(t, true, css)
	o.Sync = nil
	if _, e = m.Apply(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	fresh, e = m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if fresh.Snips.Blog[0].Body != "Owner edit" {
		t.Fatal("normal deploy replaced owner content")
	}
	c, _, e = host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(c.Sites["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	if string(p.AssetData["/css/theme.css"]) != css {
		t.Fatal("committed CSS was not published")
	}
	// A source release can also remove CSS; no prior overlay should reappear.
	o.Archive = richArchive(t, true)
	if _, e = m.Apply(context.Background(), o); e != nil {
		t.Fatal(e)
	}
	c, _, e = host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	p, e = site.Load(c.Sites["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	if _, exists := p.Assets["/css/theme.css"]; exists {
		t.Fatal("removed source CSS was retained")
	}
}
