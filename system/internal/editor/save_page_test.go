package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type invalidMediaBackend struct {
	Backend
	err error
}

func (b invalidMediaBackend) Save(Update) error { return b.err }

func TestPageMediaValidationErrorsAreSafe(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	csrf := h.sessions[key(c.Value)].csrf
	for _, tc := range []struct {
		err    error
		status int
		want   string
	}{
		{&MediaValidationError{Message: "MP4 must use fast start."}, 400, "MP4 must use fast start."},
		{fmt.Errorf("open /private/secret.json: denied"), 409, "Could not save."},
	} {
		h.options.Backend = invalidMediaBackend{Backend: h.options.Backend, err: tc.err}
		w := request(h, "POST", "/admin/page", "example.test", url.Values{"csrf": {csrf}, "edit": {`{"version":"v1","route":"/"}`}}, c)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.want) || strings.Contains(w.Body.String(), "secret.json") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestNoStandaloneMediaLibraryRoutes(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	csrf := h.sessions[key(c.Value)].csrf
	for _, path := range []string{"/admin/media", "/admin/videos"} {
		for _, method := range []string{"GET", "POST"} {
			w := request(h, method, path, "example.test", url.Values{"csrf": {csrf}, "edit": {`{"version":"v1"}`}}, c)
			if w.Code != 404 {
				t.Fatalf("%s %s = %d", method, path, w.Code)
			}
		}
	}
	if h.options.Backend.(*memoryStore).saves != 0 {
		t.Fatal("removed route saved content")
	}
}

func TestPageMediaUploadAuthorization(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	csrf := h.sessions[key(c.Value)].csrf
	for _, tc := range []struct {
		name, header, origin, field string
		cookie                      bool
		want                        int
	}{
		{"valid", csrf, "https://example.test", "media-Abcdef1234", true, 200},
		{"collection upload", csrf, "https://example.test", "media-Abcdef1234", true, 200},
		{"no header", "", "https://example.test", "media-Abcdef1234", true, 403},
		{"no session", csrf, "https://example.test", "media-Abcdef1234", false, 403},
		{"origin", csrf, "https://else.test", "media-Abcdef1234", true, 403},
		{"unexpected file", csrf, "https://example.test", "media-Zbcdef1234", true, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			m := multipart.NewWriter(&body)
			m.WriteField("csrf", csrf)
			payload := `{"version":"v1","route":"/","media":[{"id":"Abcdef1234","alt":"A photo"}]}`
			if tc.name == "collection upload" {
				payload = strings.Replace(payload, `"alt":"A photo"`, `"alt":"A photo","library":true`, 1)
			}
			m.WriteField("edit", payload)
			f, _ := m.CreateFormFile(tc.field, "../../ignored.png")
			f.Write([]byte("bytes validated by backend"))
			m.Close()
			r := httptest.NewRequest("POST", "https://example.test/admin/page", &body)
			r.Header.Set("Content-Type", m.FormDataContentType())
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("X-CSRF-Token", tc.header)
			if tc.cookie {
				r.AddCookie(c)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
		})
	}
	if h.options.Backend.(*memoryStore).saves != 2 {
		t.Fatal("unauthorized upload reached storage")
	}
}

func TestPageSaveRequiresSessionOriginAndCSRF(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	session := h.sessions[key(c.Value)]
	data, _ := json.Marshal(PageEdit{Version: "v1", Route: "/", Title: "New title", Changes: []ContentEdit{{ID: "a7Km2pQ9xB", Content: "Hello"}}})
	for _, tc := range []struct {
		name, origin, csrf string
		cookie             bool
		status             int
	}{
		{"valid", "https://example.test", session.csrf, true, 200},
		{"missing origin", "", session.csrf, true, 403},
		{"other origin", "https://evil.test", session.csrf, true, 403},
		{"bad csrf", "https://example.test", "bad", true, 403},
		{"no session", "https://example.test", session.csrf, false, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := url.Values{"csrf": {tc.csrf}, "edit": {string(data)}}
			r := httptest.NewRequest("POST", "https://example.test/admin/page", strings.NewReader(values.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("Origin", tc.origin)
			if tc.cookie {
				r.AddCookie(c)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	if h.options.Backend.(*memoryStore).saves != 1 {
		t.Fatal("unauthorized write")
	}
	for _, payload := range []string{`{}`, `{"version":"v","route":"/","unexpected":true}`, string(data) + `{}`} {
		w := request(h, "POST", "/admin/page", "example.test", url.Values{"csrf": {session.csrf}, "edit": {payload}}, c)
		if w.Code != 400 {
			t.Fatal("invalid payload accepted", w.Code)
		}
	}
	w := request(h, "POST", "/admin/page", "example.test", url.Values{"csrf": {session.csrf}, "edit": {strings.Repeat("x", 2<<20)}}, c)
	if w.Code != 400 || h.options.Backend.(*memoryStore).saves != 1 {
		t.Fatal("oversized request reached store", w.Code)
	}
}
