package editstore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/editor"
	"flea/internal/mailer"
	"golang.org/x/net/html"
)

type testMailbox chan mailer.Message

func (m testMailbox) Send(_ context.Context, message mailer.Message) error { m <- message; return nil }

func TestAuthenticatedPagePublication(t *testing.T) {
	s, public := fixture(t, false)
	mail := make(testMailbox, 1)
	h := editor.New(editor.Options{Emails: s.Emails, Name: "Test", Backend: s, Sender: mail, Public: public.ServeOwner, Render: public.RenderAdmin, SessionPath: filepath.Join(t.TempDir(), "sessions.json")})
	request := func(method, path string, values url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "https://example.test"+path, strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("X-CSRF-Token", values.Get("csrf"))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		if editor.IsPath(path) {
			h.ServeHTTP(w, r)
		} else {
			h.ServePublic(w, r, public)
		}
		return w
	}
	findCookie := func(w *httptest.ResponseRecorder) *http.Cookie {
		t.Helper()
		for _, c := range w.Result().Cookies() {
			if strings.Contains(c.Name, "flea") {
				return c
			}
		}
		t.Fatal("no cookie", w.Code)
		return nil
	}
	w := request("GET", "/admin", nil, nil)
	login := findCookie(w)
	w = request("POST", "/admin/code", url.Values{"csrf": {login.Value}, "email": {s.Emails[0]}}, login)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var code string
	select {
	case message := <-mail:
		code = regexp.MustCompile(`\b[0-9]{6}\b`).FindString(message.Text)
	case <-time.After(time.Second):
		t.Fatal("no code")
	}
	w = request("POST", "/admin/verify", url.Values{"csrf": {login.Value}, "code": {code}}, login)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	session := findCookie(w)
	w = request("GET", "/", nil, session)
	root, err := html.Parse(strings.NewReader(w.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	attrs := map[string]string{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Data == "meta" {
			name, content := "", ""
			for _, a := range n.Attr {
				if a.Key == "name" {
					name = a.Val
				}
				if a.Key == "content" {
					content = a.Val
				}
			}
			if name == "csrf-token" {
				attrs["csrf"] = content
			}
		}
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == "flea-owner-root" {
				for _, a := range n.Attr {
					attrs[a.Key] = a.Val
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	var edit editor.PageEdit
	w = request("GET", "/admin/editor-context?route=/", url.Values{"csrf": {attrs["csrf"]}}, session)
	var contextData struct {
		Page editor.PageEdit `json:"page"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &contextData); err != nil {
		t.Fatal("missing snapshot metadata", err)
	}
	edit = contextData.Page
	if edit.Version != attrs["data-version"] {
		t.Fatal("metadata newer than rendered page")
	}
	doc, err := s.Read()
	if err != nil || edit.Version != doc.Version {
		t.Fatal("rendered snapshot version mismatch", err)
	}
	edit.Title = "Saved through inline editor"
	edit.Description = "Updated details"
	edit.Changes = []editor.ContentEdit{{ID: "abcdefghij", Content: "Saved heading"}}
	data, _ := json.Marshal(edit)
	values := url.Values{"csrf": {attrs["csrf"]}, "edit": {string(data)}}
	w = request("POST", "/admin/page", values, session)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var receipt editor.PageResult
	if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil || receipt.Page.Version == "" || receipt.Page.Version == edit.Version || len(receipt.Page.Changes) != 1 || receipt.Page.Changes[0].Content != "Saved heading" {
		t.Fatal("missing canonical save receipt", err, w.Body.String())
	}
	w = request("GET", "/", nil, nil)
	if !strings.Contains(w.Body.String(), "Saved heading") || !strings.Contains(w.Body.String(), edit.Title) || strings.Contains(w.Body.String(), "flea-owner-root") {
		t.Fatal("publication failed", w.Body.String())
	}
	w = request("POST", "/admin/page", values, session)
	if w.Code != 409 {
		t.Fatal("stale browser draft accepted", w.Code)
	}
	edit.Version = receipt.Page.Version
	edit.Changes[0].Content = "Second saved heading"
	data, _ = json.Marshal(edit)
	values.Set("edit", string(data))
	if w = request("POST", "/admin/page", values, session); w.Code != 200 {
		t.Fatal("receipt version cannot save again", w.Code, w.Body.String())
	}
}
