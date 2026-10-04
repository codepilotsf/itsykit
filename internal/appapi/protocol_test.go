package appapi

import (
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"flea/internal/formguard"
)

func TestMounts(t *testing.T) {
	good := []Mount{{"contact", "/deep/contact"}, {"blog", "/blog"}, {"events", "/blog-extra"}}
	if err := ValidateMounts(good, []string{"/css/site.css"}, nil); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]Mount{
		{{"app", "/contact"}}, {{"data", "/contact"}}, {{"admin", "/contact"}}, {{"__request", "/contact"}},
		{{"contact", "/"}}, {{"contact", "/admin"}}, {{"contact", "/admin/foo"}}, {{"contact", "/_flea"}},
		{{"contact", "/x/../contact"}}, {{"contact", "/contact/"}}, {{"contact", "/contact%2f"}}, {{"contact", "/css"}},
		{{"contact", "/a"}, {"blog", "/a/b"}}, {{"contact", "/a"}, {"contact", "/b"}}, {{"contact", "/download"}},
	} {
		if ValidateMounts(bad, []string{"/css/site.css"}, []string{"/download/releases"}) == nil {
			t.Fatal("bad mount accepted", bad)
		}
	}
	m := good[0]
	for route, want := range map[string]string{"/deep/contact": "/", "/deep/contact/list": "/list"} {
		got, err := m.ServicePath(route)
		if err != nil || got != want {
			t.Fatal(route, got, err)
		}
	}
	for _, route := range []string{"/deep/contact-extra", "/deep/contact/_flea/handshake", "/deep/contact/../admin", "/deep/contact/.secret", "/deep/contact/a.html", "/deep/contact/a%2fb"} {
		if _, err := m.ServicePath(route); err == nil {
			t.Fatal("unsafe route", route)
		}
	}
}

func TestFormContract(t *testing.T) {
	g := formguard.New("test", false)
	token, cookie, err := g.Token(httptest.NewRequest("GET", "https://example.test/contact", nil))
	if err != nil {
		t.Fatal(err)
	}
	values := url.Values{"_csrf": {token}, "email": {""}, "tags": {"one", "two"}, "address.city": {"NY"}, "items[0]": {"literal"}}
	r := httptest.NewRequest("POST", "https://example.test/contact?email=not-body&query=only", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "https://example.test")
	r.AddCookie(cookie)
	got, err := DecodeForm(httptest.NewRecorder(), r, g)
	want := map[string]any{"email": "", "tags": []string{"one", "two"}, "address.city": "NY", "items[0]": "literal"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal(got, err)
	}
	for _, tc := range []struct {
		body, kind string
		status     int
	}{
		{"email=x", "application/x-www-form-urlencoded", 403},
		{values.Encode() + "&_csrf=" + token, "application/x-www-form-urlencoded", 403},
		{values.Encode() + "&email=%ff", "application/x-www-form-urlencoded", 400},
		{values.Encode() + "&email=" + strings.Repeat("a", FormLimit), "application/x-www-form-urlencoded", 400},
		{values.Encode(), "multipart/form-data; boundary=test", 415},
		{`{}`, "application/json", 415},
	} {
		r := httptest.NewRequest("POST", "https://example.test/contact", strings.NewReader(tc.body))
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("Content-Type", tc.kind)
		r.AddCookie(cookie)
		_, err := DecodeForm(httptest.NewRecorder(), r, g)
		if e, ok := err.(*RequestError); !ok || e.Status != tc.status {
			t.Fatal(err, tc.status)
		}
	}
}

func TestStrictJSON(t *testing.T) {
	for _, body := range []string{"", " ", "null", "[]", `{} {}`, `{"data":null}`, `{"DATA":{}}`, `{"data":{"x":1,"x":2}}`, `{"data":{},"data":{}}`, `{"invalid":{"email":null}}`, `{"error":[]}`, `{"unknown":"value"}`, `{"data":{"nested":` + strings.Repeat("[", 66) + "0" + strings.Repeat("]", 66) + "}}"} {
		var e Envelope
		if err := decode([]byte(body), &e); err == nil {
			t.Fatal("invalid JSON accepted", body)
		}
	}
	var e Envelope
	if err := decode([]byte(`{"data":{"items":[1,true,null],"n":1234567890123456789},"invalid":{"email":"Required"}}`), &e); err != nil {
		t.Fatal(err)
	}
	if e.Values()["invalid"].(map[string]any)["email"] != "Required" {
		t.Fatal("messages not template-compatible")
	}
}

func FuzzProtocolJSON(f *testing.F) {
	for _, seed := range []string{`{"data":{}}`, " ", `{"invalid":{"x":null}}`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > ResponseLimit {
			return
		}
		var e Envelope
		_ = decode([]byte(body), &e)
	})
}
