package admin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"flea/internal/backup"
	"flea/internal/host"
	"flea/internal/syncstate"
)

func setTestIndexing(t *testing.T, m Manager, name, domain, mode string) {
	t.Helper()
	info, err := m.SiteInfo(name)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetIndexing(context.Background(), domain, mode, info.State); err != nil {
		t.Fatal(err)
	}
}

func assertIndexing(t *testing.T, m Manager, want map[string]string) host.Config {
	t.Helper()
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Indexing, want) {
		t.Fatalf("indexing=%v, want %v", c.Indexing, want)
	}
	return c
}

func TestIndexingPersistsAcrossDeploySyncRollbackRestore(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	o.Domains = []string{"example.test", "staging.test"}
	first, err := m.Apply(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	// Two aliases of the same site are independently configured.
	setTestIndexing(t, m, o.Name, "staging.test", "noindex")
	setTestIndexing(t, m, o.Name, "example.test", "allow")
	want := map[string]string{"staging.test": "noindex", "example.test": "allow"}
	before := assertIndexing(t, m, want)
	var buf bytes.Buffer
	if err = m.Backup(o.Name, o.Branch, &buf); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if _, err = backup.Extract(bytes.NewReader(buf.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dest, ".flea.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "indexing") {
		t.Fatal("host policy entered portable backup")
	}
	backupFile := filepath.Join(t.TempDir(), "site.flea-backup")
	if err = os.WriteFile(backupFile, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	// Deploy a different site, then redeploy the original with no policy option.
	other := Options{Name: "other", Branch: "main", Commit: strings.Repeat("b", 40), Domains: []string{"other.test"}, Archive: archive(t, "other")}
	if _, err = m.Apply(ctx, other); err != nil {
		t.Fatal(err)
	}
	setTestIndexing(t, m, "other", "other.test", "noindex")
	want["other.test"] = "noindex"
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	assertIndexing(t, m, want)
	// Explicit content sync uses the same activation path and keeps policies.
	live, err := m.Snapshot(o.Name, o.Branch)
	if err != nil {
		t.Fatal(err)
	}
	o.Sync = &syncstate.Plan{Expected: live.Version(), Snips: live.Snips}
	active, err := m.Apply(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	assertIndexing(t, m, want)
	if _, err = m.Rollback(ctx, o.Name, o.Branch, first, active); err != nil {
		t.Fatal(err)
	}
	assertIndexing(t, m, want)
	// Policy changed after backup: restore must retain the current host policy.
	setTestIndexing(t, m, o.Name, "example.test", "noindex")
	want["example.test"] = "noindex"
	live, err = m.Snapshot(o.Name, o.Branch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Restore(ctx, o.Name, o.Branch, backupFile, live.Version()); err != nil {
		t.Fatal(err)
	}
	after := assertIndexing(t, m, want)
	if after.Sites["staging.test"] != after.Sites["example.test"] {
		t.Fatal("aliases split")
	}
	if after.Sites["example.test"] == before.Sites["example.test"] {
		t.Fatal("test never changed releases")
	}
	// Explicit allow clears only the framework policy on exactly one hostname.
	setTestIndexing(t, m, o.Name, "example.test", "allow")
	want["example.test"] = "allow"
	assertIndexing(t, m, want)
	info, err := m.SiteInfo(o.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.RemoveSite(ctx, o.Name, info.State); err != nil {
		t.Fatal(err)
	}
	delete(want, "example.test")
	delete(want, "staging.test")
	assertIndexing(t, m, want)
	o.Sync = nil
	if _, err = m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	assertIndexing(t, m, want) // Re-registering does not restore removed policies.
}

func TestIndexingReviewLockAndRecovery(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	if _, err := m.Apply(ctx, o); err != nil {
		t.Fatal(err)
	}
	original, err := m.SiteInfo(o.Name)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ domain, mode, expected string }{
		{"unknown.test", "noindex", original.State}, {"EXAMPLE.TEST", "noindex", original.State}, {"example.test", "index", original.State}, {"example.test", "noindex", ""}, {"example.test", "noindex", "stale"},
	} {
		if err = m.SetIndexing(ctx, tc.domain, tc.mode, tc.expected); err == nil {
			t.Fatal("accepted invalid change", tc)
		}
	}
	unlock, err := m.lock()
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetIndexing(ctx, "example.test", "noindex", original.State); err == nil {
		t.Fatal("ignored busy lock")
	}
	unlock()
	setTestIndexing(t, m, o.Name, "example.test", "noindex")
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(c.Revision, "indexing-") {
		t.Fatal("policy activation did not change readiness revision")
	}
	if err = m.RemoveSite(ctx, o.Name, original.State); err == nil {
		t.Fatal("stale removal ignored policy change")
	}
	if err = m.SetIndexing(ctx, "example.test", "allow", original.State); err == nil {
		t.Fatal("stale policy review accepted")
	}
	// A failed restart rolls back exact config bytes and retains no new policy.
	current, err := m.SiteInfo(o.Name)
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	m.Restart = func(context.Context) error {
		calls++
		if calls == 1 {
			return errTestFailure
		}
		return nil
	}
	if err = m.SetIndexing(ctx, "example.test", "allow", current.State); err == nil {
		t.Fatal("restart failure accepted")
	}
	got, err := os.ReadFile(m.ConfigPath)
	if err != nil || !bytes.Equal(got, old) {
		t.Fatal("failed policy change did not restore config", err)
	}
	// Failed recovery leaves journal; -recover restores the complete prior host policy.
	m.Restart = func(context.Context) error { return errTestFailure }
	if err = m.SetIndexing(ctx, "example.test", "allow", current.State); err == nil {
		t.Fatal("failed recovery accepted")
	}
	if _, err = os.Stat(m.pending()); err != nil {
		t.Fatal("journal lost", err)
	}
	if err = m.SetIndexing(ctx, "example.test", "allow", current.State); err == nil || !strings.Contains(err.Error(), "unfinished") {
		t.Fatal("pending recovery not blocked", err)
	}
	m.Restart = func(context.Context) error { return nil }
	if err = m.Recover(ctx); err != nil {
		t.Fatal(err)
	}
	assertIndexing(t, m, map[string]string{"example.test": "noindex"})
	// Removal also restores policy if its readiness probe fails.
	calls = 0
	m.Verify = func(context.Context, host.Config, []string) error {
		calls++
		if calls == 1 {
			return errTestFailure
		}
		return nil
	}
	if err = m.RemoveSite(ctx, o.Name, current.State); err == nil {
		t.Fatal("failed removal accepted")
	}
	assertIndexing(t, m, map[string]string{"example.test": "noindex"})
}
