package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppSettings(t *testing.T) {
	for folder, want := range map[string]string{"my-project": "My Project", "ACME_studio": "Acme Studio", "café.design": "Café Design", "bookworm20": "Bookworm20"} {
		if got := DefaultSiteName(folder); got != want {
			t.Fatalf("%s: %s", folder, got)
		}
	}
	root := fixture(t)
	put(t, root, ".flea.config.json", `{"app":{"phone":"<123>","number":9007199254740993},"admin":{"emails":[]},"custom":{"keep":true}}`)
	put(t, root, "index.html", `{app.site_name} {app.phone} {site.name} {#each snip.team.members as member}{app.phone}{/each}`)
	before, _ := os.ReadFile(filepath.Join(root, ".flea.config.json"))
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.SiteNameConfigured || p.App["site_name"] != DefaultSiteName(root) {
		t.Fatal("default", p.App)
	}
	after, _ := os.ReadFile(filepath.Join(root, ".flea.config.json"))
	if string(before) != string(after) {
		t.Fatal("load wrote config")
	}
	if err := SetSiteName(root, "My <Studio>"); err != nil {
		t.Fatal(err)
	}
	p, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !p.SiteNameConfigured {
		t.Fatal("name not saved")
	}
	body, err := p.render(p.Pages["/"], "/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "My &lt;Studio&gt; &lt;123&gt;") {
		t.Fatal(string(body))
	}
	saved, _ := os.ReadFile(filepath.Join(root, ".flea.config.json"))
	if !strings.Contains(string(saved), "9007199254740993") || !strings.Contains(string(saved), `"keep": true`) {
		t.Fatal(string(saved))
	}
	for _, bad := range []string{`{"site":{"name":"Old"}}`, `{"app":null}`, `{"app":{"site_name":false}}`, `{"app":{"site_name":""}}`} {
		put(t, root, ".flea.config.json", bad)
		if _, err := Load(root); err == nil {
			t.Fatal("accepted", bad)
		}
		if err := SetSiteName(root, "New"); err == nil {
			t.Fatal("overwrote invalid config", bad)
		}
	}
}
