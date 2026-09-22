package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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
			if data["from"] != s.From || data["text"] != "123456" {
				t.Fatal(data)
			}
			return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("secret " + s.APIKey))}, nil
		})
		err := sender.Send(context.Background(), Message{To: "owner@example.test", Subject: "Login", Text: "123456"})
		if (err == nil) != (status == 200) {
			t.Fatal(status, err)
		}
		if err != nil && strings.Contains(err.Error(), s.APIKey) {
			t.Fatal("secret leaked")
		}
	}
}
