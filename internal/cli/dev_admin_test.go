package cli

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/localenv"
	"flea/internal/site"
)

func TestDevAdminReloadsAccessAndMail(t *testing.T) {
	for _, key := range []string{"RESEND_API_KEY", "FLEA_EMAIL_FROM", "FLEA_EMAIL_NAME"} {
		old, present := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if present {
				os.Setenv(key, old)
			} else {
				os.Unsetenv(key)
			}
		})
	}
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	handler := devAdmin(root, "Test", public, nil)
	get := func(want string) {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", "http://localhost/admin", nil))
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("want %q: %s", want, w.Body.String())
		}
	}
	get("Add the first site owner")
	if _, _, err := site.AddAdmin(root, "owner@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	get("Email is not configured")
	values := map[string]string{"RESEND_API_KEY": "re_test12345678", "FLEA_EMAIL_FROM": "hello@example.com"}
	if err := localenv.Save(root, values, []string{"RESEND_API_KEY", "FLEA_EMAIL_FROM"}); err != nil {
		t.Fatal(err)
	}
	get("Send sign-in code")
	if err := os.WriteFile(filepath.Join(root, "_", "flea.config"), []byte(`{"site":{"name":"Test"},"admin":{"emails":[]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := public.Reload(root); err != nil {
		t.Fatal(err)
	}
	get("Add the first site owner")
}
