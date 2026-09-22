package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMailConfigurationIsPrivateAndNotInServerJSON(t *testing.T) {
	m, _ := fixture(t)
	before, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if MailReady(m.ConfigPath) {
		t.Fatal("unexpected settings")
	}
	if err := m.ConfigureMail(strings.NewReader(`{"api_key":"re_test123456","from":"login@example.test"}`)); err != nil {
		t.Fatal(err)
	}
	if !MailReady(m.ConfigPath) {
		t.Fatal("settings not ready")
	}
	info, err := os.Stat(filepath.Join(filepath.Dir(m.ConfigPath), "mail.env"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal(info, err)
	}
	after, err := os.ReadFile(m.ConfigPath)
	if err != nil || string(before) != string(after) {
		t.Fatal("host configuration changed", err)
	}
	for _, input := range []string{`{"api_key":"secret\ninjected","from":"a@example.test"}`, `{"api_key":"re_test123456","from":"bad"}`, `{} {}`} {
		if err := m.ConfigureMail(strings.NewReader(input)); err == nil {
			t.Fatal("invalid mail settings accepted")
		}
	}
	if !MailReady(m.ConfigPath) {
		t.Fatal("invalid input replaced valid settings")
	}
}
