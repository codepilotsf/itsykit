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
	t.Chdir(t.TempDir())
	if err := Create("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"admin", "owner@example.com"}, &out, &out); code != 0 || !strings.Contains(out.String(), "Added owner@example.com") {
		t.Fatal(code, out.String())
	}
	for _, text := range []string{"RESEND_API_KEY", "NANOWEB_EMAIL_FROM", "domain verified in Resend", "Keep .env out of Git", "source .env", "http://127.0.0.1:3000/admin"} {
		if !strings.Contains(out.String(), text) {
			t.Fatalf("missing login setup instruction %q", text)
		}
	}
	out.Reset()
	if code := Run(context.Background(), []string{"admin", "OWNER@example.com"}, &out, &out); code != 0 || !strings.Contains(out.String(), "already an admin") {
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
