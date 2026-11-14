package site

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"itsy/internal/appapi"
)

// TestContactRouteMethodPreserved pins Contact's exact route/method behavior:
// only its mount accepts POST, and the thank-you page is read-only.
func TestContactRouteMethodPreserved(t *testing.T) {
	root := fixture(t)
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.test"}}}}`)
	put(t, root, "contact/index.html", `<form method="post" action="/contact">{csrf}</form>`)
	put(t, root, "contact/thank-you.html", `<p>Thanks</p>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	mount := appapi.Mount{Name: "contact", Path: "/contact"}
	for _, tc := range []struct {
		method, path string
		notFound     bool
	}{
		{"POST", "/contact/thank-you", true},
		{"POST", "/contact/list", true},
		{"GET", "/contact/thank-you", false},
		{"GET", "/contact", false},
		{"POST", "/contact", false},
	} {
		r := httptest.NewRequest(tc.method, "https://example.test"+tc.path, strings.NewReader(""))
		r.Header.Set("Origin", "https://example.test")
		w := httptest.NewRecorder()
		s.serveInstalledApp(w, r, mount)
		if (w.Code == http.StatusNotFound) != tc.notFound {
			t.Fatalf("%s %s -> %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestAppRoundTrip(t *testing.T) {
	root := fixture(t)
	const route = "/deep/contact"
	put(t, root, "deep/contact/index.html", `<form method="post" action="/deep/contact">{csrf}<input name="email" value="{contact.data.email}"><p>{contact.invalid.email}</p><p>{contact.error}</p></form>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	var calls atomic.Int32
	var fail atomic.Bool
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer private-app-credential-12345678901234567890" || r.Header.Get("Cookie") != "" {
			t.Error("visitor headers leaked")
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/_flea/handshake" {
			io.WriteString(w, `{"name":"contact","version":"0.1.0","protocol":2}`)
			return
		}
		calls.Add(1)
		if fail.Load() {
			io.WriteString(w, "bad JSON containing secret diagnostic")
			return
		}
		if r.URL.Path != "/" {
			t.Error("mount not stripped")
		}
		if r.Method == "GET" {
			io.WriteString(w, `{"data":{}}`)
			return
		}
		var fields map[string]any
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			t.Error(err)
		}
		if _, exists := fields["_csrf"]; exists {
			t.Error("token forwarded")
		}
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(appapi.Envelope{Data: fields, Invalid: map[string]string{"email": "Enter a valid email."}})
	}))
	defer api.Close()
	client, err := appapi.New(api.URL, "contact", "private-app-credential-12345678901234567890")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Handshake(context.Background()); err != nil {
		t.Fatal(err)
	}
	mount := appapi.Mount{Name: "contact", Path: route}
	r := httptest.NewRequest("GET", "https://example.test"+route, nil)
	w := httptest.NewRecorder()
	s.serveAppRequest(w, r, mount, client)
	if w.Code != 200 || len(w.Result().Cookies()) != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	token := formToken(w.Body.String())
	submit := func(value, csrf string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "https://example.test"+route, strings.NewReader(url.Values{"email": {value}, "_csrf": {csrf}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("Authorization", "Bearer visitor-secret")
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		s.serveAppRequest(w, r, mount, client)
		return w
	}
	w = submit(`"><script>alert(1)</script>{app.site_name}`, token)
	if w.Code != 422 || !strings.Contains(w.Body.String(), "Enter a valid email.") || strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Body.String(), "{app.site_name}") || formToken(w.Body.String()) != token {
		t.Fatal(w.Code, w.Body.String())
	}
	before := calls.Load()
	w = submit("x", "bad")
	if w.Code != 403 || calls.Load() != before {
		t.Fatal("bad CSRF reached app")
	}
	for _, route := range []string{"/deep/contact/_flea/handshake", "/deep/contact-extra", "/deep/contact/no-template"} {
		r := httptest.NewRequest("GET", "https://example.test"+route, nil)
		w := httptest.NewRecorder()
		s.serveAppRequest(w, r, mount, client)
		if w.Code != 404 || calls.Load() != before {
			t.Fatal("invalid route forwarded")
		}
	}
	fail.Store(true)
	w = submit("x", token)
	if w.Code != 503 || strings.Contains(w.Body.String(), "secret diagnostic") || !strings.Contains(w.Body.String(), appapi.ErrUnavailable.Error()) {
		t.Fatal("unsafe upstream error", w.Body.String())
	}
	if w := request(s, "GET", "https://example.test/about"); w.Code != 200 {
		t.Fatal("app failure affected ordinary page")
	}
	// A service/client alone never enables a public route without registration.
	if w := request(s, "POST", "https://example.test"+route); w.Code != 405 {
		t.Fatal("app route publicly enabled")
	}
	fail.Store(false)
	var wg sync.WaitGroup
	for _, value := range []string{"alice", "bob", "carol"} {
		wg.Go(func() {
			w := submit(value, token)
			if w.Code != 422 || !strings.Contains(w.Body.String(), `value="`+value+`"`) {
				t.Error("request data crossed renders", w.Body.String())
			}
		})
	}
	wg.Wait()
	if p.App["contact"] != nil || p.App["csrf"] != nil {
		t.Fatal("project mutated")
	}
}
