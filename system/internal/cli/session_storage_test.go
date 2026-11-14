package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/site"
)

func TestLocalSessionFilesStayOutOfGit(t *testing.T) {
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err, string(out))
		}
		return string(out)
	}
	git("init", "-q")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("# Developer rule\ncache/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := prepareLocalSessions(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"itsy/runtime/sessions.json", "itsy/runtime/.sessions-lock", "itsy/runtime/.sessions-tmp"} {
		if !strings.Contains(git("check-ignore", "--", name), name) {
			t.Fatal("not ignored", name)
		}
	}
	data, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	if !strings.HasPrefix(string(data), "# Developer rule\ncache/\n") {
		t.Fatal("lost developer rules")
	}
	if err := prepareLocalSessions(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	again, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	if string(data) != string(again) {
		t.Fatal("ignore rules duplicated")
	}
	name := filepath.Join(root, "itsy", "runtime", "sessions.json")
	if err := os.WriteFile(name, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "-f", "--", "itsy/runtime/sessions.json")
	if err := prepareLocalSessions(context.Background(), p); err == nil || !strings.Contains(err.Error(), "tracked by Git") {
		t.Fatal("tracked session accepted", err)
	}
}
