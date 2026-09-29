package site

import (
	"bytes"
	"context"
	"flea/internal/snips"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOwnerRenderingPreservesLayoutAndIsRequestScoped(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<!doctype html><html><body><header class="custom">My header</header><main>{slot}</main><footer>My footer</footer></body></html>`)
	put(t, root, "index.html", `<article snip class="custom-content"><p>Hello <b>world</b></p></article>`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	server := NewProductionServer(p, nil)
	render := func(owner bool, method, path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, path, nil)
		if owner {
			server.ServeOwner(w, r, func(body []byte, _ *snips.Store, _ string) ([]byte, bool) {
				return append(body, []byte("OWNER")...), true
			})
		} else {
			server.ServeHTTP(w, r)
		}
		return w
	}
	before := render(false, "GET", "/")
	owner := render(true, "GET", "/")
	after := render(false, "GET", "/")
	if !bytes.Equal(before.Body.Bytes(), after.Body.Bytes()) || strings.Contains(before.Body.String(), "snip-start:") {
		t.Fatal("owner state leaked to visitor")
	}
	for _, want := range []string{`<header class="custom">My header</header>`, `<article class="custom-content">`, `<p>Hello <b>world</b></p>`, `<footer>My footer</footer>`, "<!--snip-start:", "<!--snip-end:", "OWNER"} {
		if !strings.Contains(owner.Body.String(), want) {
			t.Fatalf("missing %q: %s", want, owner.Body.String())
		}
	}
	for _, path := range []string{"/css/main.css", "/missing", "/flea/snips.json"} {
		if strings.Contains(render(true, "GET", path).Body.String(), "OWNER") {
			t.Fatal("decorated asset/error/private path", path)
		}
	}
	if w := render(true, "HEAD", "/"); w.Code != 200 || w.Body.Len() != 0 {
		t.Fatal("HEAD", w.Code, w.Body.String())
	}
}
