package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewInitializesGitWithoutCommit(t *testing.T) {
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"new", "demo"}, &out, io.Discard); code != 0 {
		t.Fatal(code, out.String())
	}
	branch, err := exec.Command("git", "-C", "demo", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(branch)) != "main" {
		t.Fatal(string(branch), err)
	}
	if err := exec.Command("git", "-C", "demo", "rev-parse", "--verify", "HEAD").Run(); err == nil {
		t.Fatal("created a commit")
	}
	if _, err := os.Stat("demo/.gitignore"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("demo/flea/icons/.keep"); err != nil {
		t.Fatal("missing icon directory placeholder", err)
	}
	icons, err := os.ReadDir("demo/flea/icons")
	if err != nil {
		t.Fatal(err)
	}
	for _, icon := range icons {
		if icon.IsDir() || icon.Name() == "LICENSE" {
			t.Fatal("icon cache must be flat without a license file")
		}
	}
}

func TestNewSkipGitAndMissingGit(t *testing.T) {
	t.Chdir(t.TempDir())
	for i, args := range [][]string{{"new", "one", "--skip-git"}, {"new", "--skip-git", "two"}} {
		if code := Run(context.Background(), args, io.Discard, io.Discard); code != 0 {
			t.Fatal(code)
		}
		name := []string{"one", "two"}[i]
		if _, err := os.Stat(filepath.Join(name, ".git")); !os.IsNotExist(err) {
			t.Fatal("initialized Git despite skip flag")
		}
	}
	t.Setenv("PATH", t.TempDir())
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"new", "three"}, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "Git isn't installed") {
		t.Fatal(code, out.String())
	}
	if _, err := os.Stat("three/site/index.html"); err != nil {
		t.Fatal("site missing", err)
	}
}
