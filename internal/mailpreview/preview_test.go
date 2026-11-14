package mailpreview

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"itsy/internal/devaccess"
	"itsy/internal/mailer"
)

func fixture(t *testing.T) (*Preview, *string) {
	t.Helper()
	p, err := New("http://127.0.0.1:3000", "Acme & Co", "hello@example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	address := new(string)
	p.open = func(_ context.Context, s string) error { *address = s; return nil }
	p.now = func() time.Time { return time.Date(2026, 9, 28, 14, 35, 0, 0, time.UTC) }
	return p, address
}

func get(p *Preview, address string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", address, nil)
	r.RemoteAddr = "127.0.0.1:43210"
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)
	return w
}

func TestPreviewMessageAndIsolation(t *testing.T) {
	p, address := fixture(t)
	m := mailer.Message{To: "owner@example.test", Subject: "Your sign-in code <script>alert(1)</script>", HTML: `<h1>123456</h1><script>parent.document.body.remove()</script>`, Text: "Your code is 123456", ReplyTo: "reply@example.test"}
	if err := p.Send(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	w := get(p, *address)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, text := range []string{"Acme &amp; Co", "hello@example.test", "owner@example.test", "Reply-To", "September 28, 2026 at 2:35 PM UTC", "sandbox srcdoc=", "Plain-text version", "Not sent", "123456"} {
		if !strings.Contains(w.Body.String(), text) {
			t.Errorf("missing %q", text)
		}
	}
	if strings.Contains(w.Body.String(), "<script>") || !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("unsafe preview")
	}
	for _, tc := range []struct {
		name, host, peer, origin, fetch, method string
		status                                  int
	}{
		{name: "remote peer", peer: "192.0.2.1:123", status: 404},
		{name: "rebinding host", host: "evil.example:3000", status: 404},
		{name: "wrong port", host: "127.0.0.1:4000", status: 404},
		{name: "cross origin", origin: "https://evil.example", status: 404},
		{name: "cross site", fetch: "cross-site", status: 404},
		{name: "post", method: "POST", status: 405},
		{name: "head", method: "HEAD", status: 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", *address, nil)
			r.RemoteAddr = "127.0.0.1:123"
			if tc.host != "" {
				r.Host = tc.host
			}
			if tc.peer != "" {
				r.RemoteAddr = tc.peer
			}
			if tc.method != "" {
				r.Method = tc.method
			}
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Sec-Fetch-Site", tc.fetch)
			w := httptest.NewRecorder()
			p.ServeHTTP(w, r)
			if w.Code != tc.status || strings.Contains(w.Body.String(), "123456") {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	if get(p, p.origin+Path).Code != 404 || get(p, p.origin+Path+strings.Repeat("0", 64)).Code != 404 {
		t.Fatal("preview enumeration")
	}
}

func TestPreviewSenderOverride(t *testing.T) {
	p, address := fixture(t)
	if err := p.Send(context.Background(), mailer.Message{To: "owner@example.test", From: "support@example.test", Text: "Hello"}); err != nil {
		t.Fatal(err)
	}
	w := get(p, *address)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "support@example.test") || strings.Contains(w.Body.String(), "hello@example.test") {
		t.Fatal("override missing", w.Body.String())
	}
	if err := p.Send(context.Background(), mailer.Message{To: "owner@example.test", From: "invalid", Text: "Hello"}); err == nil {
		t.Fatal("invalid sender accepted")
	}
	if err := p.Send(context.Background(), mailer.Message{To: "owner@example.test", Text: "Default"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(get(p, *address).Body.String(), "hello@example.test") {
		t.Fatal("override changed default")
	}
}

func TestLANPreviewRequiresCapabilityAndApprovedRequest(t *testing.T) {
	p, address := fixture(t)
	if err := p.Send(context.Background(), mailer.Message{To: "owner@example.test", Text: "Private LAN preview"}); err != nil {
		t.Fatal(err)
	}
	h := devaccess.WrapLAN(p, []string{"192.168.1.5"}, []string{"asahi", "asahi.local"}, 3000)
	lanURL := strings.Replace(*address, "127.0.0.1", "192.168.1.5", 1)
	for _, tc := range []struct {
		name, url, origin, peer string
		handler                 http.Handler
		want                    int
	}{
		{"hostname", strings.Replace(lanURL, "192.168.1.5", "asahi", 1), "http://asahi:3000", "100.79.67.33:4321", h, 200},
		{"LAN", lanURL, "", "192.168.1.9:4321", h, 200},
		{"same origin", lanURL, "http://192.168.1.5:3000", "192.168.1.9:4321", h, 200},
		{"default", lanURL, "", "192.168.1.9:4321", p, 404},
		{"foreign origin", lanURL, "http://evil.test", "192.168.1.9:4321", h, 404},
		{"public peer", lanURL, "", "203.0.113.8:4321", h, 403},
		{"no token", "http://192.168.1.5:3000" + Path, "", "192.168.1.9:4321", h, 404},
		{"unknown token", "http://192.168.1.5:3000" + Path + strings.Repeat("0", 64), "", "192.168.1.9:4321", h, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tc.url, nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()
			tc.handler.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatal(w.Code, w.Body.String())
			}
			if tc.want != 200 && strings.Contains(w.Body.String(), "Private LAN preview") {
				t.Fatal("preview leaked")
			}
		})
	}
}

func TestPreviewRetentionAndBrowserFailure(t *testing.T) {
	p, address := fixture(t)
	var output bytes.Buffer
	p.logger = log.New(&output, "", 0)
	p.open = func(_ context.Context, s string) error { *address = s; return errors.New("sensitive process output") }
	if err := p.Send(context.Background(), mailer.Message{Subject: "Private subject", Text: "123456"}); err != nil {
		t.Fatal(err)
	}
	first := *address
	if !strings.Contains(output.String(), first) || strings.Contains(output.String(), "123456") || strings.Contains(output.String(), "sensitive") {
		t.Fatal("unsafe fallback log")
	}
	for i := 0; i < maxMessages; i++ {
		if err := p.Send(context.Background(), mailer.Message{Text: "More mail"}); err != nil {
			t.Fatal(err)
		}
	}
	if len(p.messages) != maxMessages || get(p, first).Code != 404 {
		t.Fatal("unbounded retention")
	}
	last := *address
	p.now = func() time.Time { return time.Date(2026, 9, 28, 14, 46, 0, 0, time.UTC) }
	if get(p, last).Code != 404 || len(p.messages) != 0 {
		t.Fatal("expired message retained")
	}
	if err := p.Send(context.Background(), mailer.Message{HTML: strings.Repeat("x", (256<<10)+1)}); err == nil {
		t.Fatal("unbounded HTML")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Send(ctx, mailer.Message{}); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestPreviewConcurrentSends(t *testing.T) {
	p, _ := fixture(t)
	p.open = func(context.Context, string) error { return nil }
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Go(func() {
			if err := p.Send(context.Background(), mailer.Message{Text: "Test"}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(p.messages) != maxMessages {
		t.Fatal(len(p.messages))
	}
}

func TestPreviewRequiresLoopback(t *testing.T) {
	for _, origin := range []string{"https://127.0.0.1:3000", "http://example.com:3000", "http://0.0.0.0:3000", "http://127.0.0.1", "http://127.0.0.1:3000/path"} {
		if _, err := New(origin, "", "", nil); err == nil {
			t.Fatal(origin)
		}
	}
}

// Opt in to real browser rendering without launching a user's regular profile.
func TestPreviewBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	s := httptest.NewUnstartedServer(nil)
	p, err := New("http://"+s.Listener.Addr().String(), "Juniper Studio", "hello@juniper.example", nil)
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	s.Config.Handler = p
	s.Start()
	defer s.Close()
	address := ""
	p.open = func(_ context.Context, s string) error { address = s; return nil }
	err = p.Send(context.Background(), mailer.Message{To: "sean@example.test", Subject: "Your sign-in code for Juniper Studio", Text: "Your sign-in code is 482916", HTML: `<html><body style="margin:0;padding:48px 32px;font:16px/1.7 system-ui;color:#26332d"><p style="margin:0 0 16px;font-size:12px;letter-spacing:.14em;color:#66776c">JUNIPER STUDIO</p><h1 style="font-size:25px;font-weight:600">Welcome back.</h1><p>Use this code to sign in to your website.</p><p style="font-size:38px;letter-spacing:.2em;font-weight:600">482916</p><p style="color:#66776c;font-size:14px">This code expires in 10 minutes.</p><script>parent.document.title="UNSAFE"</script></body></html>`})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"--headless", "--disable-gpu", "--disable-background-networking", "--disable-component-update", "--no-first-run", "--no-default-browser-check", "--user-data-dir=" + t.TempDir(), "--window-size=1100,1000", "--dump-dom", "--virtual-time-budget=1500"}
	if size := os.Getenv("FLEA_PREVIEW_WINDOW_SIZE"); size != "" {
		args = append(args, "--window-size="+size)
	}
	if screenshot := os.Getenv("FLEA_PREVIEW_SCREENSHOT"); screenshot != "" {
		args = append(args, "--screenshot="+filepath.Clean(screenshot))
	}
	cmd := exec.CommandContext(ctx, browser, append(args, address)...)
	cmd.WaitDelay = time.Second
	result, err := cmd.Output()
	if !bytes.Contains(result, []byte("Email message")) || bytes.Contains(result, []byte("<title>UNSAFE")) {
		t.Fatal("browser preview failed", err)
	}
}
