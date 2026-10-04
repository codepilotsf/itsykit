package site

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFlashIsolationAndConsumption(t *testing.T) {
	s := NewProductionServer(mustLoad(t, fixture(t)), nil)
	origin := "https://example.test"
	initial := httptest.NewRequest("GET", origin+"/contact", nil)
	_, visitor, err := s.forms.Token(initial)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", origin+"/contact", nil)
	r.AddCookie(visitor)
	w := httptest.NewRecorder()
	s.storeFlash(w, r, "contact", "/contact/thanks", map[string]string{"notice": "Sent"})
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("missing flash")
	}
	flash := cookies[0]
	if !flash.Secure || !flash.HttpOnly || flash.SameSite != http.SameSiteLaxMode || flash.Path != "/" {
		t.Fatal("cookie policy")
	}
	get := func(method, host, route string, who *http.Cookie, app string, duplicate bool) map[string]string {
		request := httptest.NewRequest(method, host+route, nil)
		if who != nil {
			request.AddCookie(who)
		}
		c := *flash
		c.Name = s.flashCookie(app, "", 0).Name
		request.AddCookie(&c)
		if duplicate {
			request.AddCookie(&c)
		}
		values, finish := s.takeFlash(request, app)
		finish(httptest.NewRecorder(), false)
		return values
	}
	_, other, err := s.forms.Token(initial)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, host, route string
		who                 *http.Cookie
		app                 string
		duplicate           bool
	}{
		{"HEAD", origin, "/contact/thanks", visitor, "contact", false},
		{"GET", origin, "/about", visitor, "contact", false},
		{"GET", "https://alias.test", "/contact/thanks", visitor, "contact", false},
		{"GET", origin, "/contact/thanks", other, "contact", false},
		{"GET", origin, "/contact/thanks", nil, "contact", false},
		{"GET", origin, "/contact/thanks", visitor, "other", false},
		{"GET", origin, "/contact/thanks", visitor, "contact", true},
	} {
		if get(tc.method, tc.host, tc.route, tc.who, tc.app, tc.duplicate) != nil {
			t.Fatal("flash escaped scope", tc)
		}
	}
	request := httptest.NewRequest("GET", origin+"/contact/thanks", nil)
	request.AddCookie(visitor)
	request.AddCookie(flash)
	values, rollback := s.takeFlash(request, "contact")
	if values["notice"] != "Sent" {
		t.Fatal(values)
	}
	if values, _ := s.takeFlash(request, "contact"); values != nil {
		t.Fatal("concurrent claim")
	}
	rollback(httptest.NewRecorder(), false)
	values, commit := s.takeFlash(request, "contact")
	if values["notice"] != "Sent" {
		t.Fatal("failed render consumed flash")
	}
	out := httptest.NewRecorder()
	commit(out, true)
	if len(out.Result().Cookies()) != 1 || out.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("cookie not expired")
	}
	if values, _ := s.takeFlash(request, "contact"); values != nil {
		t.Fatal("replay")
	}
}

func TestFlashExpiryAndCapacity(t *testing.T) {
	s := NewProductionServer(mustLoad(t, fixture(t)), nil)
	r := httptest.NewRequest("GET", "https://example.test/contact", nil)
	_, cookie, err := s.forms.Token(r)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(cookie)
	for i := 0; i < 257; i++ {
		w := httptest.NewRecorder()
		s.storeFlash(w, r, "contact", "/contact/thanks", map[string]string{"notice": "Sent"})
		if i == 256 && len(w.Result().Cookies()) != 0 {
			t.Fatal("unbounded flash")
		}
	}
	if len(s.apps.flashes.entries) != 256 {
		t.Fatal("wrong capacity")
	}
	for _, entry := range s.apps.flashes.entries {
		entry.expires = time.Now().Add(-time.Second)
	}
	w := httptest.NewRecorder()
	s.storeFlash(w, r, "contact", "/contact/thanks", map[string]string{"notice": "New"})
	if len(s.apps.flashes.entries) != 1 || len(w.Result().Cookies()) != 1 {
		t.Fatal("expiry did not release capacity")
	}
}
