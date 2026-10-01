package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/contact"
	"flea/internal/localenv"
	"flea/internal/mailer"
	"flea/internal/site"
)

func TestDevAdminReloadsAccessAndMail(t *testing.T) {
	for _, key := range []string{"RESEND_API_KEY", "FLEA_EMAIL_FROM", "FLEA_EMAIL_NAME"} {
		old, present := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if present {
				os.Setenv(key, old)
			} else {
				os.Unsetenv(key)
			}
		})
	}
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	handler := devAdmin(root, "Test", public, nil, nil)
	get := func(want string) {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "http://localhost/admin", nil))
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("want %q: %s", want, w.Body.String())
		}
	}
	get("Add the first admin")
	if _, _, err := site.AddAdmin(root, "owner@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	get("Email is not configured")
	values := map[string]string{"RESEND_API_KEY": "re_test12345678", "FLEA_EMAIL_FROM": "hello@example.com"}
	if err := localenv.Save(root, values, []string{"RESEND_API_KEY", "FLEA_EMAIL_FROM"}); err != nil {
		t.Fatal(err)
	}
	get("Send sign-in code")
	if err := os.WriteFile(filepath.Join(root, ".flea.config.json"), []byte(`{"app":{"site_name":"Test"},"admin":{"emails":[]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	get("Add the first admin")
}

type previewInbox chan mailer.Message

func (p previewInbox) Send(_ context.Context, m mailer.Message) error { p <- m; return nil }

func TestDevPreviewSignsInAndRendersContactWithoutCredentials(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	if _, _, err := site.AddAdmin(root, "owner@example.test"); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !p.Dev.EmailPreview {
		t.Fatal("starter must enable previews")
	}
	public := site.NewServer(p)
	mail := make(previewInbox, 2)
	h := devAdmin(root, "Test", public, nil, mail)
	request := func(handler http.Handler, method, path string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:3000"+path, strings.NewReader(form.Encode()))
		r.RemoteAddr = "127.0.0.1:54321"
		if method == "POST" {
			r.Header.Set("Origin", "http://127.0.0.1:3000")
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	w := request(h, "GET", "/admin", nil, nil)
	var login *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "flea_dev_login" {
			login = c
		}
	}
	if login == nil || !strings.Contains(w.Body.String(), "Send sign-in code") {
		t.Fatal("preview login unavailable", w.Code, w.Body.String())
	}
	w = request(h, "POST", "/admin/code", url.Values{"csrf": {login.Value}, "email": {"owner@example.test"}}, login)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var m mailer.Message
	select {
	case m = <-mail:
	case <-time.After(2 * time.Second):
		t.Fatal("no preview message")
	}
	code := regexp.MustCompile(`\b[0-9]{6}\b`).FindString(m.Text)
	if code == "" || !strings.Contains(m.HTML, code) {
		t.Fatal("missing rendered code")
	}
	w = request(h, "POST", "/admin/verify", url.Values{"csrf": {login.Value}, "code": {code}}, login)
	if w.Code != 303 {
		t.Fatal("preview code unusable", w.Code, w.Body.String())
	}
	form := contact.New(contact.Config{To: "contact@example.test", SuccessPath: "/thank-you"}, root, mail, &contact.Budget{}, true, nil)
	form.RenderEmail = public.RenderEmail
	w = request(form, "POST", contact.Path, url.Values{"name": {"Visitor"}, "email": {"visitor@example.test"}, "message": {"Hello from the local form"}}, nil)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case m = <-mail:
	case <-time.After(time.Second):
		t.Fatal("no contact preview")
	}
	if m.To != "contact@example.test" || m.ReplyTo != "visitor@example.test" || !strings.Contains(m.HTML, "Hello from the local form") {
		t.Fatal("contact headers/body missing")
	}
}
