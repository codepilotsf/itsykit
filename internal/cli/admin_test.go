package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestAdminCommand(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "re_test12345678")
	t.Setenv("FLEA_EMAIL_FROM", "hello@example.com")
	t.Chdir(t.TempDir())
	if err := Create("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"admin", "owner@example.com"}, &out, &out); code != 0 || !strings.Contains(out.String(), "Added owner@example.com") {
		t.Fatal(code, out.String())
	}
	for _, text := range []string{"Sign in at /admin", "updates automatically", "itsy deploy"} {
		if !strings.Contains(out.String(), text) {
			t.Fatalf("missing login setup instruction %q", text)
		}
	}
	if strings.Contains(out.String(), "Email setup") || strings.Count(out.String(), "\n") > 6 {
		t.Fatal("configured site must have concise output", out.String())
	}
	out.Reset()
	if code := Run(context.Background(), []string{"admin", "OWNER@example.com"}, &out, &out); code != 0 || !strings.Contains(out.String(), "already has admin access") {
		t.Fatal(code, out.String())
	}
	for _, args := range [][]string{{"admin", "--bad"}, {"admin", "a", "b"}} {
		if code := Run(context.Background(), args, io.Discard, io.Discard); code != 2 {
			t.Fatal(args, code)
		}
	}
	if code := Run(context.Background(), []string{"admin", "bad"}, io.Discard, io.Discard); code != 1 {
		t.Fatal(code)
	}
	if code := Run(context.Background(), []string{"admin", "--help"}, io.Discard, io.Discard); code != 0 {
		t.Fatal(code)
	}
	input, err := os.CreateTemp(t.TempDir(), "input")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if _, err := input.WriteString("second@example.com\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	previous := os.Stdin
	os.Stdin = input
	defer func() { os.Stdin = previous }()
	out.Reset()
	if code := Run(context.Background(), []string{"admin"}, &out, &out); code != 0 || !strings.Contains(out.String(), "Admin email address:") || !strings.Contains(out.String(), "Added second@example.com") {
		t.Fatal(code, out.String())
	}
}

func TestAdminStartsMissingEmailSetup(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	t.Chdir(t.TempDir())
	if err := Create("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	config, err := os.ReadFile(".flea.config.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".flea.config.json", bytes.Replace(config, []byte(`"email_preview": true`), []byte(`"email_preview": false`), 1), 0644); err != nil {
		t.Fatal(err)
	}
	input, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	previous := os.Stdin
	os.Stdin = input
	defer func() { os.Stdin = previous }()
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"admin", "owner@example.com"}, &out, &out); code != 1 {
		t.Fatal("missing setup input must stop without sending", code)
	}
	for _, want := range []string{"Added owner@example.com", "Email needs setting up"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal(out.String())
		}
	}
}

func TestAdminWithPreviewNeedsNoCredentials(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	t.Chdir(t.TempDir())
	if err := Create("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"admin", "owner@example.com"}, &out, &out); code != 0 {
		t.Fatal(code, out.String())
	}
	if strings.Contains(out.String(), "Email needs") || strings.Contains(out.String(), "Resend") {
		t.Fatal(out.String())
	}
}
