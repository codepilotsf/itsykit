package site

import (
	"net/http/httptest"
	"testing"
)

func TestRootMetadataIsolation(t *testing.T) {
	root := fixture(t)
	before := fingerprint(root)
	for _, name := range []string{"README.md", "AGENTS.md", "ARCHITECTURE.md", ".git/config", "node_modules/package/index.html", "_/deploy.json", ".env"} {
		put(t, root, name, "PRIVATE {broken directive}")
	}
	if fingerprint(root) != before {
		t.Fatal("metadata triggered reload")
	}
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(p)
	for _, path := range []string{"/README.md", "/AGENTS.md", "/ARCHITECTURE.md", "/.git/config", "/node_modules/package", "/_/deploy.json", "/.env", "/_layout.html", "/index.html"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	if _, ok := p.Pages["/"]; !ok {
		t.Fatal("missing root page")
	}
	if _, ok := p.Assets["/css/main.css"]; !ok {
		t.Fatal("missing root asset")
	}
}
