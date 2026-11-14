package editor

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"itsy/internal/mailer"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/templates"
)

type inbox chan mailer.Message

func testEditor(t *testing.T, options Options) *Handler {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	if err := site.SetSiteName(root, "FleaCMS"); err != nil {
		t.Fatal(err)
	}
	if err := site.IgnoreSessionFiles(root); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	options.Render = site.NewServer(p).RenderAdmin
	if options.SessionPath == "" {
		options.SessionPath = filepath.Join(root, "itsy", "runtime", "sessions.json")
	}
	return New(options)
}

func TestFirstOwnerSetupIsLocalOnly(t *testing.T) {
	for _, local := range []bool{true, false} {
		h := testEditor(t, Options{Name: "New site", Local: local})
		w := request(h, "GET", "/admin", "localhost", nil)
		if local {
			if w.Code != 503 || !strings.Contains(w.Body.String(), "itsy admin admin@example.com") || strings.Contains(w.Body.String(), "restart itsy dev") {
				t.Fatal(w.Code, w.Body.String())
			}
			w = request(h, "GET", "/admin", "external.test", nil)
			if w.Code != 403 || strings.Contains(w.Body.String(), "itsy admin") {
				t.Fatal("local setup instructions exposed to an external host")
			}
		} else if w.Code != 404 {
			t.Fatal("production without owners must remain disabled", w.Code)
		}
	}
}

func TestOwnerOnboardingDoesNotRevealAccess(t *testing.T) {
	h, mail := fixture(t)
	var responses []string
	for _, email := range []string{"owner@example.test", "unknown@example.test"} {
		w := request(h, "GET", "/admin", "example.test", nil)
		if w.Header().Get("Referrer-Policy") != "same-origin" {
			t.Fatal("browser form POST must retain its Origin")
		}
		if !strings.Contains(w.Body.String(), "Sign in to FleaCMS") || !strings.Contains(w.Body.String(), "No password needed") || !strings.Contains(w.Body.String(), `class="admin-access"`) {
			t.Fatal(w.Body.String())
		}
		c := cookie(t, w, h.cookieName("login"))
		w = request(h, "POST", "/admin/code", "example.test", url.Values{"csrf": {c.Value}, "email": {email}}, c)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "Check your spam folder") {
			t.Fatal(w.Code, w.Body.String())
		}
		responses = append(responses, strings.ReplaceAll(w.Body.String(), c.Value, "CSRF"))
	}
	if responses[0] != responses[1] {
		t.Fatal("code-request page reveals whether an address has access")
	}
	code(t, mail)
}

func TestMissingEmailSettingsInstructions(t *testing.T) {
	for _, local := range []bool{true, false} {
		h := testEditor(t, Options{Emails: []string{"owner@example.test"}, Local: local})
		w := request(h, "GET", "/admin", "localhost", nil)
		if w.Code != http.StatusServiceUnavailable || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
			t.Fatal(w.Code, w.Header())
		}
		for _, text := range []string{"itsy email", "site&#39;s folder"} {
			if strings.Contains(w.Body.String(), text) != local {
				t.Fatalf("local=%v, unexpected instructions for %q: %s", local, text, w.Body.String())
			}
		}
	}
}

func (i inbox) Send(_ context.Context, m mailer.Message) error { i <- m; return nil }

type memoryStore struct {
	saves int
	mu    sync.Mutex
}

