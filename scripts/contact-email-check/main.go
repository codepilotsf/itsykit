// Opt-in real-delivery check. Never run without approval for the exact addresses.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"flea/internal/contact"
	"flea/internal/mailer"
	"flea/internal/site"
	"flea/templates"
)

type checkedSender struct {
	next     mailer.Sender
	from, to string
	calls    atomic.Int32
}

func (s *checkedSender) Send(ctx context.Context, m mailer.Message) error {
	if m.From != s.from || m.To != s.to || m.ReplyTo != s.to || !strings.Contains(m.Text, "Flea Contact app delivery test") || !strings.Contains(m.HTML, "Flea Contact app delivery test") {
		return fmt.Errorf("unexpected test email; nothing sent")
	}
	if s.calls.Add(1) != 1 {
		return fmt.Errorf("only one email is allowed")
	}
	return s.next.Send(ctx, m)
}
func main() {
	from := flag.String("from", "", "explicitly approved sender override")
	to := flag.String("to", "", "explicitly approved recipient")
	send := flag.Bool("send", false, "authorize exactly one real test email")
	flag.Parse()
	if !*send || !mailer.ValidAddress(*from) || !mailer.ValidAddress(*to) {
		fmt.Fprintln(os.Stderr, "Explicit -send, -from and -to are required.")
		os.Exit(1)
	}
	if err := run(*from, *to); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(from, to string) error {
	real := mailer.Env()
	if real == nil {
		return fmt.Errorf("valid server email settings are required")
	}
	root, err := os.MkdirTemp("", "flea-email-site-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	write := func(name string, data []byte) error {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return err
		}
		return os.WriteFile(p, data, 0600)
	}
	cfg, _ := json.Marshal(map[string]any{"app": map[string]string{"site_name": "Flea Contact delivery test"}, "apps": map[string]any{"contact": map[string]any{"mount": "/contact", "email": map[string]string{"to": to, "from": from}}}})
	if err = write(".flea.config.json", cfg); err != nil {
		return err
	}
	for name, body := range map[string]string{"site/contact/index.html": `<form method="post" action="/contact">{csrf}</form><p>{contact.invalid.email}</p><p>{contact.error}</p>`, "site/contact/thanks.html": `<h1>Thanks</h1><p>{contact.flash.notice}</p>`} {
		if err = write(name, []byte(body)); err != nil {
			return err
		}
	}
	for _, name := range []string{"main.layout.html", "sign-in.email.html", "sign-in.email.txt", "contact.email.html", "contact.email.txt"} {
		b, err := templates.Files.ReadFile("mailers/" + name)
		if err != nil {
			return err
		}
		if err = write("mailers/"+name, b); err != nil {
			return err
		}
	}
	project, err := site.Load(root)
	if err != nil {
		return err
	}
	public := site.NewProductionServer(project, nil)
	sender := &checkedSender{next: real, from: from, to: to}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	stop := public.StartApps(ctx, sender, &contact.Budget{})
	defer stop()
	server := httptest.NewTLSServer(public)
	defer server.Close()
	client := server.Client()
	client.Timeout = 20 * time.Second
	defer client.CloseIdleConnections()
	client.Jar, _ = cookiejar.New(nil)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Get(server.URL + "/contact")
	if err != nil {
		return fmt.Errorf("test form unavailable")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if err != nil || response.StatusCode != 200 {
		return fmt.Errorf("test form failed")
	}
	match := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindSubmatch(body)
	if len(match) != 2 {
		return fmt.Errorf("missing form protection")
	}
	post := func(email, message string) (*http.Response, error) {
		fields := url.Values{"_csrf": {string(match[1])}, "name": {"Flea delivery test"}, "email": {email}, "message": {message}}
		r, err := http.NewRequestWithContext(ctx, "POST", server.URL+"/contact", strings.NewReader(fields.Encode()))
		if err != nil {
			return nil, err
		}
		r.Header.Set("Origin", server.URL)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return client.Do(r)
	}
	// Readiness retries contain invalid data and cannot send an email.
	deadline := time.Now().Add(8 * time.Second)
	for {
		response, err = post("invalid", "Readiness only")
		if err == nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if response.StatusCode == 422 {
				break
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Contact readiness failed; no valid submission attempted")
		}
		time.Sleep(50 * time.Millisecond)
	}
	message := "Flea Contact app delivery test\n\nThis single test was requested by Sean. It verifies the app-level sender override, Reply-To, and HTML/plain-text email rendering. No reply is required.\n\nTest time: " + time.Now().UTC().Format(time.RFC3339)
	// Exactly one valid submission. Never retry a timeout or uncertain failure.
	response, err = post(to, message)
	if err != nil {
		return fmt.Errorf("delivery outcome uncertain; do not retry automatically")
	}
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != 303 || response.Header.Get("Location") != "/contact/thanks" || sender.calls.Load() != 1 {
		return fmt.Errorf("Contact did not confirm acceptance (HTTP %d); no automatic retry", response.StatusCode)
	}
	response, err = client.Get(server.URL + "/contact/thanks")
	if err != nil {
		return fmt.Errorf("email accepted; thank-you request failed")
	}
	body, err = io.ReadAll(io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if err != nil || response.StatusCode != 200 || !strings.Contains(string(body), "Your message was sent.") {
		return fmt.Errorf("email accepted; flash check failed")
	}
	fmt.Println("One email accepted by Resend; Contact returned 303; thank-you flash rendered. Inbox receipt is not yet verified.")
	return nil
}
