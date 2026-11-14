package formguard

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"itsy/internal/devaccess"
)

func post(cookie *http.Cookie, token string) (*http.Request, url.Values) {
	r := httptest.NewRequest("POST", "https://example.test/contact", nil)
	r.Header.Set("Origin", "https://example.test")
	r.AddCookie(cookie)
	return r, url.Values{Field: {token}}
}

func TestVisitorBinding(t *testing.T) {
	g := New("site-one", false)
	now := time.Unix(1800000000, 0)
	g.now = func() time.Time { return now }
	token, cookie, err := g.Token(httptest.NewRequest("GET", "https://example.test/contact", nil))
	if err != nil || cookie == nil {
		t.Fatal(err)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.Domain != "" || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode || cookie.Name != "__Host-flea-form" {
		t.Fatal("unsafe cookie attributes")
	}
	r, fields := post(cookie, token)
	if err := g.Verify(r, fields); err != nil {
		t.Fatal(err)
	}
	again, replacement, err := g.Token(r)
	if err != nil || replacement != nil || again != token {
		t.Fatal("valid visitor rotated", err)
	}
	otherToken, otherCookie, err := g.Token(httptest.NewRequest("GET", "https://example.test/contact", nil))
	if err != nil || otherToken == token || otherCookie.Value == cookie.Value {
		t.Fatal("visitors shared identity")
	}
	cases := map[string]func(*http.Request, url.Values){
		"missing":          func(r *http.Request, v url.Values) { delete(v, Field) },
		"duplicate":        func(r *http.Request, v url.Values) { v.Add(Field, token) },
		"wrong token":      func(r *http.Request, v url.Values) { v.Set(Field, otherToken) },
		"no cookie":        func(r *http.Request, v url.Values) { r.Header.Del("Cookie") },
		"duplicate cookie": func(r *http.Request, v url.Values) { r.AddCookie(cookie) },
		"wrong cookie":     func(r *http.Request, v url.Values) { r.Header.Del("Cookie"); r.AddCookie(otherCookie) },
		"tampered cookie": func(r *http.Request, v url.Values) {
			c := *cookie
			c.Value = strings.Repeat("a", 96)
			r.Header.Del("Cookie")
			r.AddCookie(&c)
		},
		"missing origin":   func(r *http.Request, v url.Values) { r.Header.Del("Origin") },
		"null origin":      func(r *http.Request, v url.Values) { r.Header.Set("Origin", "null") },
		"duplicate origin": func(r *http.Request, v url.Values) { r.Header.Add("Origin", "https://example.test") },
		"foreign origin":   func(r *http.Request, v url.Values) { r.Header.Set("Origin", "https://other.test") },
		"cross host": func(r *http.Request, v url.Values) {
			r.Host = "other.test"
			r.Header.Set("Origin", "https://other.test")
		},
		"query token": func(r *http.Request, v url.Values) { r.URL.RawQuery = "%5fcsrf=" + token },
		"no TLS":      func(r *http.Request, v url.Values) { r.TLS = nil; r.Header.Set("X-Forwarded-Proto", "https") },
		"GET":         func(r *http.Request, v url.Values) { r.Method = "GET" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			r, v := post(cookie, token)
			change(r, v)
			if g.Verify(r, v) == nil {
				t.Fatal("accepted")
			}
		})
	}
	if New("site-two", false).Verify(r, fields) == nil {
		t.Fatal("cross-site accepted")
	}
	if New("site-one", false).Verify(r, fields) == nil {
		t.Fatal("restart accepted old token")
	}
	now = now.Add(Lifetime)
	if g.Verify(r, fields) == nil {
		t.Fatal("expired token accepted")
	}
	if _, c, err := g.Token(r); err != nil || c == nil {
		t.Fatal("expired visitor not renewed", err)
	}
}

func TestDevOriginGate(t *testing.T) {
	g := New("local", true)
	r := httptest.NewRequest("GET", "http://localhost:3000/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	_, c, err := g.Token(r)
	if err != nil || c.Secure || !c.HttpOnly || c.Name == New("other", true).cookieName() {
		t.Fatal("dev cookie", err)
	}
	for _, target := range []string{"http://evil.test:3000/", "http://192.168.1.2:3000/"} {
		q := httptest.NewRequest("GET", target, nil)
		q.RemoteAddr = "127.0.0.1:1234"
		if _, _, err := g.Token(q); err == nil {
			t.Fatal("unapproved host accepted")
		}
	}
	r.RemoteAddr = "100.79.67.33:1234"
	if _, _, err := g.Token(r); err == nil {
		t.Fatal("unapproved peer accepted")
	}
	called := false
	h := devaccess.WrapLAN(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, _, err := g.Token(r); err != nil {
			t.Error(err)
		}
	}), []string{"192.168.1.2"}, []string{"asahi", "asahi.local"}, 3000)
	r = httptest.NewRequest("GET", "http://asahi:3000/", nil)
	r.RemoteAddr = "100.79.67.33:1234"
	h.ServeHTTP(httptest.NewRecorder(), r)
	if !called {
		t.Fatal("approved LAN blocked")
	}
}

func TestConcurrentIssuance(t *testing.T) {
	g := New("shared", false)
	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			token, c, err := g.Token(httptest.NewRequest("GET", "https://example.test/", nil))
			if err != nil {
				t.Error(err)
				return
			}
			r, v := post(c, token)
			if err := g.Verify(r, v); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}