func (s *memoryStore) Read() (Document, error) {
	return Document{Store: snips.NewStore(), Routes: []string{"/", "/about"}, Version: "revision"}, nil
}
func (s *memoryStore) Save(Update) error { s.mu.Lock(); defer s.mu.Unlock(); s.saves++; return nil }
func fixture(t *testing.T) (*Handler, inbox) {
	t.Helper()
	mail := make(inbox, 20)
	return testEditor(t, Options{Emails: []string{"owner@example.test"}, Name: "My site", Sender: mail, Backend: &memoryStore{}}), mail
}
func request(h *Handler, method, path, host string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "https://"+host+path, strings.NewReader(form.Encode()))
	r.Host = host
	r.TLS = &tls.ConnectionState{}
	r.RemoteAddr = "192.0.2.1:1234"
	if method == "POST" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Origin", "https://"+host)
	}
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func cookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("missing cookie %s: %v", name, w.Header())
	return nil
}
func code(t *testing.T, mail inbox) string {
	t.Helper()
	select {
	case m := <-mail:
		if m.To != "owner@example.test" {
			t.Fatal(m.To)
		}
		return regexp.MustCompile(`\b[0-9]{6}\b`).FindString(m.Text)
	case <-time.After(time.Second):
		t.Fatal("no login email")
		return ""
	}
}
func begin(t *testing.T, h *Handler, mail inbox) (*http.Cookie, string) {
	t.Helper()
	w := request(h, "GET", "/admin", "example.test", nil)
	c := cookie(t, w, h.cookieName("login"))
	w = request(h, "POST", "/admin/code", "example.test", url.Values{"csrf": {c.Value}, "email": {"owner@example.test"}}, c)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	return c, code(t, mail)
}
func login(t *testing.T, h *Handler, mail inbox) *http.Cookie {
	t.Helper()
	c, value := begin(t, h, mail)
	w := request(h, "POST", "/admin/verify", "example.test", url.Values{"csrf": {c.Value}, "code": {value}}, c)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w.Header().Get("Location") != "/" {
		t.Fatal("login must return to homepage", w.Header())
	}
	return cookie(t, w, h.cookieName("session"))
}
func TestLoginEditingLogoutAndCookieIsolation(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteStrictMode || c.Path != "/" || c.Domain != "" || c.MaxAge != 30*86400 || c.Expires.IsZero() {
		t.Fatal(c)
	}
	w := request(h, "GET", "/admin", "example.test", nil, c)
	if w.Code != 303 || w.Header().Get("Location") != "/" {
		t.Fatal("signed-in owner should return to site", w.Code)
	}
	s := h.sessions[key(c.Value)]
	oldSave := request(h, "POST", "/admin/save", "example.test", url.Values{"csrf": {s.csrf}}, c)
	if oldSave.Code != 404 {
		t.Fatal("retired dashboard save remains exposed")
	}

	w = request(h, "GET", "/admin", "other.test", nil, c)
	if w.Code == 303 {
		t.Fatal("session crossed hosts")
	}
	w = request(h, "POST", "/admin/logout", "example.test", url.Values{"csrf": {s.csrf}}, c)
	if w.Code != 303 {
		t.Fatal(w.Code)
	}
	w = request(h, "GET", "/admin", "example.test", nil, c)
	if w.Code == 303 {
		t.Fatal("logged-out session accepted")
	}
}
func TestCodesExpireLimitAttemptsAndAreSingleUse(t *testing.T) {
	for _, mode := range []string{"expiry", "attempts", "replay"} {
		t.Run(mode, func(t *testing.T) {
			h, mail := fixture(t)
			c, value := begin(t, h, mail)
			if mode == "expiry" {
				future := time.Now().Add(codeLife + time.Second)
				h.now = func() time.Time { return future }
			}
			if mode == "attempts" {
				for i := 0; i < 5; i++ {
					request(h, "POST", "/admin/verify", "example.test", url.Values{"csrf": {c.Value}, "code": {"invalid"}}, c)
				}
			}
			w := request(h, "POST", "/admin/verify", "example.test", url.Values{"csrf": {c.Value}, "code": {value}}, c)
			if mode == "replay" {
				if w.Code != 303 {
					t.Fatal(w.Code)
				}
				w = request(h, "POST", "/admin/verify", "example.test", url.Values{"csrf": {c.Value}, "code": {value}}, c)
			}
			if w.Code != 401 {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
}
func TestConcurrentCodeConsumption(t *testing.T) {
	h, mail := fixture(t)
	c, value := begin(t, h, mail)
	statuses := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func() {
			statuses <- request(h, "POST", "/admin/verify", "example.test", url.Values{"csrf": {c.Value}, "code": {value}}, c).Code
		}()
	}
	a, b := <-statuses, <-statuses
	if !((a == 303 && b == 401) || (a == 401 && b == 303)) {
		t.Fatal(a, b)
	}
}
func TestSessionExpiry(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	future := time.Now().Add(h.sessionDuration() + time.Second)
	h.now = func() time.Time { return future }
	w := request(h, "GET", "/admin", "example.test", nil, c)
	if w.Code == 303 {
		t.Fatal("expired session accepted")
	}
}
func TestUnlistedAddressRateLimitsAndOrigin(t *testing.T) {
	h, mail := fixture(t)
	w := request(h, "GET", "/admin", "example.test", nil)
	c := cookie(t, w, h.cookieName("login"))
	for i := 0; i < 4; i++ {
		w = request(h, "POST", "/admin/code", "example.test", url.Values{"csrf": {c.Value}, "email": {"stranger@example.test"}}, c)
		want := 200
		if i == 3 {
			want = 429
		}
		if w.Code != want {
			t.Fatal(w.Code)
		}
	}
	select {
	case <-mail:
		t.Fatal("sent to unlisted address")
	default:
	}
	r := httptest.NewRequest("POST", "https://example.test/admin/code", strings.NewReader("csrf="+c.Value))
	r.AddCookie(c)
	r.TLS = &tls.ConnectionState{}
	r.Header.Set("Origin", "https://attacker.test")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	r = httptest.NewRequest("GET", "http://example.test/admin", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("trusted forwarded TLS")
	}
}
func TestDashboardEscapesSiteName(t *testing.T) {
	h, _ := fixture(t)
	h.options.Name = `<script>alert(1)</script>`
	w := request(h, "GET", "/admin", "example.test", nil)
	if strings.Contains(w.Body.String(), "<script>") {
		t.Fatal("unescaped site name")
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatal("missing CSP")
	}
}
