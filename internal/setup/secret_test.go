package setup

import (
	"bytes"
	"context"
	"io"
	"itsy/internal/mailer"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskedSecret(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"re_secret\n", "re_secret"}, {"wrong\x15re_key\n", "re_key"}, {"re_keX\x7fy\n", "re_key"}, {"\n", ""}} {
		var out bytes.Buffer
		got, err := askMasked(context.Background(), strings.NewReader(tc.input), &out, "API key")
		if err != nil || got != tc.want {
			t.Fatal(got, err)
		}
		if strings.Contains(out.String(), "re_") || (tc.want != "" && !strings.Contains(out.String(), "*")) {
			t.Fatal("secret echo", out.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := askMasked(ctx, strings.NewReader("key\n"), io.Discard, "Key"); err == nil {
		t.Fatal("ignored cancel")
	}
}

func TestOptionalEmailSkip(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_NAME", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "_"), 0755)
	os.WriteFile(filepath.Join(root, ".itsy.config.json"), []byte(`{"app":{"site_name":"Test"}}`), 0644)
	var out bytes.Buffer
	err := configureEmailFlow(context.Background(), root, strings.NewReader("\n"), &out, func(mailer.Settings) mailer.Checker { t.Fatal("checked skipped key"); return nil }, true)
	if err != nil || strings.Contains(out.String(), "Sender") || strings.Contains(out.String(), "updated") {
		t.Fatal(out.String(), err)
	}
	if _, err := os.Stat(filepath.Join(root, ".env")); !os.IsNotExist(err) {
		t.Fatal("skip wrote .env")
	}
}
