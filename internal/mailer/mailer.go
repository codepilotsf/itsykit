// Package mailer keeps email delivery independent of authentication.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"
)

// From is an optional trusted configuration override, never a visitor field.
type Message struct{ To, Subject, Text, ReplyTo, HTML, From string }
type Sender interface {
	Send(context.Context, Message) error
}

// ValidAddress accepts one plain email address, without a display name.
func ValidAddress(value string) bool {
	a, err := mail.ParseAddress(value)
	return err == nil && a.Address == value && len(value) <= 254 && !strings.ContainsAny(value, "\r\n\x00")
}

type Settings struct {
	APIKey string `json:"api_key"`
	From   string `json:"from"`
	Name   string `json:"name,omitempty"`
}

func (s Settings) Validate() error {
	if len(s.Name) > 100 || strings.ContainsAny(s.Name, "\r\n\x00") {
		return fmt.Errorf("Please enter a sender name on one line, up to 100 characters.")
	}
	if !regexp.MustCompile(`^re_[A-Za-z0-9_-]{8,256}$`).MatchString(s.APIKey) {
		return fmt.Errorf("Please enter a valid Resend API key.")
	}
	a, err := mail.ParseAddress(s.From)
	if err != nil || a.Address != s.From || len(s.From) > 254 {
		return fmt.Errorf("Please enter a sender email address from your verified Resend domain.")
	}
	return nil
}

// Env is for the production process: systemd loads the root-only EnvironmentFile.
// Local development reads the site's .env through LocalSettings instead.
func Env() Sender {
	s := Settings{APIKey: os.Getenv("RESEND_API_KEY"), From: os.Getenv("FLEA_EMAIL_FROM"), Name: os.Getenv("FLEA_EMAIL_NAME")}
	if s.Validate() != nil {
		return nil
	}
	return NewResend(s)
}

type Resend struct {
	settings Settings
	client   *http.Client
}

func NewResend(s Settings) *Resend {
	return &Resend{s, &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (s *Resend) Send(ctx context.Context, m Message) error {
	from := s.settings.From
	if m.From != "" {
		if !ValidAddress(m.From) {
			return fmt.Errorf("The configured sender email address is invalid.")
		}
		from = m.From
	}
	if s.settings.Name != "" {
		from = (&mail.Address{Name: s.settings.Name, Address: from}).String()
	}
	data, err := json.Marshal(struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
		ReplyTo string   `json:"reply_to,omitempty"`
		HTML    string   `json:"html,omitempty"`
	}{from, []string{m.To}, m.Subject, m.Text, m.ReplyTo, m.HTML})
	if err != nil {
		return fmt.Errorf("Could not prepare the login email.")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Could not prepare the login email.")
	}
	req.Header.Set("Authorization", "Bearer "+s.settings.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	// Never return a provider body, request, or transport error: these may include
	// credentials, recipient addresses, or the one-time code.
	if err != nil {
		return fmt.Errorf("Email delivery is unavailable. Please try again later.")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Email delivery is unavailable (HTTP %d).", resp.StatusCode)
	}
	return nil
}
