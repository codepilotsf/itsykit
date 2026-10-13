package admin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"flea/internal/host"
)

var errTestFailure = errors.New("simulated restart failure")

func sitesFixture(t *testing.T) Manager {
	t.Helper()
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if err := os.Mkdir(data, 0755); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, "server.json")
	if err := os.WriteFile(config, []byte(`{"listen":":80","sites":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	return Manager{ConfigPath: config, DataDir: data, Restart: func(context.Context) error { return nil }, Verify: func(context.Context, host.Config, []string) error { return nil }}
}

func applyHTTPSite(t *testing.T, m Manager, name string, domains []string, canonical string) string {
	t.Helper()
	o := Options{Archive: archive(t, name), Name: name, Domains: domains, HTTPS: true, Email: "dev@example.test", AgreeTOS: true, Canonical: canonical, Branch: "main", Commit: strings.Repeat("a", 40)}
	id, err := m.Apply(context.Background(), o)
	if err != nil {
		t.Fatalf("apply %s: %v", name, err)
	}
	return id
}

func TestSitesDeterministicInventory(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test", "www.a.test"}, "a.test")
	beta := applyHTTPSite(t, m, "beta", []string{"b.test"}, "")

	before, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	first, err := m.Sites()
	if err != nil {
		t.Fatal(err)
	}
	second, err := m.Sites()
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("listing mutated the server config")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("inventory is not deterministic:\n%+v\n%+v", first, second)
	}
	if len(first) != 2 || first[0].Name != "alpha" || first[1].Name != "beta" {
		t.Fatalf("unexpected ordering: %+v", first)
	}
	alpha := first[0]
	if !reflect.DeepEqual(alpha.Domains, []string{"a.test", "www.a.test"}) {
		t.Fatalf("domains: %+v", alpha.Domains)
	}
	if alpha.Branch != "main" || alpha.Release == "" || !alpha.Managed {
		t.Fatalf("alpha summary: %+v", alpha)
	}
	if alpha.Retained != filepath.Join(m.DataDir, "sites", "alpha") {
		t.Fatalf("retained path: %s", alpha.Retained)
	}
	if len(alpha.Redirects) != 1 || alpha.Redirects[0] != "www.a.test -> a.test" {
		t.Fatalf("redirects: %+v", alpha.Redirects)
	}
	if len(alpha.HTTPS) != 2 {
		t.Fatalf("https: %+v", alpha.HTTPS)
	}
	if first[1].Release != beta || first[1].Release == "" {
		t.Fatalf("beta release: %+v", first[1])
	}
}

func TestRemoveSiteRejectsStaleStateAndPreservesOthers(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test", "www.a.test"}, "a.test")
	applyHTTPSite(t, m, "beta", []string{"b.test"}, "")
	info, err := m.SiteInfo("alpha")
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveSite(context.Background(), "alpha", "stale-digest"); err == nil {
		t.Fatal("stale reviewed state was accepted")
	}
	after, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("stale removal mutated the config")
	}
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Sites["a.test"] != "" || c.Sites["www.a.test"] != "" {
		t.Fatalf("target routing remains: %+v", c.Sites)
	}
	if c.Sites["b.test"] == "" || c.SiteBranches["beta"] == "" {
		t.Fatalf("unrelated site was changed: %+v", c)
	}
	if c.SiteBranches["alpha"] != "" || c.Commits["alpha"] != "" {
		t.Fatal("target bookkeeping remains")
	}
	if len(c.Redirects) != 0 {
		t.Fatalf("target redirect remains: %+v", c.Redirects)
	}
	if c.Content["a.test"] != "" {
		t.Fatal("target content mapping remains")
	}
	for _, d := range c.TLS.Domains {
		if d == "a.test" || d == "www.a.test" {
			t.Fatalf("target TLS entry remains: %+v", c.TLS.Domains)
		}
	}
	if _, err := os.Stat(filepath.Join(m.DataDir, "sites", "alpha", "releases", info.Release)); err != nil {
		t.Fatalf("release files were deleted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(m.DataDir, "sites", "alpha")); err != nil {
		t.Fatalf("site directory was deleted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(m.DataDir, "sites", "alpha", "history", info.Release+".json")); err != nil {
		t.Fatalf("history was deleted: %v", err)
	}
	sites, err := m.Sites()
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != 1 || sites[0].Name != "beta" {
		t.Fatalf("listing after removal: %+v", sites)
	}
}

// Removal deletes the site's commit bookkeeping, so a changed commit alone must
// invalidate the reviewed digest and leave the config untouched.
func TestRemoveSiteRejectsChangedCommitWithoutMutation(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test"}, "")
	info, err := m.SiteInfo("alpha")
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Commits["alpha"] = strings.Repeat("b", 40)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(m.ConfigPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	err = m.RemoveSite(context.Background(), "alpha", info.State)
	if err == nil || !strings.Contains(err.Error(), "changed since you reviewed") {
		t.Fatalf("commit-only change was accepted: %v", err)
	}
	after, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("stale removal mutated the config")
	}
	if _, err := m.SiteInfo("alpha"); err != nil {
		t.Fatalf("site registration was lost: %v", err)
	}
}

// Removing the site that owns the last HTTPS hostname must clear the TLS object,
// because host validation rejects a TLS config with no domains.
func TestRemoveSiteClearsTLSWhenLastHTTPSDomainRemoved(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test"}, "")
	if _, err := m.Apply(context.Background(), Options{Archive: archive(t, "beta"), Name: "beta", Domains: []string{"b.test"}, Branch: "main", Commit: strings.Repeat("b", 40)}); err != nil {
		t.Fatal(err)
	}
	info, err := m.SiteInfo("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.TLS != nil {
		t.Fatalf("last HTTPS hostname removal left TLS configured: %+v", c.TLS)
	}
	if c.Sites["b.test"] == "" || c.SiteBranches["beta"] == "" {
		t.Fatalf("unrelated site was changed: %+v", c)
	}
}

func TestRemoveLastSiteReadinessAndRetention(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "only", []string{"only.test"}, "")
	var verified []string
	m.Verify = func(_ context.Context, _ host.Config, domains []string) error {
		verified = append([]string{}, domains...)
		return nil
	}
	info, err := m.SiteInfo("only")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveSite(context.Background(), "only", info.State); err != nil {
		t.Fatal(err)
	}
	if len(verified) != 0 {
		t.Fatalf("last-site removal should probe generic readiness, got %+v", verified)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Sites) != 0 || c.TLS != nil {
		t.Fatalf("last site was not fully unregistered: %+v", c)
	}
	if _, err := os.Stat(filepath.Join(m.DataDir, "sites", "only", "releases", info.Release)); err != nil {
		t.Fatalf("last-site files were deleted: %v", err)
	}
	sites, err := m.Sites()
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) != 0 {
		t.Fatalf("expected no sites, got %+v", sites)
	}
}

func TestRemoveSiteBlocksPendingAndBusy(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test"}, "")
	info, err := m.SiteInfo("alpha")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.pending(), []byte(`{"listen":":80","sites":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err == nil {
		t.Fatal("removal ignored a pending journal")
	}
	os.Remove(m.pending())
	unlock, err := m.lock()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err == nil {
		t.Fatal("removal ignored the deployment lock")
	}
	unlock()
}

func TestRemoveSiteRestartFailureRollsBack(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test"}, "")
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := m.SiteInfo("alpha")
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
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err == nil {
		t.Fatal("restart failure reported as success")
	}
	got, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(old) {
		t.Fatal("failed removal did not restore the previous config")
	}
	if _, err := os.Stat(m.pending()); !os.IsNotExist(err) {
		t.Fatalf("successful rollback should clear the journal: %v", err)
	}

	// A recovery that cannot restart retains the journal for -recover.
	m.Restart = func(context.Context) error { return errTestFailure }
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err == nil {
		t.Fatal("unrecoverable restart reported as success")
	}
	if _, err := os.Stat(m.pending()); err != nil {
		t.Fatalf("failed recovery lost the journal: %v", err)
	}
	m.Restart = func(context.Context) error { return nil }
	if err := m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(m.pending()); !os.IsNotExist(err) {
		t.Fatalf("recovery left the journal: %v", err)
	}
}

func TestRemoveSiteProbeFailureRollsBack(t *testing.T) {
	m := sitesFixture(t)
	applyHTTPSite(t, m, "alpha", []string{"a.test"}, "")
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := m.SiteInfo("alpha")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	m.Verify = func(context.Context, host.Config, []string) error {
		calls++
		if calls == 1 {
			return errTestFailure
		}
		return nil
	}
	if err := m.RemoveSite(context.Background(), "alpha", info.State); err == nil {
		t.Fatal("probe failure reported as success")
	}
	got, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(old) {
		t.Fatal("failed removal did not restore the previous config")
	}
	if _, err := os.Stat(m.pending()); !os.IsNotExist(err) {
		t.Fatalf("successful rollback should clear the journal: %v", err)
	}
}

func TestRemoveSiteUnmanagedFailsClosed(t *testing.T) {
	m := sitesFixture(t)
	root := filepath.Join(t.TempDir(), "custom", "releases", "1")
	if err := os.MkdirAll(filepath.Join(root, "site"), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, ".flea.config.json"), []byte(`{"app":{"site_name":"Custom"}}`), 0644)
	os.WriteFile(filepath.Join(root, "site", "index.html"), []byte("hi"), 0644)
	c := host.Config{Listen: ":80", Sites: map[string]string{"x.test": root}}
	data, _ := json.Marshal(c)
	if err := os.WriteFile(m.ConfigPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := m.SiteInfo("custom"); err == nil || !strings.Contains(err.Error(), "not managed") {
		t.Fatalf("unmanaged site accepted by review: %v", err)
	}
	if err := m.RemoveSite(context.Background(), "custom", "anything"); err == nil || !strings.Contains(err.Error(), "not managed") {
		t.Fatalf("unmanaged site accepted by removal: %v", err)
	}
}

func TestRemoveSiteEntriesFailsClosedOnAmbiguousRedirect(t *testing.T) {
	c := host.Config{
		Sites:     map[string]string{"a.test": "/data/sites/a/releases/1"},
		Redirects: map[string]string{"www.other.test": "a.test"},
	}
	g := siteGroup{name: "a", root: "/data/sites/a/releases/1", domains: []string{"a.test"}, managed: true}
	if err := removeSiteEntries(&c, g); err == nil || !strings.Contains(err.Error(), "crosses site ownership") {
		t.Fatalf("ambiguous redirect: %v", err)
	}
}

func TestRemoveSiteEntriesRemovesOnlyTargetOwnedEntries(t *testing.T) {
	c := host.Config{
		Sites:        map[string]string{"a.test": "/data/sites/a/releases/1", "www.a.test": "/data/sites/a/releases/1", "b.test": "/data/sites/b/releases/1"},
		Content:      map[string]string{"a.test": "/data/sites/a/content/1.json", "b.test": "/data/sites/b/content/1.json"},
		Downloads:    map[string]string{"a.test": "/store/a", "b.test": "/store/b"},
		Redirects:    map[string]string{"www.a.test": "a.test", "www.b.test": "b.test"},
		SiteBranches: map[string]string{"a": "main", "b": "main"},
		Commits:      map[string]string{"a": strings.Repeat("a", 40), "b": strings.Repeat("b", 40)},
		TLS:          &host.TLSConfig{Listen: ":443", Email: "dev@example.test", AgreeTOS: true, CacheDir: "/var/lib/flea/certs", Domains: []string{"a.test", "www.a.test", "b.test"}},
	}
	g := siteGroup{name: "a", root: "/data/sites/a/releases/1", domains: []string{"a.test", "www.a.test"}, managed: true}
	if err := removeSiteEntries(&c, g); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Sites["a.test"]; ok {
		t.Fatal("target routing remains")
	}
	if c.Sites["b.test"] == "" {
		t.Fatal("unrelated routing lost")
	}
	if _, ok := c.Content["a.test"]; ok {
		t.Fatal("target content remains")
	}
	if c.Content["b.test"] == "" {
		t.Fatal("unrelated content lost")
	}
	if _, ok := c.Downloads["a.test"]; ok {
		t.Fatal("target download mapping remains")
	}
	if c.Downloads["b.test"] == "" {
		t.Fatal("unrelated download mapping lost")
	}
	if _, ok := c.Redirects["www.a.test"]; ok {
		t.Fatal("target redirect remains")
	}
	if c.Redirects["www.b.test"] != "b.test" {
		t.Fatal("unrelated redirect lost")
	}
	if _, ok := c.SiteBranches["a"]; ok {
		t.Fatal("target branch bookkeeping remains")
	}
	if c.SiteBranches["b"] != "main" {
		t.Fatal("unrelated branch bookkeeping lost")
	}
	if c.TLS == nil || len(c.TLS.Domains) != 1 || c.TLS.Domains[0] != "b.test" {
		t.Fatalf("TLS allowlist: %+v", c.TLS)
	}
	if c.TLS.Email != "dev@example.test" || !c.TLS.AgreeTOS {
		t.Fatal("shared TLS settings were lost")
	}
}
