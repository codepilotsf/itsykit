package site

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// buildSessionApp compiles the independent stdlib sample app for this platform.
func buildSessionApp(t *testing.T) string {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain unavailable")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository root")
	}
	repo := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	out := filepath.Join(t.TempDir(), "session")
	cmd := exec.Command(goTool, "build", "-o", out, "./samples/session-app")
	cmd.Dir = repo
	cmd.Env = os.Environ()
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building sample app: %v\n%s", err, combined)
	}
	return out
}

func copyExecutable(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, data, 0500); err != nil {
		t.Fatal(err)
	}
}

func cookieNamed(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func cookieValue(w *httptest.ResponseRecorder, name string) string {
	if c := cookieNamed(w, name); c != nil {
		return c.Value
	}
	return ""
}

func TestCustomAppSessionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the sample custom app")
	}
	exe := buildSessionApp(t)
	root := fixture(t)
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"session":{"mount":"/session"}}}`)
	put(t, root, "itsy/apps/session/app.json", `{"name":"session","version":"0.1.0","protocol":2,"executable":{"`+runtime.GOOS+"-"+runtime.GOARCH+`":"session"}}`)
	copyExecutable(t, exe, filepath.Join(root, "itsy", "apps", "session", "session"))
	put(t, root, "session/index.html", `<p id="state">{session.data.authenticated}</p><p id="invalid">{session.invalid.action}</p><form method="post" action="/session">{csrf}<input name="action" value="login"></form>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)

	ctx, cancel := context.WithCancel(context.Background())
	stop := s.StartApps(ctx, nil, nil)
	defer func() { cancel(); stop() }()
	waitReady := func() {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, running, ok := s.appFor(p, "/session"); ok && running != nil && running.manager != nil && running.manager.Client() != nil {
				return
			}
			if time.Now().After(deadline) {
				t.Fatal("custom app did not become ready")
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	waitReady()

	const sessionCookie = "__Host-flea-session-session"
	get := func(cookies ...*http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("GET", "https://example.test/session", nil)
		for _, c := range cookies {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	post := func(form url.Values, csrf *http.Cookie, session *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", "https://example.test/session", strings.NewReader(form.Encode()))
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if csrf != nil {
			r.AddCookie(csrf)
		}
		if session != nil {
			r.AddCookie(session)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}

	// Unauthenticated read calls the app and renders false.
	w := get()
	if w.Code != 200 || !strings.Contains(w.Body.String(), "false") {
		t.Fatalf("initial read: %d %s", w.Code, w.Body.String())
	}
	formCookie := w.Result().Cookies()[0]
	csrf := formToken(w.Body.String())
	if csrf == "" {
		t.Fatal("no CSRF token rendered")
	}

	// Login sets one app session cookie and renders authenticated.
	w = post(url.Values{"action": {"login"}, "_csrf": {csrf}}, formCookie, nil)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	session := cookieNamed(w, sessionCookie)
	if session == nil || session.Value == "" {
		t.Fatalf("login did not set a session cookie: %v", w.Result().Cookies())
	}
	if !session.HttpOnly || !session.Secure || session.Path != "/" || session.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unsafe session cookie: %+v", session)
	}
	if n := len(w.Result().Cookies()); n > 2 {
		t.Fatalf("unexpected cookie count %d", n)
	}
	if strings.Contains(w.Body.String(), session.Value) {
		t.Fatal("session token rendered into the page")
	}

	// The token authenticates a later read and never appears in the body.
	w = get(session)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "true") || strings.Contains(w.Body.String(), session.Value) {
		t.Fatalf("authenticated read: %d %s", w.Code, w.Body.String())
	}

	// Rotation replaces the token; the old one is revoked.
	w = post(url.Values{"action": {"rotate"}, "_csrf": {csrf}}, formCookie, session)
	rotated := cookieNamed(w, sessionCookie)
	if rotated == nil || rotated.Value == session.Value || rotated.Value == "" {
		t.Fatalf("rotation did not replace the token: %v", w.Result().Cookies())
	}
	if !strings.Contains(get(session).Body.String(), "false") {
		t.Fatal("rotated token still authenticates")
	}
	if !strings.Contains(get(rotated).Body.String(), "true") {
		t.Fatal("rotated token does not authenticate")
	}

	// Concurrent responses each carry at most one session cookie. Requests above
	// the bounded per-client concurrency fail closed with the generic 503 rather
	// than waiting in a queue.
	var wg sync.WaitGroup
	var okCount atomic.Int32
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out := post(url.Values{"action": {"rotate"}, "_csrf": {csrf}}, formCookie, rotated)
			switch out.Code {
			case 200:
				okCount.Add(1)
				if n := len(out.Result().Cookies()); n != 1 {
					t.Errorf("concurrent rotate set %d cookies", n)
				}
			case 503:
				if n := len(out.Result().Cookies()); n != 0 {
					t.Errorf("unavailable rotate set %d cookies", n)
				}
			default:
				t.Errorf("concurrent rotate: %d", out.Code)
			}
		}()
	}
	wg.Wait()
	if okCount.Load() == 0 {
		t.Fatal("no concurrent rotate succeeded")
	}
	if !strings.Contains(get(formCookie, rotated).Body.String(), "false") {
		t.Fatal("concurrent rotation left the original token valid")
	}

	// Duplicate and cross-app cookies are rejected, not forwarded.
	r := httptest.NewRequest("GET", "https://example.test/session", nil)
	r.AddCookie(rotated)
	r.AddCookie(rotated)
	dup := httptest.NewRecorder()
	s.ServeHTTP(dup, r)
	if !strings.Contains(dup.Body.String(), "false") {
		t.Fatal("duplicate session cookie accepted")
	}
	other := &http.Cookie{Name: "__Host-flea-session-other", Value: rotated.Value}
	if !strings.Contains(get(other).Body.String(), "false") {
		t.Fatal("cross-app cookie accepted")
	}
	malformed := &http.Cookie{Name: sessionCookie, Value: "short"}
	if !strings.Contains(get(malformed).Body.String(), "false") {
		t.Fatal("malformed session cookie accepted")
	}

	// Logout clears the cookie.
	w = post(url.Values{"action": {"logout"}, "_csrf": {csrf}}, formCookie, rotated)
	if cleared := cookieNamed(w, sessionCookie); cleared == nil || cleared.MaxAge >= 0 {
		t.Fatalf("logout did not clear the cookie: %v", w.Result().Cookies())
	}

	// Expiry is owned by the app: a short-lived session stops authenticating.
	w = post(url.Values{"action": {"login-short"}, "_csrf": {csrf}}, formCookie, nil)
	short := cookieNamed(w, sessionCookie)
	if short == nil {
		t.Fatalf("short login set no cookie: %v", w.Result().Cookies())
	}
	time.Sleep(1500 * time.Millisecond)
	w = get(short)
	if !strings.Contains(w.Body.String(), "false") {
		t.Fatalf("expired session still authenticates: %s", w.Body.String())
	}
	if cleared := cookieNamed(w, sessionCookie); cleared == nil || cleared.MaxAge >= 0 {
		t.Fatalf("expired session was not cleared: %v", w.Result().Cookies())
	}

	// Persistence across an app restart: a live token survives a new child.
	w = post(url.Values{"action": {"login"}, "_csrf": {csrf}}, formCookie, nil)
	persistent := cookieNamed(w, sessionCookie)
	if persistent == nil {
		t.Fatal("restart login set no cookie")
	}
	cancel()
	stop()
	ctx, cancel = context.WithCancel(context.Background())
	stop = s.StartApps(ctx, nil, nil)
	waitReady()
	if !strings.Contains(get(persistent).Body.String(), "true") {
		t.Fatal("session did not survive an app restart")
	}
	cancel()
	stop()
}
