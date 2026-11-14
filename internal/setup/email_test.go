package setup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/localenv"
	"itsy/internal/mailer"
)

type connectionCheck func(context.Context) (string, error)

func (f connectionCheck) Check(ctx context.Context) (string, error) { return f(ctx) }

func TestEmailWizard(t *testing.T) {
	for _, fail := range []bool{false, true} {
		dir := t.TempDir()
		os.Mkdir(filepath.Join(dir, "_"), 0755)
		os.WriteFile(filepath.Join(dir, "itsy.config.json"), []byte(`{"app":{"site_name":"Test"}}`), 0644)
		var out bytes.Buffer
		err := configureEmail(context.Background(), dir, strings.NewReader("re_test123456\nMySite Support\nhello@example.test\n"), &out, func(s mailer.Settings) mailer.Checker {
			if s.Name != "MySite Support" || s.From != "hello@example.test" {
				t.Fatal(s.Name, s.From)
			}
			return connectionCheck(func(context.Context) (string, error) {
				if fail {
					return "", fmt.Errorf("connection failed")
				}
				return "Connected to Resend.", nil
			})
		})
		if (err != nil) != fail {
			t.Fatal(err)
		}
		if strings.Contains(out.String(), "re_test123456") {
			t.Fatal("key leaked")
		}
		if strings.Contains(out.String(), "Connected to Resend") || strings.Contains(out.String(), "not been tested") {
			t.Fatal("unnecessary connection notes", out.String())
		}
		values, readErr := localenv.Read(dir)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if fail && len(values) != 0 {
			t.Fatal("saved despite failed check")
		}
		if !fail && (values["FLEA_EMAIL_NAME"] != "MySite Support" || !strings.Contains(out.String(), ".env updated.")) {
			t.Fatal(out.String())
		}
	}
}

func TestEmailRefusesTrackedEnv(t *testing.T) {
	dir := repoSite(t)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("OTHER=value\n"), 0600)
	cmd := exec.Command("git", "-C", dir, "add", "-f", ".env")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	err := configureEmail(context.Background(), dir, strings.NewReader(""), &bytes.Buffer{}, func(mailer.Settings) mailer.Checker { t.Fatal("checked before rejecting tracked .env"); return nil })
	if err == nil || !strings.Contains(err.Error(), "tracked by Git") {
		t.Fatal(err)
	}
}
