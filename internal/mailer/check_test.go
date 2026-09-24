package mailer

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConnectionCheck(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		ok     bool
	}{
		{200, `{"data":[]}`, true},
		{401, `{"name":"restricted_api_key"}`, true},
		{403, `{"name":"restricted_api_key"}`, false},
		{401, `{"name":"invalid_api_key","message":"secret"}`, false},
		{429, `{}`, false}, {500, `{}`, false}, {302, `{}`, false},
	} {
		s := NewResend(Settings{APIKey: "re_test123456", From: "hello@example.test"})
		s.client.Transport = transport(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/domains" {
				t.Fatal("check attempted to send")
			}
			return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		_, err := s.Check(context.Background())
		if (err == nil) != tc.ok {
			t.Fatal(tc, err)
		}
		if err != nil && strings.Contains(err.Error(), "secret") {
			t.Fatal("provider body leaked")
		}
	}
}

func TestLocalEnvironmentPrecedence(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("RESEND_API_KEY=re_file123456\nFLEA_EMAIL_NAME=MySite Support\nFLEA_EMAIL_FROM=hello@example.test\n"), 0600)
	t.Setenv("RESEND_API_KEY", "re_shell123456")
	t.Setenv("FLEA_EMAIL_FROM", "")
	s, err := LocalSettings(dir)
	if err != nil || s.APIKey != "re_shell123456" || s.From != "" {
		t.Fatal(err, s.From)
	}
}
