package contact

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	contactapp "itsy/apps/contact"
	"itsy/internal/mailer"
	"time"
)

type sendFunc func(context.Context, mailer.Message) error

func (f sendFunc) Send(c context.Context, m mailer.Message) error { return f(c, m) }

var submission = contactapp.Submission{Name: "Visitor", Email: "visitor@example.test", Message: "Hello"}

const notification = `{"name":"Visitor","email":"visitor@example.test","message":"Hello"}`

func notify(d *Delivery, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, r)
	return w
}
func delivery(sender mailer.Sender) *Delivery {
	return &Delivery{Email: Email{To: "owner@example.test", From: "support@example.test"}, Sender: sender, Budget: &Budget{}, Render: func(_ string, fields map[string]string) (mailer.Message, error) {
		return mailer.Message{Text: fields["message"]}, nil
	}}
}
func TestDeliveryAndLimits(t *testing.T) {
	calls := 0
	d := delivery(sendFunc(func(ctx context.Context, m mailer.Message) error {
		calls++
		if m.To != "owner@example.test" || m.From != "support@example.test" || m.ReplyTo != submission.Email || m.Text != "Hello" {
			t.Fatal(m)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("missing deadline")
		}
		return nil
	}))
	if w := notify(d, notification); w.Code != 403 {
		t.Fatal("unreserved notification", w.Code)
	}
	for i := 0; i < 4; i++ {
		release, ok := d.Reserve(context.Background(), "127.0.0.1", submission)
		if i == 3 {
			if ok {
				release()
				t.Fatal("missing IP limit")
			}
			break
		}
		if !ok {
			t.Fatal("early limit")
		}
		if w := notify(d, strings.Replace(notification, "Hello", "Other", 1)); w.Code != 403 {
			t.Fatal("changed submission", w.Code)
		}
		if w := notify(d, notification); w.Code != 204 {
			t.Fatal(w.Code, w.Body.String())
		}
		if w := notify(d, notification); w.Code != 403 {
			t.Fatal("duplicate notification", w.Code)
		}
		release()
	}
	if calls != 3 {
		t.Fatal(calls)
	}
}
func TestFailureIsSingleUse(t *testing.T) {
	d := delivery(sendFunc(func(context.Context, mailer.Message) error { return errors.New("secret provider response") }))
	release, ok := d.Reserve(context.Background(), "127.0.0.1", submission)
	if !ok {
		t.Fatal("reserve")
	}
	defer release()
	w := notify(d, notification)
	if w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(w.Code, w.Body.String())
	}
	if notify(d, notification).Code != 403 {
		t.Fatal("failed delivery was retried")
	}
}
func TestConcurrentRequestsAndCancellation(t *testing.T) {
	started := make(chan struct{})
	d := delivery(sendFunc(func(ctx context.Context, _ mailer.Message) error { close(started); <-ctx.Done(); return ctx.Err() }))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	release, ok := d.Reserve(ctx, "127.0.0.1", submission)
	if !ok {
		t.Fatal("reserve")
	}
	defer release()
	done := make(chan int, 1)
	go func() { done <- notify(d, notification).Code }()
	<-started
	if undo, ok := d.Reserve(ctx, "127.0.0.2", submission); ok {
		undo()
		t.Fatal("concurrent submission admitted")
	}
	cancel()
	if code := <-done; code != 503 {
		t.Fatal(code)
	}
	if d.Budget.active != 0 {
		t.Fatal("capacity leaked")
	}
}
func TestInvalidCapabilityAndConfig(t *testing.T) {
	d := delivery(sendFunc(func(context.Context, mailer.Message) error { t.Fatal("sent rejected request"); return nil }))
	for _, body := range []string{`{}`, `{"name":"A","name":"B"}`, `{"name":"Visitor","email":"visitor@example.test","message":"Hello","to":"evil@example.test"}`} {
		if w := notify(d, body); w.Code != 422 {
			t.Fatal(w.Code)
		}
	}
	for _, value := range []string{"", "Owner <owner@example.test>", "a@example.test\r\nBcc: evil@example.test", "invalid"} {
		if (Email{To: value}).Validate() == nil {
			t.Fatal(value)
		}
	}
	if (Email{To: "owner@example.test", From: "invalid"}).Validate() == nil {
		t.Fatal("invalid from")
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
