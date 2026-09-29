package site

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminReloadScriptDevelopmentOnly(t *testing.T) {
	root := fixture(t)
	for _, name := range []string{"index", "code", "setup", "error"} {
		put(t, root, "admin/"+name+".html", `<p>Before</p>`)
	}
	p := mustLoad(t, root)
	s := NewServer(p)
	for _, mode := range []string{"login", "verify", "setup", "error"} {
		body, err := s.RenderAdmin(mode, nil)
		if err != nil || !strings.Contains(s.AdminReload(), reloadPath) {
			t.Fatalf("%s lacks current reload script: %s, %v", mode, body, err)
		}
		body, err = NewProductionServer(p, nil).RenderAdmin(mode, nil)
		if err != nil || bytes.Contains(body, []byte(reloadPath)) || NewProductionServer(p, nil).AdminReload() != "" {
			t.Fatalf("production %s has reload script: %s, %v", mode, body, err)
		}
	}
	before := s.revision
	put(t, root, "admin/index.html", `<p>After</p>`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	body, err := s.RenderAdmin("login", nil)
	if err != nil || s.revision == before || !bytes.Contains(reloadScript(s.revision), []byte(s.AdminReload())) || !bytes.Contains(body, []byte("After")) {
		t.Fatalf("admin reload did not publish updated content and revision: %s, %v", body, err)
	}
}

func TestAdminTemplatesArePrivateAndEditable(t *testing.T) {
	root := fixture(t)
	put(t, root, "admin/index.html", `<p>Original</p>`)
	put(t, root, "admin/dashboard.html", `<p>Private dashboard</p>`)
	put(t, root, "admin/secret.txt", `private asset`)
	s := NewServer(mustLoad(t, root))
	for _, path := range []string{"/admin", "/admin/index", "/admin/dashboard", "/admin/dashboard.html", "/admin/secret.txt"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("%s returned %d", path, w.Code)
		}
	}
	put(t, root, "admin/index.html", `<p>Edited directly</p>`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	got, err := s.RenderAdmin("login", nil)
	if err != nil || !strings.Contains(string(got), "Edited directly") {
		t.Fatal(string(got), err)
	}
	if _, err := s.RenderAdmin("verify", nil); err == nil {
		t.Fatal("missing template received a fallback")
	}
}

func TestAdminUsesCurrentLayoutWithoutCompilingContent(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<html><head><title>{page.title} | {app.site_name}</title></head><body>{render header}{slot}<footer>Site footer</footer></body></html>`)
	put(t, root, "_header.html", `<header>Site header</header>`)
	put(t, root, "admin/index.html", `<form>{admin.message}</form>`)
	s := NewServer(mustLoad(t, root))
	content := map[string]any{"message": "<script>{app.site_name}</script>"}
	got, err := s.RenderAdmin("login", content)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Admin | Test", "<header>Site header</header>", "&lt;script&gt;{app.site_name}&lt;/script&gt;", "<footer>Site footer</footer>"} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	put(t, root, "_header.html", `<header>Updated header</header>`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	got, err = s.RenderAdmin("login", content)
	if err != nil || !strings.Contains(string(got), "Updated header") {
		t.Fatal(string(got), err)
	}
	put(t, root, "main.layout.html", `{render missing}`)
	if s.Reload(root) == nil {
		t.Fatal("accepted invalid layout")
	}
	if _, err := s.RenderAdmin("login", content); err == nil {
		t.Fatal("served stale layout")
	}
}

func TestAdminWithoutLayout(t *testing.T) {
	root := fixture(t)
	put(t, root, "admin/index.html", `<form>Sign in</form>`)
	s := NewServer(mustLoad(t, root))
	got, err := s.RenderAdmin("login", nil)
	if err != nil || !strings.Contains(string(got), `<form>Sign in</form>`) {
		t.Fatal(string(got), err)
	}
}
