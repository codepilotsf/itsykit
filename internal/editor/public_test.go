package editor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/site"
	"itsy/templates"
)

func TestLegacyToolbarRemainsUsableWithAuthenticatedCSP(t *testing.T) {
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "site", "admin", "_toolbar.html")
	markup, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.ReplaceAll(string(markup), "itsy-", "flea-")
	if err := os.WriteFile(file, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := site.LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	h, mail := fixture(t)
	h.options.Public = public.ServeOwner
	cookie := login(t, h, mail)
	r := httptest.NewRequest("GET", "https://example.test/", nil)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.ServePublic(w, r, public)
	_, tail, _ := strings.Cut(legacy, "<script>")
	initializer, _, _ := strings.Cut(tail, "</script>")
	sum := sha256.Sum256([]byte(initializer))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `id="flea-owner-root"`) || !strings.Contains(w.Header().Get("Content-Security-Policy"), "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'") {
		t.Fatal("legacy toolbar lost its supported initializer", w.Code, w.Header())
	}
}

func TestAuthenticatedPublicPreview(t *testing.T) {
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(`<article snip><p>Real page content</p></article>`), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := site.LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	h, mail := fixture(t)
	h.options.Public = public.ServeOwner
	c := login(t, h, mail)
	visit := func(path string, cookie *http.Cookie, host string, secure bool) *httptest.ResponseRecorder {
		scheme := "http"
		if secure {
			scheme = "https"
		}
		r := httptest.NewRequest("GET", scheme+"://"+host+path, nil)
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServePublic(w, r, public)
		return w
	}
	for _, tc := range []struct {
		name, host string
		cookie     *http.Cookie
		secure     bool
	}{
		{"anonymous", "example.test", nil, true},
		{"forged", "example.test", &http.Cookie{Name: c.Name, Value: "forged"}, true},
		{"other host", "other.test", c, true},
		{"plain HTTP", "example.test", c, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := visit("/", tc.cookie, tc.host, tc.secure)
			if w.Code != 200 || strings.Contains(w.Body.String(), "itsy-owner-root") || strings.Contains(w.Body.String(), "snip-start:") || strings.Contains(w.Body.String(), "csrf-token") {
				t.Fatal("private markup exposed", w.Code)
			}
		})
	}
	w := visit("/", c, "example.test", true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `id="itsy-owner-root"`) || !strings.Contains(w.Body.String(), "<!--snip-start:") || !strings.Contains(w.Body.String(), "Real page content") {
		t.Fatal("missing authenticated editor", w.Code)
	}
	if strings.Contains(w.Body.String(), "setInterval(()=>fetch('/_flea/reload'") {
		t.Fatal("reload can discard preview edits")
	}
	if w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(w.Header())
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src https://example.test/admin/assets/editor.js") {
		t.Fatal("external editor blocked")
	}
	// Layout-authored critical CSS is allowed; editor decoration adds no styles.
	anonymous := visit("/", nil, "example.test", true).Body.String()
	if strings.Count(w.Body.String(), "<style") != strings.Count(anonymous, "<style") {
		t.Fatal("editor injected inline styling")
	}
	for _, forbidden := range []string{"data-page=", "data-photos=", "data-csrf=", "DOMPurify", "data-flea-navigation"} {
		if strings.Contains(w.Body.String(), forbidden) {
			t.Fatalf("source contains injected %s", forbidden)
		}
	}
	if !strings.Contains(w.Body.String(), `name="csrf-token"`) || !strings.Contains(w.Body.String(), `src="/admin/assets/editor.js"`) {
		t.Fatal("explicit editor integration missing")
	}
	asset := visit("/css/global.css", c, "example.test", true)
	if asset.Code != 200 || strings.Contains(asset.Body.String(), "itsy-owner-root") {
		t.Fatal("asset decorated")
	}
	if len(asset.Result().Cookies()) != 0 {
		t.Fatal("asset request renewed session")
	}
	if h.options.Backend.(*memoryStore).saves != 0 {
		t.Fatal("preview saved content")
	}
	if err := os.WriteFile(filepath.Join(root, "site", "main.layout.html"), []byte("<!doctype html><html><body>{slot}</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(visit("/", c, "example.test", true).Body.String(), "itsy-owner-root") {
		t.Fatal("toolbar injected without template opt-in")
	}
	s := h.sessions[key(c.Value)]
	request(h, "POST", "/admin/logout", "example.test", url.Values{"csrf": {s.csrf}}, c)
	if strings.Contains(visit("/", c, "example.test", true).Body.String(), "itsy-owner-root") {
		t.Fatal("logged-out session retained editor")
	}
}
