package site

import (
	"itsy/internal/snips"
	"itsy/internal/viewer"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCurrentAdminIsRequestScoped(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{#if app.current_admin}<meta name="csrf-token" content="{app.current_admin.csrf_token}"><p>{app.current_admin.email}</p>{:else}<p>Visitor</p>{/if}`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	for _, token := range []string{"first-token", "second-token"} {
		r := httptest.NewRequest("GET", "https://example.test/", nil)
		r = r.WithContext(viewer.With(r.Context(), viewer.Admin{Email: "<owner@example.test>", CSRF: token}))
		w := httptest.NewRecorder()
		s.ServeOwner(w, r, func(body []byte, _ *snips.Store, _ string) ([]byte, bool) { return body, true })
		if !strings.Contains(w.Body.String(), token) || !strings.Contains(w.Body.String(), "&lt;owner@example.test&gt;") {
			t.Fatal(w.Body.String())
		}
	}
	body := request(s, "GET", "/").Body.String()
	if strings.Contains(body, "csrf-token") || !strings.Contains(body, "Visitor") {
		t.Fatal("identity leaked", body)
	}
	if _, ok := p.App["current_admin"]; ok {
		t.Fatal("shared snapshot mutated")
	}
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test","current_admin":{"email":"forged"}}}`)
	if _, err := Load(root); err == nil {
		t.Fatal("configured identity accepted")
	}
}
