package site

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"flea/internal/appapi"
)

// hardeningApp starts a fake protocol service and returns a ready client.
func hardeningApp(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *appapi.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_flea/handshake" {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"name":"contact","version":"1","protocol":2}`)
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	client, err := appapi.New(server.URL, "contact", "private-app-credential-12345678901234567890")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	if err := client.Handshake(context.Background()); err != nil {
		t.Fatal(err)
	}
	return client
}

// TestBrowserCannotInjectAppMetadata proves the browser's own X-Flea-Session,
// Authorization and Cookie headers never reach the app or change its identity.
func TestBrowserCannotInjectAppMetadata(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<p>read</p>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)

	var got http.Header
	client := hardeningApp(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"data":{"ok":true}}`)
	})
	mount := appapi.Mount{Name: "contact", Path: "/s"}
	r := httptest.NewRequest("GET", "https://example.test/s", nil)
	r.Header.Set("X-Flea-Session", "browser-injected-token-0123456789")
	r.Header.Set("Authorization", "Bearer browser-injected")
	r.Header.Set("Cookie", "other=browser-injected-token-0123456789")
	w := httptest.NewRecorder()
	s.serveAppRequest(w, r, mount, client)
	if w.Code != 200 {
		t.Fatalf("read failed: %d %s", w.Code, w.Body.String())
	}
	if got.Get("X-Flea-Session") != "" {
		t.Fatalf("browser X-Flea-Session forwarded: %q", got.Get("X-Flea-Session"))
	}
	if got.Get("Authorization") != "Bearer private-app-credential-12345678901234567890" {
		t.Fatalf("browser Authorization leaked: %q", got.Get("Authorization"))
	}
	if got.Get("Cookie") != "" {
		t.Fatalf("browser cookie forwarded: %q", got.Get("Cookie"))
	}
}

// TestInvalidRedirectNeverCommits proves an app response with a bad destination
// fails closed and changes no cookie.
func TestInvalidRedirectNeverCommits(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<form method="post" action="/s">{csrf}</form>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	client := hardeningApp(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"redirect":"../escape","session":{"token":"session-token-0123456789abcdef","max_age":60}}`)
	})
	mount := appapi.Mount{Name: "contact", Path: "/s"}
	get := httptest.NewRequest("GET", "https://example.test/s", nil)
	w := httptest.NewRecorder()
	s.serveAppRequest(w, get, mount, client)
	csrf := formToken(w.Body.String())
	formCookie := w.Result().Cookies()[0]
	post := httptest.NewRequest("POST", "https://example.test/s", strings.NewReader(url.Values{"_csrf": {csrf}, "action": {"go"}}.Encode()))
	post.Header.Set("Origin", "https://example.test")
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.AddCookie(formCookie)
	w = httptest.NewRecorder()
	s.serveAppRequest(w, post, mount, client)
	if w.Code != 503 {
		t.Fatalf("invalid redirect status %d", w.Code)
	}
	if cookieNamed(w, "__Host-flea-session-contact") != nil {
		t.Fatal("invalid destination committed a session cookie")
	}
}

// TestInvalidEnvelopeNeverCommits proves an envelope with an unknown key is a
// service failure that changes no cookie, even when it also carries a session.
func TestInvalidEnvelopeNeverCommits(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<form method="post" action="/s">{csrf}</form>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	client := hardeningApp(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"data":{},"bogus":1,"session":{"token":"session-token-0123456789abcdef","max_age":60}}`)
	})
	mount := appapi.Mount{Name: "contact", Path: "/s"}
	get := httptest.NewRequest("GET", "https://example.test/s", nil)
	w := httptest.NewRecorder()
	s.serveAppRequest(w, get, mount, client)
	csrf := formToken(w.Body.String())
	formCookie := w.Result().Cookies()[0]
	post := httptest.NewRequest("POST", "https://example.test/s", strings.NewReader(url.Values{"_csrf": {csrf}, "action": {"go"}}.Encode()))
	post.Header.Set("Origin", "https://example.test")
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.AddCookie(formCookie)
	w = httptest.NewRecorder()
	s.serveAppRequest(w, post, mount, client)
	if w.Code != 503 {
		t.Fatalf("invalid envelope status %d", w.Code)
	}
	if cookieNamed(w, "__Host-flea-session-contact") != nil {
		t.Fatal("invalid envelope committed a session cookie")
	}
}

// TestCrossAppLiveTokenRejected proves a live token for one app is never read
// for another app, even though both are valid sessions for their owner.
func TestCrossAppLiveTokenRejected(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<p>read</p>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	const live = "session-token-0123456789abcdef"
	other := &http.Cookie{Name: "__Host-flea-session-other", Value: live}
	r := httptest.NewRequest("GET", "https://example.test/s", nil)
	r.AddCookie(other)
	if token := s.sessionToken(r, "contact"); token != "" {
		t.Fatalf("cross-app token read: %q", token)
	}
	// A valid cookie accompanied by a malformed duplicate of the same name must
	// fail closed even though r.Cookies() would drop the malformed fragment.
	r = httptest.NewRequest("GET", "https://example.test/s", nil)
	r.Header.Set("Cookie", "__Host-flea-session-contact="+live+"; __Host-flea-session-contact")
	if token := s.sessionToken(r, "contact"); token != "" {
		t.Fatalf("malformed duplicate accepted: %q", token)
	}
	// The production cookie is host-only, so it is never sent to another site.
	prod := s.sessionCookie("contact", live, 60)
	if prod.Domain != "" || !prod.Secure || prod.Path != "/" {
		t.Fatalf("unsafe production cookie: %+v", prod)
	}
}

// TestHeadCallsAppAsGetButSendsNoBody pins HEAD semantics: the app is asked for
// the GET representation, but the browser receives headers only.
func TestHeadCallsAppAsGetButSendsNoBody(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<p>read</p>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	var method string
	client := hardeningApp(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"data":{"ok":true}}`)
	})
	mount := appapi.Mount{Name: "contact", Path: "/s"}
	r := httptest.NewRequest("HEAD", "https://example.test/s", nil)
	w := httptest.NewRecorder()
	s.serveAppRequest(w, r, mount, client)
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("HEAD response %d body=%q", w.Code, w.Body.String())
	}
	if method != "GET" {
		t.Fatalf("HEAD reached the app as %q", method)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("HEAD content type %q", w.Header().Get("Content-Type"))
	}
}
