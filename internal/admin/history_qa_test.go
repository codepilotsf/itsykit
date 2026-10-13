package admin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"flea/internal/host"
	"flea/internal/site"
	"flea/internal/snips"
)

func TestQARollbackPreservesLiveStateAndOtherSites(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	if err := os.WriteFile(m.ConfigPath, []byte(`{"listen":":80","sites":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Archive = richArchive(t, true)
	o.Domains = []string{"example.test", "www.example.test"}
	o.Canonical = "example.test"
	o.HTTPS = true
	o.Email = "admin@example.test"
	o.AgreeTOS = true
	first, e := m.Apply(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	css := ":root {--brand-1: #123456;}"
	o.Commit = strings.Repeat("b", 40)
	o.Archive = richArchive(t, true, css)
	second, e := m.Apply(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	other := Options{Name: "other", Branch: "staging", Commit: strings.Repeat("c", 40), Domains: []string{"other.test"}, Archive: archive(t, "other")}
	if _, e = m.Apply(ctx, other); e != nil {
		t.Fatal(e)
	}
	before, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	content, e := snips.ReadFile(before.Content["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	content.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "Owner edit after deploy"}
	content.Page("/").Title = "Owner title"
	if e = snips.WriteFile(before.Content["example.test"], content); e != nil {
		t.Fatal(e)
	}
	result, e := m.Rollback(ctx, o.Name, o.Branch, first, second)
	if e != nil {
		t.Fatal(e)
	}
	after, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		t.Fatal(e)
	}
	if after.Sites["example.test"] != after.Sites["www.example.test"] || filepath.Base(after.Sites["example.test"]) != result {
		t.Fatal("aliases split", after)
	}
	if after.Sites["other.test"] != before.Sites["other.test"] || after.Content["other.test"] != before.Content["other.test"] || !reflect.DeepEqual(after.TLS, before.TLS) || !reflect.DeepEqual(after.Redirects, before.Redirects) {
		t.Fatal("unrelated state or routing changed")
	}
	if after.Commits[o.Name] != strings.Repeat("a", 40) {
		t.Fatal("wrong commit")
	}
	state, e := m.Snapshot(o.Name, o.Branch)
	if e != nil {
		t.Fatal(e)
	}
	if state.Snips.Page("/").Title != "Owner title" || state.Snips.Page("/").Snips["a7Km2pQ9xB"].Content != "Owner edit after deploy" {
		t.Fatal("live edits lost", state)
	}
	p, e := site.Load(after.Sites["example.test"])
	if e != nil {
		t.Fatal(e)
	}
	if _, exists := p.Assets["/css/theme.css"]; exists {
		t.Fatal("rollback retained later release CSS")
	}
	records, e := m.History(o.Name, o.Branch)
	if e != nil || len(records) != 3 || records[0].Previous != second {
		t.Fatal(records, e)
	}
}

func TestQARollbackRejectsTargetsAndRecoversFailedProbe(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	first, e := m.Apply(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	second, e := m.Apply(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(m.ConfigPath)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct{ name, branch, target, expected string }{{o.Name, o.Branch, "../escape", second}, {o.Name, o.Branch, first, first}, {o.Name, "staging", first, second}, {"other", o.Branch, first, second}, {o.Name, o.Branch, second, second}} {
		if _, e = m.Rollback(ctx, tc.name, tc.branch, tc.target, tc.expected); e == nil {
			t.Fatal("accepted invalid rollback", tc)
		}
	}
	calls := 0
	m.Verify = func(context.Context, host.Config, []string) error {
		calls++
		if calls == 1 {
			return errors.New("probe failed")
		}
		return nil
	}
	if _, e = m.Rollback(ctx, o.Name, o.Branch, first, second); e == nil {
		t.Fatal("accepted failed probe")
	}
	after, e := os.ReadFile(m.ConfigPath)
	if e != nil || string(after) != string(before) {
		t.Fatal("did not restore exact active configuration", e)
	}
	records, e := m.History(o.Name, o.Branch)
	if e != nil || len(records) != 2 || records[0].ID != second {
		t.Fatal("failed rollback entered history", records, e)
	}
	if _, e = os.Stat(m.pending()); !os.IsNotExist(e) {
		t.Fatal("recovery journal remains", e)
	}
}

func TestQAFailedExtractionDoesNotEnterHistory(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	first, e := m.Apply(ctx, o)
	if e != nil {
		t.Fatal(e)
	}
	bad := filepath.Join(t.TempDir(), "bad.tar.gz")
	if e = os.WriteFile(bad, []byte("invalid archive"), 0600); e != nil {
		t.Fatal(e)
	}
	o.Archive = bad
	if _, e = m.Apply(ctx, o); e == nil {
		t.Fatal("accepted bad archive")
	}
	records, e := m.History(o.Name, o.Branch)
	if e != nil || len(records) != 1 || records[0].ID != first {
		t.Fatal(records, e)
	}
	receipts, e := os.ReadDir(filepath.Join(m.DataDir, "sites", o.Name, "history"))
	if e != nil || len(receipts) != 1 {
		t.Fatal(receipts, e)
	}
}
