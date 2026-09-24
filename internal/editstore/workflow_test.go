package editstore

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/editor"
	"flea/internal/mailer"
)

type testSender chan mailer.Message

func (s testSender) Send(_ context.Context, m mailer.Message) error { s <- m; return nil }

func TestHTTPSLoginEditAndPublicSnapshot(t *testing.T) {
	store, public := fixture(t, true)
	inbox := make(testSender, 1)
	dashboard := editor.New(editor.Options{Emails: store.Emails, SessionPath: filepath.Join(store.Root, "_", "sessions.json"), Name: "Test", Sender: inbox, Backend: store, Render: public.RenderAdmin})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if editor.IsPath(r.URL.Path) {
			dashboard.ServeHTTP(w, r)
		} else {
			public.ServeHTTP(w, r)
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Jar, _ = cookiejar.New(nil)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	call := func(method, path string, values url.Values) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, strings.NewReader(values.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		if method == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Origin", server.URL)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode, string(data)
	}
	field := func(body, name string) string {
		t.Helper()
		match := regexp.MustCompile(`name="` + name + `" value="([^"]+)"`).FindStringSubmatch(body)
		if len(match) != 2 {
			t.Fatalf("missing %s", name)
		}
		return match[1]
	}
	status, body := call("GET", "/admin", nil)
	if status != 200 {
		t.Fatal(status)
	}
	csrf := field(body, "csrf")
	status, _ = call("POST", "/admin/code", url.Values{"csrf": {csrf}, "email": {"owner@example.test"}})
	if status != 200 {
		t.Fatal(status)
	}
	var code string
	select {
	case mail := <-inbox:
		code = regexp.MustCompile(`\b[0-9]{6}\b`).FindString(mail.Text)
	case <-time.After(time.Second):
		t.Fatal("no email")
	}
	status, _ = call("POST", "/admin/verify", url.Values{"csrf": {csrf}, "code": {code}})
	if status != 303 {
		t.Fatal(status)
	}
	status, body = call("GET", "/admin", nil)
	if status != 200 {
		t.Fatal(status)
	}
	csrf, version := field(body, "csrf"), field(body, "version")
	status, _ = call("POST", "/admin/save", url.Values{"csrf": {csrf}, "version": {version}, "id": {"abcdefghij"}, "content": {"Live <strong>owner edit</strong><script>bad()</script>"}})
	if status != 303 {
		t.Fatal(status)
	}
	status, body = call("GET", "/", nil)
	if status != 200 || !strings.Contains(body, "Live <strong>owner edit</strong>") || strings.Contains(body, "<script>") {
		t.Fatal(status, body)
	}
	doc, err := store.Read()
	if err != nil || doc.Store.Pages["/"].Fleas["abcdefghij"].Content != "Live <strong>owner edit</strong>" {
		t.Fatal(doc, err)
	}
	status, _ = call("POST", "/admin/save", url.Values{"csrf": {csrf}, "version": {version}, "id": {"abcdefghij"}, "content": {"Stale overwrite"}})
	if status != 409 {
		t.Fatal("stale form accepted", status)
	}
	status, _ = call("POST", "/admin/logout", url.Values{"csrf": {csrf}})
	if status != 303 {
		t.Fatal(status)
	}
	_, body = call("GET", "/admin", nil)
	if strings.Contains(body, "Save content") {
		t.Fatal("logout did not revoke access")
	}
	t.Log("HTTPS smoke passed: code delivery, login, CSRF form, sanitized save, live public snapshot, durable content, stale edit rejection, logout")
}
