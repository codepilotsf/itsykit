package contact

import (
	"context"
	"errors"
	"flea/internal/mailer"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type sendFunc func(context.Context, mailer.Message) error

func (f sendFunc) Send(c context.Context, m mailer.Message) error { return f(c, m) }
func request() *http.Request {
	r := httptest.NewRequest("POST", "https://example.test"+Path, strings.NewReader("name=Visitor&email=visitor%40example.test&message=Hello"))
	r.Header.Set("Origin", "https://example.test")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}
func TestDeliveryAndLimits(t *testing.T) {
	calls := 0
	h := New(Config{To: "owner@example.test"}, "site", sendFunc(func(ctx context.Context, m mailer.Message) error {
		calls++
		if m.To != "owner@example.test" || m.ReplyTo != "visitor@example.test" || !strings.Contains(m.Text, "Hello") {
			t.Fatal(m)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("missing deadline")
		}
		return nil
	}), &Budget{}, false, nil)
	for i := 0; i < 4; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, request())
		want := 200
		if i == 3 {
			want = 429
		}
		if w.Code != want {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if calls != 3 {
		t.Fatal(calls)
	}
}
func TestRejectedRequestsNeverSend(t *testing.T) {
	cases := []struct {
		name   string
		change func(*http.Request)
		status int
	}{
		{"foreign origin", func(r *http.Request) { r.Header.Set("Origin", "https://other.test") }, 403},
		{"missing origin", func(r *http.Request) { r.Header.Del("Origin") }, 403},
		{"http", func(r *http.Request) { r.TLS = nil }, 403},
		{"query recipient", func(r *http.Request) { r.URL.RawQuery = "to=evil@example.test" }, 400},
		{"method", func(r *http.Request) { r.Method = "GET" }, 405},
		{"upload", func(r *http.Request) { r.Header.Set("Content-Type", "multipart/form-data") }, 415},
		{"encoded path", func(r *http.Request) { r.URL.RawPath = "/%5fflea/contact" }, 404},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := New(Config{To: "owner@example.test"}, "site", sendFunc(func(context.Context, mailer.Message) error { t.Fatal("sent rejected request"); return nil }), &Budget{}, false, nil)
			r := request()
			tc.change(r)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	for _, body := range []string{"name=A&email=a%40b.test&message=X&to=evil", "name=A&name=B&email=a%40b.test&message=X", "name=A&email=a%40b.test&message=X&site_id=other", "name=A&email=invalid&message=X", "name=A&email=a%40b.test&message=" + strings.Repeat("X", 33000)} {
		h := New(Config{To: "owner@example.test"}, "site", nil, &Budget{}, false, nil)
		r := request()
		r.Body = http.NoBody
		r = httptest.NewRequest("POST", "https://example.test"+Path, strings.NewReader(body))
		r.Header = request().Header
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
}
func TestFailureAndHoneypot(t *testing.T) {
	h := New(Config{To: "owner@example.test"}, "site", sendFunc(func(context.Context, mailer.Message) error { return errors.New("secret provider response") }), &Budget{}, false, nil)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, request())
		if w.Code != 502 || strings.Contains(w.Body.String(), "secret") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	r := httptest.NewRequest("POST", "https://example.test"+Path, strings.NewReader("name=A&email=a%40b.test&message=X&website=spam"))
	r.Header = request().Header
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
func TestBudget(t *testing.T) {
	b := &Budget{}
	now := time.Now()
	for i := 0; i < 4; i++ {
		if !b.acquire(now) {
			t.Fatal("early limit")
		}
	}
	if b.acquire(now) {
		t.Fatal("concurrency limit")
	}
	for i := 0; i < 4; i++ {
		b.release()
	}
	for i := 4; i < 200; i++ {
		if !b.acquire(now) {
			t.Fatal(i)
		}
		b.release()
	}
	if b.acquire(now) {
		t.Fatal("hourly limit")
	}
	if !b.acquire(now.Add(time.Hour)) {
		t.Fatal("did not recover")
	}
	b.release()
}

func TestConcurrentRequestsAndCancellation(t *testing.T) {
	started := make(chan struct{})
	h := New(Config{To: "owner@example.test"}, "site", sendFunc(func(ctx context.Context, _ mailer.Message) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}), &Budget{}, false, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	go func() { w := httptest.NewRecorder(); h.ServeHTTP(w, request().WithContext(ctx)); done <- w.Code }()
	<-started
	w := httptest.NewRecorder()
	h.ServeHTTP(w, request())
	if w.Code != 429 {
		t.Fatal(w.Code)
	}
	cancel()
	if code := <-done; code != 502 {
		t.Fatal(code)
	}
	if h.active || h.budget.active != 0 {
		t.Fatal("capacity leaked")
	}
}
func TestConfig(t *testing.T) {
	for _, s := range []string{"Owner <owner@example.test>", "a@example.test\r\nBcc: evil@example.test", "invalid"} {
		c := Config{To: s}
		if c.Validate() == nil {
			t.Fatal(s)
		}
	}
}
