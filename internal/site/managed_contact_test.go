package site

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"flea/internal/appbundle"
	"flea/internal/contact"
	"flea/internal/mailer"
)

type appSender func(context.Context, mailer.Message) error

func (f appSender) Send(ctx context.Context, m mailer.Message) error { return f(ctx, m) }

func TestManagedContact(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing Contact bundle")
		}
		t.Skip("run scripts/build.sh for the managed-process integration test")
	}
	root := fixture(t)
	put(t, root, ".flea.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/deep/contact","email":{"to":"owner@example.test","from":"support@example.test"}}}}`)
	put(t, root, "deep/contact/index.html", `<form method="post" action="/deep/contact">{csrf}<input name="name" value="{contact.data.name}"><input name="email" value="{contact.data.email}"><textarea name="message">{contact.data.message}</textarea><p>{contact.invalid.email}</p><p>{contact.error}</p></form>`)
	put(t, root, "deep/contact/thanks.html", `<h1>Thanks</h1><p>{contact.flash.notice}</p>`)
	put(t, root, "mailers/main.layout.html", `<html>{slot}</html>`)
	for _, kind := range []string{"sign-in", "contact"} {
		put(t, root, "mailers/"+kind+".email.html", `<p>{email.message}</p>`)
		put(t, root, "mailers/"+kind+".email.txt", `{email.message}`)
	}
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	var calls atomic.Int32
	var fail atomic.Bool
	stop := s.StartApps(context.Background(), appSender(func(ctx context.Context, m mailer.Message) error {
		calls.Add(1)
		if m.To != "owner@example.test" || m.From != "support@example.test" || m.ReplyTo != "v@example.test" || m.Text != "Hello" {
			t.Error("wrong notification", m)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Error("missing deadline")
		}
		if fail.Load() {
			return errors.New("provider secret")
		}
		return nil
	}), &contact.Budget{})
	defer stop()
	deadline := time.Now().Add(8 * time.Second)
	for {
		_, running, _ := s.appFor(p, "/deep/contact")
		if running != nil && running.manager != nil && running.manager.Client() != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Contact did not become ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
	get := func(route string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "https://example.test"+route, nil)
		for _, c := range cookies {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	w := get("/deep/contact")
	if w.Code != 200 || len(w.Result().Cookies()) != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	token := formToken(w.Body.String())
	post := func(email string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "https://example.test/deep/contact", strings.NewReader(url.Values{"_csrf": {token}, "name": {"Visitor"}, "email": {email}, "message": {"Hello"}}.Encode()))
		r.Header.Set("Origin", "https://example.test")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	w = post("invalid")
	if w.Code != 422 || calls.Load() != 0 || !strings.Contains(w.Body.String(), "valid email") || !strings.Contains(w.Body.String(), "Hello") {
		t.Fatal(w.Code, w.Body.String())
	}
	w = post("v@example.test")
	if w.Code != 303 || w.Header().Get("Location") != "/deep/contact/thanks" || calls.Load() != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	flash := w.Result().Cookies()
	if len(flash) != 1 {
		t.Fatal("no flash cookie", w.Header())
	}
	if get("/about", cookie, flash[0]).Code != 200 {
		t.Fatal("ordinary page failed")
	}
	w = get("/deep/contact/thanks", cookie, flash[0])
	if !strings.Contains(w.Body.String(), "Your message was sent.") {
		t.Fatal("missing flash", w.Body.String())
	}
	if strings.Contains(get("/deep/contact/thanks", cookie, flash[0]).Body.String(), "Your message was sent.") {
		t.Fatal("flash replayed")
	}
	fail.Store(true)
	w = post("v@example.test")
	if w.Code != 503 || strings.Contains(w.Body.String(), "provider secret") || !strings.Contains(w.Body.String(), "Hello") {
		t.Fatal(w.Code, w.Body.String())
	}
	before := calls.Load()
	put(t, root, ".flea.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/deep/contact","email":{"to":"changed@example.test"}}}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	w = post("v@example.test")
	if w.Code != 503 || calls.Load() != before {
		t.Fatal("used stale app config", w.Code)
	}
	if get("/about").Code != 200 {
		t.Fatal("config change broke ordinary page")
	}
}
