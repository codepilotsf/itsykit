package editstore

import (
	"context"
	"encoding/json"
	"html"
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
	dashboard := editor.New(editor.Options{Emails: store.Emails, SessionPath: filepath.Join(store.Root, "flea", "runtime", "sessions.json"), Name: "Test", Sender: inbox, Backend: store, Render: public.RenderAdmin, Public: public.ServeOwner})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if editor.IsPath(r.URL.Path) {
			dashboard.ServeHTTP(w, r)
		} else {
			dashboard.ServePublic(w, r, public)
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Jar, _ = cookiejar.New(nil)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	var csrf string
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
		req.Header.Set("X-CSRF-Token", csrf)
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
	csrf = field(body, "csrf")
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
	status, body = call("GET", "/", nil)
	if status != 200 {
		t.Fatal(status)
	}
	attribute := func(name string) string {
		t.Helper()
		match := regexp.MustCompile(`data-` + name + `="([^"]*)"`).FindStringSubmatch(body)
		if len(match) != 2 {
			t.Fatal("missing", name)
		}
		return html.UnescapeString(match[1])
	}
	meta := regexp.MustCompile(`name="csrf-token" content="([^"]+)"`).FindStringSubmatch(body)
	if len(meta) != 2 {
		t.Fatal("missing CSRF meta")
	}
	csrf = html.UnescapeString(meta[1])
	version := attribute("version")
	status, body = call("GET", "/admin/editor-context?route=/", nil)
	var edit editor.PageEdit
	var contextData struct {
		Page editor.PageEdit `json:"page"`
	}
	if err := json.Unmarshal([]byte(body), &contextData); err != nil {
		t.Fatal(err)
	}
	edit = contextData.Page
	if status != 200 || edit.Version != version {
		t.Fatal("wrong context snapshot", status)
	}
	edit.Changes = []editor.ContentEdit{{ID: "abcdefghij", Content: "Live owner edit<script>bad()</script>"}}
	data, _ := json.Marshal(edit)
	values := url.Values{"csrf": {csrf}, "edit": {string(data)}}
	status, _ = call("POST", "/admin/page", values)
	if status != 200 {
		t.Fatal(status)
	}
	status, body = call("GET", "/", nil)
	if status != 200 || !strings.Contains(body, "Live owner edit") || strings.Contains(body, "bad()") {
		t.Fatal(status, body)
	}
	doc, err := store.Read()
	if err != nil || doc.Store.Pages["/"].Snips["abcdefghij"].Content != "Live owner edit" {
		t.Fatal(doc, err)
	}
	status, _ = call("POST", "/admin/page", values)
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
