package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestResendRequestAndSafeErrors(t *testing.T) {
	s := Settings{APIKey: "re_test123456", From: "login@example.test"}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, status := range []int{200, 401, 429, 500, 302} {
		sender := NewResend(s)
		sender.client.Transport = transport(func(r *http.Request) (*http.Response, error) {
			if r.URL.String() != "https://api.resend.com/emails" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer "+s.APIKey {
				t.Fatal("wrong request")
			}
			var data map[string]any
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Fatal(err)
			}
			if data["from"] != s.From || data["text"] != "123456" || data["html"] != "<b>123456</b>" || data["reply_to"] != "visitor@example.test" {
				t.Fatal(data)
			}
			return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("secret " + s.APIKey))}, nil
		})
		err := sender.Send(context.Background(), Message{To: "owner@example.test", Subject: "Login", Text: "123456", HTML: "<b>123456</b>", ReplyTo: "visitor@example.test"})
		if (err == nil) != (status == 200) {
			t.Fatal(status, err)
		}
		if err != nil && strings.Contains(err.Error(), s.APIKey) {
			t.Fatal("secret leaked")
		}
	}
}

func TestResendSenderName(t *testing.T) {
	s := NewResend(Settings{APIKey: "re_test123456", Name: "MySite Support", From: "hello@example.test"})
	s.client.Transport = transport(func(r *http.Request) (*http.Response, error) {
		var payload struct {
			From string `json:"from"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		address, err := mail.ParseAddress(payload.From)
		if err != nil || address.Name != "MySite Support" || address.Address != "hello@example.test" {
			t.Fatal(payload.From, err)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})
	if err := s.Send(context.Background(), Message{To: "owner@example.test", Subject: "Test", Text: "Hello"}); err != nil {
		t.Fatal(err)
	}
	if err := (Settings{APIKey: "re_test123456", From: "hello@example.test", Name: "Name\r\nBcc: stranger@example.test"}).Validate(); err == nil {
		t.Fatal("accepted header injection")
	}
}

func TestConfiguredSenderOverride(t *testing.T) {
	s := NewResend(Settings{APIKey: "re_test123456", From: "info@example.test", Name: "Support"})
	calls := 0
	s.client.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls++
		var payload struct {
			From string `json:"from"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		a, err := mail.ParseAddress(payload.From)
		if err != nil || a.Address != "support@example.test" || a.Name != "Support" {
			t.Fatal(payload.From, err)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	})
	if err := s.Send(context.Background(), Message{From: "support@example.test", To: "owner@example.test"}); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"bad", "Other <other@example.test>", "a@example.test\r\nBcc: other@example.test"} {
		if err := s.Send(context.Background(), Message{From: value}); err == nil {
			t.Fatal("invalid override accepted")
		}
	}
	if calls != 1 || s.settings.From != "info@example.test" {
		t.Fatal("default mutated or invalid request sent")
	}
}
