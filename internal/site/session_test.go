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

// TestSessionOnRedirectCommitsCookie checks that a session set alongside a
// redirect is committed before the 303, so the browser stores it.
func TestSessionOnRedirectCommitsCookie(t *testing.T) {
	root := fixture(t)
	put(t, root, "s/index.html", `<form method="post" action="/s">{csrf}<input name="action" value="go"></form>`)
	put(t, root, "s/next.html", `<p>next</p>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	const token = "session-token-0123456789abcdef"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/_flea/handshake" {
			io.WriteString(w, `{"name":"contact","version":"1","protocol":2}`)
			return
		}
		if r.Method == "GET" {
			io.WriteString(w, `{"data":{}}`)
			return
		}
		io.WriteString(w, `{"redirect":"next","session":{"token":"`+token+`","max_age":300}}`)
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
	if w.Code != 303 || w.Header().Get("Location") != "/s/next" {
		t.Fatalf("redirect: %d %q", w.Code, w.Header().Get("Location"))
	}
	cookie := cookieNamed(w, "__Host-flea-session-contact")
	if cookie == nil || cookie.Value != token || cookie.MaxAge != 300 {
		t.Fatalf("session cookie not committed with redirect: %v", w.Result().Cookies())
	}
}

// TestSessionCookieIsolation checks the cookie name policy: production uses the
// __Host- prefix per app, and development names include a site identifier so two
// sites on one loopback host never share a session cookie.
func TestSessionCookieIsolation(t *testing.T) {
	first := NewServer(mustLoad(t, fixture(t)))
	second := NewServer(mustLoad(t, fixture(t)))
	if first.sessionCookieName("alpha") == first.sessionCookieName("beta") {
		t.Fatal("apps share a session cookie name")
	}
	if first.sessionCookieName("alpha") == second.sessionCookieName("alpha") {
		t.Fatal("sites share a session cookie name")
	}
	prod := NewProductionServer(first.project, nil)
	if got := prod.sessionCookieName("alpha"); got != "__Host-flea-session-alpha" {
		t.Fatalf("production cookie name %q", got)
	}
}
