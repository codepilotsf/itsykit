package appapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const secret = "private-test-credential-12345678901234567890"

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)
	c, err := New(s.URL, "contact", secret)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

func TestTransportContract(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+secret || r.Header.Get("Cookie") != "" {
			t.Error("wrong credentials")
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/_flea/handshake" {
			io.WriteString(w, `{"name":"contact","version":"0.1.0","protocol":1}`)
			return
		}
		b, _ := io.ReadAll(r.Body)
		if r.Method == "GET" {
			if len(b) != 0 || r.URL.RawQuery != "tag=one&tag=two&literal=%2B" || r.URL.Path != "/list" {
				t.Error("GET changed", r.URL, string(b))
			}
			io.WriteString(w, `{"data":{"title":"Articles"}}`)
			return
		}
		if r.URL.Path != "/" || string(b) != `{"email":"bad"}` || r.Header.Get("Content-Type") != "application/json" {
			t.Error("POST changed", r.URL, string(b))
		}
		w.WriteHeader(422)
		io.WriteString(w, `{"data":{"email":"bad"},"invalid":{"email":"Use a valid email."}}`)
	})
	m := Mount{"contact", "/deep/contact"}
	if _, _, err := c.Request(context.Background(), m, "GET", m.Path, "", nil); err == nil || calls.Load() != 0 {
		t.Fatal("unverified app called")
	}
	if err := c.Handshake(context.Background()); err != nil {
		t.Fatal(err)
	}
	e, status, err := c.Request(context.Background(), m, "GET", m.Path+"/list", "tag=one&tag=two&literal=%2B", nil)
	if err != nil || status != 200 || e.Data["title"] != "Articles" {
		t.Fatal(e, status, err)
	}
	e, status, err = c.Request(context.Background(), m, "POST", m.Path, "", map[string]any{"email": "bad"})
	if err != nil || status != 422 || e.Invalid["email"] == "" {
		t.Fatal(e, status, err)
	}
	before := calls.Load()
	for _, route := range []string{m.Path + "/_flea/handshake", m.Path + "-extra"} {
		if _, _, err := c.Request(context.Background(), m, "GET", route, "", nil); err == nil {
			t.Fatal("private route accepted")
		}
	}
	if calls.Load() != before {
		t.Fatal("unsafe request forwarded")
	}
}

func TestBadHandshakes(t *testing.T) {
	for _, body := range []string{`{"name":"wrong","version":"1","protocol":1}`, `{"name":"contact","version":"1","protocol":2}`, `{"name":"contact","protocol":1}`, `{"Name":"contact","version":"1","protocol":1}`, strings.Repeat(" ", HandshakeLimit+1)} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, body)
		})
		if err := c.Handshake(context.Background()); err == nil || c.ready.Load() {
			t.Fatal("bad handshake accepted")
		}
	}
}

func TestBadResponses(t *testing.T) {
	for _, tc := range []struct {
		status     int
		kind, body string
	}{
		{200, "text/html", `{}`}, {200, "application/json", " "}, {200, "application/json", `{"data":null}`},
		{200, "application/json", `{"redirect":"thanks"}`}, {200, "application/json", `{"flash":{"notice":null}}`},
		{200, "application/json", `{"error":"failure"}`}, {422, "application/json", `{}`}, {503, "application/json", `{}`},
		{401, "application/json", `{"error":"secret diagnostic"}`}, {200, "application/json", strings.Repeat(" ", ResponseLimit+1)},
	} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", tc.kind)
			w.WriteHeader(tc.status)
			io.WriteString(w, tc.body)
		})
		c.ready.Store(true)
		_, status, err := c.Request(context.Background(), Mount{"contact", "/contact"}, "GET", "/contact", "", nil)
		if status != 503 || err != ErrUnavailable {
			t.Fatal("unsafe response accepted", tc.status, err)
		}
	}
}

func TestRedirectAndTimeout(t *testing.T) {
	var leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Add(1) }))
	defer target.Close()
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.Redirect(w, r, target.URL, 307)
	})
	c.ready.Store(true)
	if _, _, err := c.Request(context.Background(), Mount{"contact", "/contact"}, "POST", "/contact", "", map[string]any{"email": "x"}); err == nil || leaked.Load() != 0 {
		t.Fatal("followed redirect")
	}
	var calls atomic.Int32
	slow := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	})
	slow.ready.Store(true)
	slow.timeout = 30 * time.Millisecond
	if _, _, err := slow.Request(context.Background(), Mount{"contact", "/contact"}, "POST", "/contact", "", map[string]any{}); err != ErrUnavailable || calls.Load() != 1 {
		t.Fatal("timeout or POST retry", err, calls.Load())
	}
}

func TestConcurrencyBounds(t *testing.T) {
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{}`)
	})
	c.ready.Store(true)
	done := make(chan error, 4)
	for range 4 {
		go func() {
			_, _, err := c.Request(context.Background(), Mount{"contact", "/contact"}, "GET", "/contact", "", nil)
			done <- err
		}()
	}
	for range 4 {
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("requests did not enter")
		}
	}
	_, _, err := c.Request(context.Background(), Mount{"contact", "/contact"}, "GET", "/contact", "", nil)
	close(release)
	if err != ErrUnavailable {
		t.Fatal("capacity limit not enforced")
	}
	for range 4 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func TestEndpointPolicy(t *testing.T) {
	for _, endpoint := range []string{"http://example.test:5000", "http://localhost:5000", "https://user:pass@example.test", "https://example.test/path", "https://example.test?secret=x", "https://example.test#fragment"} {
		if c, err := New(endpoint, "contact", secret); err == nil {
			c.Close()
			t.Fatal("unsafe endpoint accepted", endpoint)
		}
	}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"name":"contact","version":"1","protocol":1}`)
	}))
	defer s.Close()
	c, err := New(s.URL, "contact", secret)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if c.Handshake(context.Background()) == nil {
		t.Fatal("untrusted certificate accepted")
	}
}
