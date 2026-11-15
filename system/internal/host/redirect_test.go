package host

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWWWRedirect(t *testing.T) {
	file := testConfig(t, ":80")
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Sites["www.example.test"] = c.Sites["example.test"]
	c.Redirects = map[string]string{"www.example.test": "example.test"}
	data, _ := json.Marshal(c)
	os.WriteFile(file, data, 0644)
	_, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/hello%2Fworld?q=1", nil)
	r.Host = "www.example.test"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 308 || w.Header().Get("Location") != "http://example.test/hello%2Fworld?q=1" {
		t.Fatal(w.Code, w.Header())
	}
	c.TLS = &TLSConfig{Domains: []string{"example.test", "www.example.test"}}
	w = httptest.NewRecorder()
	redirectHTTPS(c, handler).ServeHTTP(w, r)
	if w.Code != 308 || w.Header().Get("Location") != "https://example.test/hello%2Fworld?q=1" {
		t.Fatal(w.Code, w.Header())
	}
	r = httptest.NewRequest("GET", HealthPath, nil)
	r.Host = "www.example.test"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	c.TLS = nil
	c.Redirects["www.example.test"] = "evil.test"
	data, _ = json.Marshal(c)
	os.WriteFile(file, data, 0644)
	if _, _, err := Load(file, nil); err == nil {
		t.Fatal("accepted redirect outside site")
	}
}

func TestDomainRedirectAcrossSites(t *testing.T) {
	file := testConfig(t, ":80")
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Distinct site roots matter: fleacms.com and itsykit.dev are separate
	// managed deployments, whereas itsyweb.dev is an alias of the new site.
	oldRoot := filepath.Join(t.TempDir(), "old")
	if err := os.MkdirAll(filepath.Join(oldRoot, "site"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldRoot, "itsy.config.json"), []byte(`{"app":{"site_name":"Old site"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldRoot, "site", "index.html"), []byte("old site"), 0644); err != nil {
		t.Fatal(err)
	}
	c.Sites["old.test"] = oldRoot
	c.Sites["www.old.test"] = oldRoot
	c.Sites["alias.test"] = c.Sites["example.test"]
	c.Redirects = map[string]string{"old.test": "example.test", "www.old.test": "example.test", "alias.test": "example.test"}
	c.TLS = &TLSConfig{Listen: ":443", Domains: []string{"old.test", "www.old.test", "alias.test", "example.test"}, Email: "dev@example.test", AgreeTOS: true, CacheDir: t.TempDir()}
	data, _ := json.Marshal(c)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	_, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"old.test", "www.old.test", "alias.test"} {
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
			for _, secure := range []bool{false, true} {
				r := httptest.NewRequest(method, "/docs/a%2Fb?query=one%20two&x=1", nil)
				r.Host = name
				if secure {
					r.TLS = &tls.ConnectionState{ServerName: name}
				}
				w := httptest.NewRecorder()
				if secure {
					tlsHandler(c, handler).ServeHTTP(w, r)
				} else {
					redirectHTTPS(c, handler).ServeHTTP(w, r)
				}
				if w.Code != http.StatusPermanentRedirect || w.Header().Get("Location") != "https://example.test/docs/a%2Fb?query=one%20two&x=1" {
					t.Fatalf("%s %s secure=%v: %d %s", method, name, secure, w.Code, w.Header().Get("Location"))
				}
			}
		}
	}
	// Internal health probes still bypass public redirection.
	r := httptest.NewRequest(http.MethodGet, HealthPath, nil)
	r.Host = "old.test"
	w := httptest.NewRecorder()
	redirectHTTPS(c, handler).ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Header().Get("Location") != "" {
		t.Fatal(w.Code, w.Header())
	}
	// An unregistered Host never acquires a redirect to the registered target.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "unknown.test"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound || w.Header().Get("Location") != "" {
		t.Fatal(w.Code, w.Header())
	}
	// The destination's own TLS policy determines the final scheme.
	c.TLS.Domains = []string{"old.test"}
	r = httptest.NewRequest(http.MethodGet, "/docs", nil)
	r.Host = "old.test"
	w = httptest.NewRecorder()
	redirectHTTPS(c, handler).ServeHTTP(w, r)
	if w.Header().Get("Location") != "http://example.test/docs" {
		t.Fatal(w.Header())
	}
}

func TestDomainRedirectRejectsUnsafeRules(t *testing.T) {
	for name, redirects := range map[string]map[string]string{
		"unknown source":      {"unknown.test": "example.test"},
		"unknown destination": {"www.example.test": "unknown.test"},
		"self":                {"example.test": "example.test"},
		"chain":               {"www.example.test": "example.test", "example.test": "last.test"},
		"cycle":               {"www.example.test": "example.test", "example.test": "www.example.test"},
		"URL destination":     {"www.example.test": "https://example.test/elsewhere"},
		"invalid source":      {"evil.test/path": "example.test"},
	} {
		t.Run(name, func(t *testing.T) {
			file := testConfig(t, ":80")
			c, _, err := Load(file, nil)
			if err != nil {
				t.Fatal(err)
			}
			c.Sites["www.example.test"] = c.Sites["example.test"]
			c.Sites["last.test"] = c.Sites["example.test"]
			c.Redirects = redirects
			data, _ := json.Marshal(c)
			if err := os.WriteFile(file, data, 0644); err != nil {
				t.Fatal(err)
			}
			if _, _, err := Load(file, nil); err == nil {
				t.Fatal("accepted unsafe host redirect")
			}
		})
	}
}
