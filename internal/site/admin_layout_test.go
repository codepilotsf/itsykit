package site

import (
	"net/http/httptest"
	"strings"
	"testing"
)

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
	put(t, root, "_layout.html", `<html><head><title>{page.title} | {site.name}</title></head><body>{render header}{slot}<footer>Site footer</footer></body></html>`)
	put(t, root, "_header.html", `<header>Site header</header>`)
	put(t, root, "admin/index.html", `<form>{admin.message}</form>`)
	s := NewServer(mustLoad(t, root))
	content := map[string]any{"message": "<script>{site.name}</script>"}
	got, err := s.RenderAdmin("login", content)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Admin | Test", "<header>Site header</header>", "&lt;script&gt;{site.name}&lt;/script&gt;", "<footer>Site footer</footer>"} {
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
	put(t, root, "_layout.html", `{render missing}`)
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
