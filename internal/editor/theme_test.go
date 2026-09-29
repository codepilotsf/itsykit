package editor

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestThemeSaveAuthorization(t *testing.T) {
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	s := h.sessions[key(cookie.Value)]
	for _, tc := range []struct {
		origin, csrf string
		logged       bool
		want         int
	}{{"https://example.test", s.csrf, true, 204}, {"", s.csrf, true, 403}, {"https://other.test", s.csrf, true, 403}, {"https://example.test", "bad", true, 403}, {"https://example.test", s.csrf, false, 403}} {
		values := url.Values{"csrf": {tc.csrf}, "edit": {`{"version":"v","changes":{"--brand-1":"red"}}`}}
		r := httptest.NewRequest("POST", "https://example.test/admin/theme", strings.NewReader(values.Encode()))
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if tc.logged {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if h.options.Backend.(*memoryStore).saves != 1 {
		t.Fatal("unauthorized theme write")
	}
	w := request(h, "GET", "/admin/theme", "example.test", nil)
	if w.Code == 200 {
		t.Fatal("theme exposed without authentication")
	}
}
