package localenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveRoundTripPreservesOtherSettings(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("# keep this\nOTHER='$(do-not-run)'\nRESEND_API_KEY=old\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("!.env\n"), 0644)
	values := map[string]string{"RESEND_API_KEY": "re_test123456", "FLEA_EMAIL_NAME": "My \"Site\" $Support", "FLEA_EMAIL_FROM": "hello@example.test"}
	keys := []string{"RESEND_API_KEY", "FLEA_EMAIL_NAME", "FLEA_EMAIL_FROM"}
	for i := 0; i < 2; i++ {
		if err := Save(dir, values, keys); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range values {
		if got[key] != value {
			t.Fatal(key, got[key])
		}
	}
	if got["OTHER"] != "$(do-not-run)" {
		t.Fatal(got)
	}
	info, _ := os.Stat(filepath.Join(dir, ".env"))
	if info.Mode().Perm() != 0600 {
		t.Fatal(info.Mode())
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if strings.Count(string(data), "\n/.env\n") != 1 {
		t.Fatal(string(data))
	}
}

func TestRefusesLinksAndMalformedFiles(t *testing.T) {
	for _, name := range []string{".env", ".gitignore"} {
		dir := t.TempDir()
		outside := filepath.Join(t.TempDir(), "file")
		os.WriteFile(outside, []byte("KEEP=value\n"), 0600)
		os.Symlink(outside, filepath.Join(dir, name))
		if err := Save(dir, map[string]string{"KEY": "value"}, []string{"KEY"}); err == nil {
			t.Fatal("accepted link", name)
		}
	}
	if _, err := parse([]byte("KEY=\"unterminated")); err == nil {
		t.Fatal("accepted malformed value")
	}
}
