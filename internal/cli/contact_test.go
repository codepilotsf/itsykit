package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/appbundle"
)

func enableTestContact(t *testing.T, root string) {
	t.Helper()
	file := filepath.Join(root, ".flea.config.json")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err = json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	cfg["apps"] = map[string]any{"contact": map[string]any{"mount": "/contact", "email": map[string]any{"to": "contact@example.test"}}}
	data, err = json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
}

// Real child-process coverage when testing a bundled build. The rest of the
// admin/LAN workflow also runs in source-only builds.
func submitTestContact(t *testing.T, h http.Handler, origin, peer, message string) bool {
	t.Helper()
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("Contact bundle missing")
		}
		t.Log("Contact process check requires scripts/build.sh")
		return false
	}
	get := httptest.NewRequest("GET", origin+"/contact", nil)
	get.RemoteAddr = peer
	w := httptest.NewRecorder()
	h.ServeHTTP(w, get)
	match := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(w.Body.String())
	if w.Code != 200 || len(match) != 2 || len(w.Result().Cookies()) == 0 {
		t.Fatal("missing Contact form", w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	post := func(email string) *httptest.ResponseRecorder {
		fields := url.Values{"_csrf": {match[1]}, "name": {"Visitor"}, "email": {email}, "message": {message}}
		r := httptest.NewRequest("POST", origin+"/contact", strings.NewReader(fields.Encode()))
		r.RemoteAddr = peer
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	deadline := time.Now().Add(8 * time.Second)
	for {
		w = post("invalid")
		if w.Code == 422 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Contact not ready", w.Code, w.Body.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
	w = post("visitor@example.test")
	if w.Code != 303 || w.Header().Get("Location") != "/contact/thanks" {
		t.Fatal(w.Code, w.Body.String())
	}
	return true
}
