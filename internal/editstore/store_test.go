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
	if _, err := os.Stat(filepath.Join(s.Root, "_", ".snips.lock")); !os.IsNotExist(err) {
		t.Fatal("local lock file created", err)
	}
}

func fixture(t *testing.T, remote bool) (*Store, *site.Server) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"_/flea.config": `{"site":{"name":"Editor test"},"admin":{"emails":["owner@example.test"]}}`,
		"index.html":    `<h1>{snip abcdefghij}</h1><p>{page.title}</p><p>{page.description}</p>`,
		"_/snips.json":  `{"pages":{"/":{"snips":{"abcdefghij":{"label":"Heading","content":"Original"}}}},"shared":{"snips":{}}}`,
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
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewProductionServer(p, nil)
	s := &Store{Root: root, ContentPath: filepath.Join(root, "_", "snips.json"), Emails: p.Admin.Emails, Publish: public.Publish}
	if remote {
		dir := t.TempDir()
		s.ConfigPath = filepath.Join(dir, "server.json")
		s.ContentPath = filepath.Join(dir, "content.json")
		if err := os.WriteFile(s.ContentPath, []byte(files["_/snips.json"]), 0644); err != nil {
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
			u := editor.Update{Version: doc.Version, ID: "abcdefghij", Content: `Hello <em>owner</em><script>alert(1)</script>{site.name}`}
			if err := s.Save(u); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(u); err == nil {
				t.Fatal("stale edit overwritten")
			}
			w := httptest.NewRecorder()
			public.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), "Hello <em>owner</em>") || strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "{site.name}") {
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
	if err := os.WriteFile(filepath.Join(s.Root, "_", "flea.config"), []byte(`{"site":{"name":"Changed"},"admin":{"emails":[]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(); err == nil {
		t.Fatal("continued after access changed")
	}
}
