package host

import (
	"encoding/json"
	"net/http/httptest"
	"os"
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
