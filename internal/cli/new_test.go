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

func configureNewTestGit(t *testing.T, config string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(file, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", file)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func TestNewInitializesGitWithCommit(t *testing.T) {
	configureNewTestGit(t, "[user]\nname = Flea Test\nemail = flea@example.test\n")
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"new", "demo"}, &out, io.Discard); code != 0 {
		t.Fatal(code, out.String())
	}
	if !strings.HasSuffix(out.String(), "Start developing:\n  cd demo\n  flea dev\n\n") || strings.Contains(out.String(), "flea deploy") || strings.Contains(out.String(), "offer to commit") {
		t.Fatal("new should finish with local development instructions", out.String())
	}
	branch, err := exec.Command("git", "-C", "demo", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(branch)) != "main" {
		t.Fatal(string(branch), err)
	}
	if err := exec.Command("git", "-C", "demo", "rev-parse", "--verify", "HEAD").Run(); err != nil {
		t.Fatal("missing initial commit", err)
	}
	status, err := exec.Command("git", "-C", "demo", "status", "--porcelain").Output()
	if err != nil || len(status) != 0 {
		t.Fatal("new site should be clean", string(status), err)
	}
	files, err := exec.Command("git", "-C", "demo", "ls-tree", "-r", "--name-only", "HEAD").Output()
	for _, file := range []string{".gitignore", ".flea.config.json", "site/index.html", "flea/icons/.keep", "flea/skills/"} {
		if err != nil || !strings.Contains(string(files), file) {
			t.Fatal("initial commit missing", file, err)
		}
	}
	remotes, err := exec.Command("git", "-C", "demo", "remote").Output()
	if err != nil || len(remotes) != 0 {
		t.Fatal("should not configure a remote", string(remotes), err)
	}
	if !strings.Contains(out.String(), "git remote add origin <repository-url>") || !strings.Contains(out.String(), "git push -u origin main") {
		t.Fatal("missing remote instructions", out.String())
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

func TestNewCommitFailureKeepsSite(t *testing.T) {
	configureNewTestGit(t, "[user]\nuseConfigOnly = true\n")
	for _, key := range []string{"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL", "EMAIL"} {
		t.Setenv(key, "")
	}
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"new", "demo"}, &out, io.Discard); code != 0 {
		t.Fatal("site creation should succeed", code, out.String())
	}
	if !strings.Contains(out.String(), "initial Git commit could not be created") || !strings.Contains(out.String(), "git commit -m \"Initial site\"") || strings.Contains(out.String(), "git push") {
		t.Fatal("missing recovery instructions", out.String())
	}
	if _, err := os.Stat("demo/site/index.html"); err != nil {
		t.Fatal("site lost after commit failure", err)
	}
	if err := exec.Command("git", "-C", "demo", "rev-parse", "--verify", "HEAD").Run(); err == nil {
		t.Fatal("unexpected commit without identity")
	}
}

func TestNewRespectsCommitHooks(t *testing.T) {
	root := t.TempDir()
	hooks := filepath.Join(root, "hooks")
	if err := os.Mkdir(hooks, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooks, "pre-commit"), []byte("#!/bin/sh\necho 'Test hook refused commit' >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	configureNewTestGit(t, "[user]\nname = Flea Test\nemail = flea@example.test\n[core]\nhooksPath = "+hooks+"\n")
	t.Chdir(root)
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"new", "demo"}, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "Test hook refused commit") {
		t.Fatal(code, out.String())
	}
	if err := exec.Command("git", "-C", "demo", "rev-parse", "--verify", "HEAD").Run(); err == nil {
		t.Fatal("bypassed commit hook")
	}
}

func TestNewIgnoresGitLocationOverrides(t *testing.T) {
	configureNewTestGit(t, "[user]\nname = Flea Test\nemail = flea@example.test\n")
	root := t.TempDir()
	other := filepath.Join(root, "other")
	if err := os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	if data, err := exec.Command("git", "-C", other, "init", "--quiet", "-b", "other").CombinedOutput(); err != nil {
		t.Fatal(string(data), err)
	}
	if err := os.WriteFile(filepath.Join(other, "untouched.txt"), []byte("other project"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	t.Run("create with overrides", func(t *testing.T) {
		t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
		t.Setenv("GIT_COMMON_DIR", filepath.Join(other, ".git"))
		t.Setenv("GIT_WORK_TREE", other)
		t.Setenv("GIT_INDEX_FILE", filepath.Join(other, ".git", "index"))
		var out bytes.Buffer
		if code := Run(context.Background(), []string{"new", "demo"}, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "with an initial commit") {
			t.Fatal(code, out.String())
		}
	})
	if err := exec.Command("git", "-C", "demo", "rev-parse", "--verify", "HEAD").Run(); err != nil {
		t.Fatal("new site missing commit", err)
	}
	if err := exec.Command("git", "-C", other, "rev-parse", "--verify", "HEAD").Run(); err == nil {
		t.Fatal("committed another repository")
	}
	status, err := exec.Command("git", "-C", other, "status", "--porcelain").Output()
	if err != nil || string(status) != "?? untouched.txt\n" {
		t.Fatal("changed another repository's index", string(status), err)
	}
	branch, err := exec.Command("git", "-C", other, "symbolic-ref", "--short", "HEAD").Output()
	if err != nil || string(branch) != "other\n" {
		t.Fatal("changed another repository's branch", string(branch), err)
	}
}
