package editor

import (
	"bytes"
	"encoding/json"
	"flea/internal/snips"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBlogSaveRequiresSessionOriginAndCSRF(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	csrf := h.sessions[key(c.Value)].csrf
	raw, _ := json.Marshal(BlogEdit{Version: "v1", Post: snips.BlogPost{ID: "Abcdef1234", Title: "Post", Slug: "post", Status: "draft"}, Action: "save"})
	for _, tc := range []struct {
		csrf, origin string
		logged       bool
		want         int
	}{{csrf, "https://example.test", true, 200}, {"wrong", "https://example.test", true, 403}, {csrf, "https://else.test", true, 403}, {csrf, "https://example.test", false, 403}} {
		r := httptest.NewRequest("POST", "https://example.test/admin/blog/save", bytes.NewBufferString(url.Values{"csrf": {tc.csrf}, "edit": {string(raw)}}.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Origin", tc.origin)
		if tc.logged {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	// Multipart authentication happens before parsing the upload body.
	for _, header := range []string{"", csrf} {
		var body bytes.Buffer
		m := multipart.NewWriter(&body)
		m.WriteField("csrf", csrf)
		m.WriteField("edit", string(raw))
		f, _ := m.CreateFormFile("unexpected", "image.png")
		f.Write([]byte("invalid"))
		m.Close()
		r := httptest.NewRequest("POST", "https://example.test/admin/blog/save", &body)
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("Content-Type", m.FormDataContentType())
		r.Header.Set("X-CSRF-Token", header)
		r.AddCookie(c)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := 400
		if header == "" {
			want = 403
		}
		if w.Code != want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
func TestBlogAdminRequiresAuthentication(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	for _, path := range []string{"/admin/blog", "/admin/blog/edit"} {
		if got := request(h, "GET", path, "example.test", nil).Code; got != 303 {
			t.Fatal(path, got)
		}
		if got := request(h, "GET", path, "example.test", nil, c).Code; got != 404 {
			t.Fatal("disabled blog admin", got)
		}
	}
}

type enabledBlogBackend struct{ Backend }

func (b enabledBlogBackend) Read() (Document, error) {
	return Document{Store: snips.NewStore(), Version: "revision", BlogMount: "/blog"}, nil
}
func TestBlogAdminUsesScopedScriptPolicy(t *testing.T) {
	h, mail := fixture(t)
	h.options.Backend = enabledBlogBackend{h.options.Backend}
	c := login(t, h, mail)
	w := request(h, "GET", "/admin/blog/edit", "example.test", nil, c)
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src https://example.test/admin/assets/blog.js") || !strings.Contains(w.Body.String(), `method="post" action="/admin/blog/save"`) {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
}
