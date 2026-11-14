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

	"itsy/internal/contact"
	"itsy/internal/devaccess"
	"itsy/internal/editor"
	"itsy/internal/site"
)

// Exercise the real dev admin/store/public pipeline as a different LAN device.
// Only email delivery is captured; authentication and publication are unchanged.
func TestLANDevWorkflow(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	root := filepath.Join(t.TempDir(), "site")
	if err := createSite(root, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := site.AddAdmin(root, "owner@example.test"); err != nil {
		t.Fatal(err)
	}
	enableTestContact(t, root)
	p, err := site.LoadDev(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	mail := make(previewInbox, 4)
	admin := devAdmin(root, "LAN test", public, nil, mail)
	stopApps := public.StartApps(context.Background(), mail, &contact.Budget{})
	defer stopApps()
	h := devaccess.WrapLAN(admin, []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	request := func(handler http.Handler, method, path string, values url.Values, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "http://asahi:3000"+path, strings.NewReader(values.Encode()))
		r.RemoteAddr = "100.79.67.33:54321"
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("X-CSRF-Token", values.Get("csrf"))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	const origin = "http://asahi:3000"
	req := func(method, path string, values url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
		return request(h, method, path, values, cookie, origin)
	}
	cookie := func(w *httptest.ResponseRecorder, name string) *http.Cookie {
		t.Helper()
		for _, c := range w.Result().Cookies() {
			if c.Name == name {
				return c
			}
		}
		t.Fatalf("missing %s cookie: %d", name, w.Code)
		return nil
	}
	if w := request(admin, "GET", "/admin", nil, nil, origin); w.Code != 403 {
		t.Fatal("LAN enabled without middleware", w.Code)
	}
	if w := request(public, "POST", "/contact", nil, nil, origin); w.Code != 403 {
		t.Fatal("contact LAN enabled by default", w.Code)
	}
	productionAdmin := devaccess.WrapLAN(editor.New(editor.Options{Emails: []string{"owner@example.test"}}), []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	if w := request(productionAdmin, "GET", "/admin", nil, nil, origin); w.Code != 403 || !strings.Contains(w.Body.String(), "HTTPS") {
		t.Fatal("LAN context bypassed production HTTPS")
	}
	productionForm := devaccess.WrapLAN(site.NewProductionServer(p, nil), []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	if w := request(productionForm, "POST", "/contact", nil, nil, origin); w.Code != 403 {
		t.Fatal("LAN context bypassed production contact HTTPS")
	}
	w := req("GET", "/admin", nil, nil)
	login := cookie(w, "flea_dev_login")
	w = req("POST", "/admin/code", url.Values{"csrf": {login.Value}, "email": {"owner@example.test"}}, login)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var code string
	select {
	case message := <-mail:
		code = regexp.MustCompile(`\b[0-9]{6}\b`).FindString(message.Text)
	case <-time.After(2 * time.Second):
		t.Fatal("no sign-in preview")
	}
	w = req("POST", "/admin/verify", url.Values{"csrf": {login.Value}, "code": {code}}, login)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	session := cookie(w, "flea_dev_session")
	if session.Secure || !session.HttpOnly || session.SameSite != http.SameSiteStrictMode {
		t.Fatal("unexpected dev cookie policy")
	}
	w = req("GET", "/admin", nil, session)
	csrfMatch := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindStringSubmatch(w.Body.String())
	if len(csrfMatch) != 2 {
		t.Fatal("missing admin CSRF", w.Body.String())
	}
	csrf := csrfMatch[1]
	for _, method := range []string{"GET", "POST"} {
		if w := req(method, "/admin/theme", url.Values{"csrf": {csrf}}, session); w.Code != 404 {
			t.Fatal("removed Theme endpoint", method, w.Code)
		}
	}
	if w := req("POST", "/admin/activity", url.Values{"csrf": {csrf}}, session); w.Code != 204 {
		t.Fatal("activity", w.Code)
	}
	if submitTestContact(t, h, origin, "192.168.1.9:54321", "LAN contact test") {
		select {
		case message := <-mail:
			if message.To != "contact@example.test" {
				t.Fatal("wrong recipient")
			}
		case <-time.After(time.Second):
			t.Fatal("missing contact preview")
		}
	}
	before := req("GET", "/_flea/reload", nil, nil).Body.String()
	if err := os.WriteFile(filepath.Join(root, "site/test.html"), []byte("<p>LAN reload test</p>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	if w := req("GET", "/_flea/reload", nil, nil); w.Code != 200 || w.Body.String() == before {
		t.Fatal("reload revision did not change")
	}
	if w := req("POST", "/", nil, nil); w.Code != 405 || !strings.Contains(w.Body.String(), "Check the form") {
		t.Fatal("LAN diagnostic unavailable", w.Code)
	}
	for _, path := range []string{"/itsy.config.json", "/itsy/runtime/sessions.json", "/mailers/sign-in.email.html"} {
		if w := req("GET", path, nil, nil); w.Code != 404 {
			t.Fatalf("private path %s: %d", path, w.Code)
		}
	}
	if w := req("POST", "/admin/logout", url.Values{"csrf": {csrf}}, session); w.Code != 303 {
		t.Fatal("logout", w.Code)
	}
	if w := req("GET", "/", nil, session); strings.Contains(w.Body.String(), "itsy-owner-root") {
		t.Fatal("logged-out session still works")
	}
}
