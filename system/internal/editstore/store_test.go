package editstore

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"itsy/internal/editor"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/state"
	"itsy/templates"
)

func fixture(t *testing.T, remote bool) (*Store, *site.Server) {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"itsy.config.json": `{"app":{"site_name":"Editor test"},"admin":{"emails":["owner@example.test"]},"apps":{"blog":{"mount":"/blog"}}}`,
		"site/index.html":  `<h1>Original</h1><p>{page.title}</p><p>{page.description}</p>`,
		"itsy/snips.json":  `{}`,
	}
	files["site/blog/_index.html"] = `{#each blog.data.posts as post}{post.title}{/each}`
	files["site/blog/_post.html"] = `{blog.data.post.body}`
	layout, err := templates.Files.ReadFile("site/admin.layout.html")
	if err != nil {
		t.Fatal(err)
	}
	files["site/admin.layout.html"] = string(layout)
	files["site/main.layout.html"] = `<html><head></head><body>{slot}</body></html>`
	for _, name := range []string{"index", "code", "setup", "error", "blog", "blog-edit"} {
		data, err := templates.Files.ReadFile("site/admin/" + name + ".html")
		if err != nil {
			t.Fatal(err)
		}
		files["site/admin/"+name+".html"] = string(data)
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
	s := &Store{Root: root, ContentPath: filepath.Join(root, "itsy", "snips.json"), Emails: p.Admin.Emails, Publish: public.Publish}
	if remote {
		dir := t.TempDir()
		s.ConfigPath = filepath.Join(dir, "server.json")
		s.ContentPath = filepath.Join(dir, "content.json")
		if err := os.WriteFile(s.ContentPath, []byte(files["itsy/snips.json"]), 0644); err != nil {
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
func TestDeploymentLockRecoveryAndReleaseChecks(t *testing.T) {
	s, _ := fixture(t, true)
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	u := editor.Update{Version: doc.Version, Blog: &editor.BlogEdit{}}
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
	if _, err := s.Read(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "itsy.config.json"), []byte(`{"app":{"site_name":"Changed"},"admin":{"emails":[]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(); err == nil {
		t.Fatal("continued after access changed")
	}
}

func TestBlogSavePreservesSourceCSSAndRejectsChangedStyles(t *testing.T) {
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
			if err := s.Save(blogUpdate(doc.Version)); err != nil {
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
			if err := s.Save(blogUpdate(doc.Version)); err == nil {
				t.Fatal("accepted content draft after source CSS changed")
			}
		})
	}
}

func blogUpdate(version string) editor.Update {
	return editor.Update{Version: version, Blog: &editor.BlogEdit{Action: "save", Post: snips.BlogPost{ID: "Abcdef1234", Title: "Post", Slug: "post", Status: "draft", Body: "<p>Text</p>"}}}
}
