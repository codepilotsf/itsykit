package site

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"itsy/internal/appbundle"
	"itsy/internal/contact"
	"itsy/internal/mailer"
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
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/deep/contact","email":{"to":"owner@example.test","from":"support@example.test"}}}}`)
	put(t, root, "deep/contact/index.html", `<form method="post" action="/deep/contact">{csrf}<input name="name" value="{contact.data.name}"><input name="email" value="{contact.data.email}"><textarea name="message">{contact.data.message}</textarea><p>{contact.invalid.email}</p><p>{contact.error}</p></form>`)
	put(t, root, "deep/contact/thank-you.html", `<h1>Thanks</h1><p>{contact.flash.notice}</p>`)
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
	if w.Code != 303 || w.Header().Get("Location") != "/deep/contact/thank-you" || calls.Load() != 1 {
		t.Fatal(w.Code, w.Body.String())
	}
	flash := w.Result().Cookies()
	if len(flash) != 1 {
		t.Fatal("no flash cookie", w.Header())
	}
	if get("/about", cookie, flash[0]).Code != 200 {
		t.Fatal("ordinary page failed")
	}
	w = get("/deep/contact/thank-you", cookie, flash[0])
	if !strings.Contains(w.Body.String(), "Your message was sent.") {
		t.Fatal("missing flash", w.Body.String())
	}
	if strings.Contains(get("/deep/contact/thank-you", cookie, flash[0]).Body.String(), "Your message was sent.") {
		t.Fatal("flash replayed")
	}
	if get("/deep/contact/thanks").Code != 404 {
		t.Fatal("superseded success route must not be an alias")
	}
	fail.Store(true)
	w = post("v@example.test")
	if w.Code != 503 || strings.Contains(w.Body.String(), "provider secret") || !strings.Contains(w.Body.String(), "Hello") {
		t.Fatal(w.Code, w.Body.String())
	}
	before := calls.Load()
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/deep/contact","email":{"to":"changed@example.test"}}}}`)
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

// managedContactRoot builds a minimal site whose Contact app is enabled at
// /deep/contact with the mailers Site runtime expects. It mirrors the fixture
// used by TestManagedContact without changing it.
func managedContactRoot(t *testing.T) string {
	t.Helper()
	root := fixture(t)
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/deep/contact","email":{"to":"owner@example.test","from":"support@example.test"}}}}`)
	put(t, root, "deep/contact/index.html", `<form method="post" action="/deep/contact">{csrf}<input name="name" value="{contact.data.name}"><input name="email" value="{contact.data.email}"><textarea name="message">{contact.data.message}</textarea><p>{contact.invalid.email}</p><p>{contact.error}</p></form>`)
	put(t, root, "deep/contact/thank-you.html", `<h1>Thanks</h1><p>{contact.flash.notice}</p>`)
	put(t, root, "mailers/main.layout.html", `<html>{slot}</html>`)
	for _, kind := range []string{"sign-in", "contact"} {
		put(t, root, "mailers/"+kind+".email.html", `<p>{email.message}</p>`)
		put(t, root, "mailers/"+kind+".email.txt", `{email.message}`)
	}
	return root
}

