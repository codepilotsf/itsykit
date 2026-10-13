package editstore

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"flea/internal/editor"
	"flea/internal/site"
	"flea/internal/state"
	"flea/templates"
)

func TestLocalSavesDuringDevReloads(t *testing.T) {
	s, public := fixture(t, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchDone := make(chan struct{})
	go func() { defer close(watchDone); public.Watch(ctx, s.Root, nil) }()
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 15; i++ {
			doc, err := s.Read()
			if err == nil {
				err = s.Save(editor.Update{Version: doc.Version, ID: "abcdefghij", Content: strings.Repeat("x", i+1)})
			}
			if err != nil {
				done <- err
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("local save and reload deadlocked")
	}
	cancel()
	select {
	case <-watchDone:
	case <-time.After(3 * time.Second):
		t.Fatal("watcher did not stop")
	}
	if _, err := os.Stat(filepath.Join(s.Root, "flea", ".snips.lock")); !os.IsNotExist(err) {
		t.Fatal("local lock file created", err)
	}
}

func TestMetadataWithoutSnips(t *testing.T) {
	s, public := fixture(t, false)
	if err := os.WriteFile(filepath.Join(s.Root, "site", "about.html"), []byte(`<h1>About</h1><p>{page.title}</p><p>{page.description}</p>`), 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(doc.Routes, ","), "/about") {
		t.Fatal(doc.Routes)
	}
	if err := s.Save(editor.Update{Version: doc.Version, Route: "/about", Title: "About us", Description: "Our story"}); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	public.ServeHTTP(w, httptest.NewRequest("GET", "/about", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "About us") || !strings.Contains(w.Body.String(), "Our story") {
		t.Fatal(w.Code, w.Body.String())
	}
	doc, err = s.Read()
	if err != nil || doc.Store.Pages["/about"].Title != "About us" {
		t.Fatal(doc, err)
	}
}

func fixture(t *testing.T, remote bool) (*Store, *site.Server) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".flea.config.json": `{"app":{"site_name":"Editor test"},"admin":{"emails":["owner@example.test"]}}`,
		"site/index.html":   `<h1>{snip abcdefghij}</h1><p>{page.title}</p><p>{page.description}</p>`,
		"flea/snips.json":   `{"pages":{"/":{"snips":{"abcdefghij":{"label":"Heading","content":"Original"}}}},"shared":{"snips":{}}}`,
	}
	layout, err := templates.Files.ReadFile("site/admin.layout.html")
	if err != nil {
		t.Fatal(err)
	}
	files["site/admin.layout.html"] = string(layout)
	files["site/main.layout.html"] = `<html><head>{#if app.current_admin}<meta name="csrf-token" content="{app.current_admin.csrf_token}">{/if}</head><body>{#if app.current_admin}{render /admin/toolbar}{/if}{slot}</body></html>`
	toolbar, err := templates.Files.ReadFile("site/admin/_toolbar.html")
	if err != nil {
		t.Fatal(err)
	}
	files["site/admin/_toolbar.html"] = string(toolbar)
	for _, name := range []string{"index", "code", "setup", "error"} {
		data, err := templates.Files.ReadFile("site/admin/" + name + ".html")
		if err != nil {
			t.Fatal(err)
		}
		files["site/admin/"+name+".html"] = string(data)
	}
	// Keep reload tests offline: their toolbar references the canonical icon cache.
	// Missing icons would let LoadEditing perform network I/O under the content lock.
	icons, err := templates.Files.ReadDir("flea/icons")
	if err != nil {
		t.Fatal(err)
	}
	for _, icon := range icons {
		if icon.IsDir() {
			continue
		}
		name := "flea/icons/" + icon.Name()
		body, err := templates.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = string(body)
	}
	for name, body := range files {
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := site.IgnoreSessionFiles(root); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewProductionServer(p, nil)
	s := &Store{Root: root, ContentPath: filepath.Join(root, "flea", "snips.json"), Emails: p.Admin.Emails, Publish: public.Publish}
	if remote {
		dir := t.TempDir()
		s.ConfigPath = filepath.Join(dir, "server.json")
		s.ContentPath = filepath.Join(dir, "content.json")
		if err := os.WriteFile(s.ContentPath, []byte(files["flea/snips.json"]), 0644); err != nil {
			t.Fatal(err)
		}
		c := map[string]any{"sites": map[string]string{"example.test": root}, "content": map[string]string{"example.test": s.ContentPath}}
		data, _ := json.Marshal(c)
		if err := os.WriteFile(s.ConfigPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return s, public
}
func TestSavesPublishSanitizedContentAndRejectStaleForms(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, public := fixture(t, remote)
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			u := editor.Update{Version: doc.Version, ID: "abcdefghij", Content: `Hello <em>owner</em><script>alert(1)</script>{app.site_name}`}
			if err := s.Save(u); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(u); err == nil {
				t.Fatal("stale edit overwritten")
			}
			w := httptest.NewRecorder()
			public.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), "Hello <em>owner</em>") || strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "{app.site_name}") {
				t.Fatal(w.Body.String())
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Save(editor.Update{Version: doc.Version, Route: "/", Title: `<script>title</script>`, Description: "Details"}); err != nil {
				t.Fatal(err)
			}
			w = httptest.NewRecorder()
			public.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			if strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "&lt;script&gt;title") {
				t.Fatal(w.Body.String())
			}
		})
	}
}
func TestDeploymentLockRecoveryAndReleaseChecks(t *testing.T) {
	s, _ := fixture(t, true)
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	u := editor.Update{Version: doc.Version, ID: "abcdefghij", Content: "Edited"}
	unlock, err := state.Lock(filepath.Join(filepath.Dir(s.ConfigPath), "deploy.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(u); err == nil {
		t.Fatal("saved during deployment")
	}
	unlock()
	pending := filepath.Join(filepath.Dir(s.ConfigPath), "deploy.pending.json")
	if err := os.WriteFile(pending, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(u); err == nil {
		t.Fatal("saved during recovery")
	}
	if err := os.Remove(pending); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ConfigPath, []byte(`{"sites":{},"content":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(u); err == nil {
		t.Fatal("saved into obsolete release")
	}
}
func TestRejectUnknownTargetsAndChangedAccess(t *testing.T) {
	s, _ := fixture(t, false)
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []editor.Update{{Version: doc.Version, ID: "../../private"}, {Version: doc.Version, Route: "/missing"}, {Version: doc.Version, Route: "/", Title: strings.Repeat("x", 201)}} {
		if err := s.Save(u); err == nil {
			t.Fatal("invalid edit accepted", u)
		}
	}
	if err := os.WriteFile(filepath.Join(s.Root, ".flea.config.json"), []byte(`{"app":{"site_name":"Changed"},"admin":{"emails":[]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(); err == nil {
		t.Fatal("continued after access changed")
	}
}

func TestSnipSavePreservesSourceCSSAndRejectsChangedStyles(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, public := fixture(t, remote)
			cssPath := filepath.Join(s.Root, "site", "css", "theme.css")
			if err := os.MkdirAll(filepath.Dir(cssPath), 0755); err != nil {
				t.Fatal(err)
			}
			css := ":root { --brand-1: #123456; }"
			if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Save(editor.Update{Version: doc.Version, ID: "abcdefghij", Content: "Updated text"}); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			public.ServeHTTP(w, httptest.NewRequest("GET", "/css/theme.css", nil))
			if w.Code != 200 || w.Body.String() != css {
				t.Fatal("content save changed source styling", w.Code, w.Body.String())
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(cssPath, []byte(":root { --brand-1: #654321; }"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(editor.Update{Version: doc.Version, ID: "abcdefghij", Content: "Stale draft"}); err == nil {
				t.Fatal("accepted content draft after source CSS changed")
			}
		})
	}
}