// waitForContact returns the running app once the managed child has handshaked.
func waitForContact(t *testing.T, s *Server, p *Project) *runningApp {
	t.Helper()
	_, running, ok := s.appFor(p, "/deep/contact")
	if !ok || running == nil || running.manager == nil {
		t.Fatal("Contact was not registered as a running app")
	}
	deadline := time.Now().Add(8 * time.Second)
	for {
		if running.manager.Client() != nil {
			return running
		}
		if time.Now().After(deadline) {
			t.Fatal("Contact did not become ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// contactForm fetches the mount and returns its anonymous form cookie and CSRF token.
func contactForm(t *testing.T, s *Server) (*http.Cookie, string) {
	t.Helper()
	r := httptest.NewRequest("GET", "https://example.test/deep/contact", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("contact form: %d %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("contact form cookies = %v", cookies)
	}
	token := formToken(w.Body.String())
	if token == "" {
		t.Fatal("contact form missing CSRF token")
	}
	return cookies[0], token
}

func contactSubmit(ctx context.Context, s *Server, cookie *http.Cookie, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequestWithContext(ctx, "POST", "https://example.test/deep/contact", strings.NewReader(url.Values{"_csrf": {token}, "name": {"Visitor"}, "email": {"v@example.test"}, "message": {"Hello"}}.Encode()))
	r.Header.Set("Origin", "https://example.test")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

// TestManagedContactShutdownDuringSubmission keeps a real Contact child busy in
// an active delivery, then shuts the app down the way dev/host does. The blocked
// admission must complete within a bound, the child must be reaped, and the same
// submission must never be attempted a second time.
func TestManagedContactShutdownDuringSubmission(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing Contact bundle")
		}
		t.Skip("run scripts/build.sh for the managed-process integration test")
	}
	root := managedContactRoot(t)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	var calls atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	ctx, cancel := context.WithCancel(context.Background())
	stop := s.StartApps(ctx, appSender(func(sendCtx context.Context, m mailer.Message) error {
		if calls.Add(1) == 1 {
			close(entered)
		}
		select {
		case <-release:
			return nil
		case <-sendCtx.Done():
			return sendCtx.Err()
		}
	}), &contact.Budget{})
	defer func() {
		unblock()
		cancel()
		stop()
	}()
	running := waitForContact(t, s, p)
	cookie, token := contactForm(t, s)
	submitCtx, submitCancel := context.WithCancel(context.Background())
	defer submitCancel()
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() { result <- contactSubmit(submitCtx, s, cookie, token) }()
	select {
	case <-entered:
	case <-time.After(8 * time.Second):
		t.Fatal("submission never reached the email sender")
	}
	cancel()
	stopped := make(chan struct{})
	go func() { stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(6 * time.Second):
		t.Fatal("shutdown exceeded six seconds")
	}
	select {
	case w := <-result:
		if w.Code != 503 {
			t.Fatalf("submission during shutdown: %d %s", w.Code, w.Body.String())
		}
	case <-time.After(8 * time.Second):
		t.Fatal("submission did not complete within a bounded time")
	}
	unblock()
	time.Sleep(200 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("delivery attempts = %d, want exactly 1", got)
	}
	if running.manager.Client() != nil {
		t.Fatal("app client still published after shutdown")
	}
}

// TestOrdinaryPagesUsableWhenContactUnavailable takes Contact down after it has
// run, leaving a manager with no published client, and checks the rest of the
// site keeps serving while the mount renders its form but fails submissions closed.
func TestOrdinaryPagesUsableWhenContactUnavailable(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing Contact bundle")
		}
		t.Skip("run scripts/build.sh for the managed-process integration test")
	}
	root := managedContactRoot(t)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	stop := s.StartApps(ctx, appSender(func(context.Context, mailer.Message) error {
		calls.Add(1)
		return nil
	}), &contact.Budget{})
	defer stop()
	running := waitForContact(t, s, p)
	// Stop the child but keep the site running with the app still registered.
	cancel()
	deadline := time.Now().Add(8 * time.Second)
	for running.manager.Client() != nil {
		if time.Now().After(deadline) {
			t.Fatal("Contact client stayed published after shutdown")
		}
		time.Sleep(20 * time.Millisecond)
	}
	for _, route := range []string{"/", "/about", "/css/main.css"} {
		if w := request(s, "GET", "https://example.test"+route); w.Code != 200 {
			t.Fatalf("%s while Contact unavailable: %d", route, w.Code)
		}
	}
	cookie, token := contactForm(t, s)
	if w := contactSubmit(context.Background(), s, cookie, token); w.Code != 503 {
		t.Fatalf("submission while Contact unavailable: %d %s", w.Code, w.Body.String())
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("delivery attempted %d times without a running app", got)
	}
}
