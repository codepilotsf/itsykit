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

	"itsy/internal/site"
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
	if !strings.HasSuffix(out.String(), "Start developing:\n  cd demo\n  itsy dev\n\n") || strings.Contains(out.String(), "itsy deploy") || strings.Contains(out.String(), "offer to commit") {
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
	for _, file := range []string{".gitignore", "itsy.config.json", "site/index.html", "itsy/icons/.keep", "itsy/skills/"} {
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
	if _, err := os.Stat("demo/itsy/icons/.keep"); err != nil {
		t.Fatal("missing icon directory placeholder", err)
	}
	icons, err := os.ReadDir("demo/itsy/icons")
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

func TestNewDirectoryPaths(t *testing.T) {
	for _, kind := range []string{"current", "existing", "relative", "absolute", "spaces", "shell characters", "leading dash"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			target := "my-project"
			wantCD := "cd my-project"
			args := []string{"new", "--skip-git"}
			switch kind {
			case "current":
				target, wantCD = ".", ""
			case "existing":
				if err := os.Mkdir(target, 0755); err != nil {
					t.Fatal(err)
				}
			case "relative":
				target = "./my-project"
			case "absolute":
				target = filepath.Join(root, target)
				wantCD = "cd " + target
			case "spaces":
				target, wantCD = "my project", "cd 'my project'"
			case "shell characters":
				target, wantCD = "my project's $files", "cd 'my project'\"'\"'s $files'"
			case "leading dash":
				target, wantCD = "-project", "cd './-project'"
				args = append(args, "--")
			}
			args = append(args, target)
			var out, errOut bytes.Buffer
			if code := Run(context.Background(), args, &out, &errOut); code != 0 {
				t.Fatalf("code=%d stdout=%s stderr=%s", code, &out, &errOut)
			}
			p, err := site.Load(target)
			if err != nil {
				t.Fatal(err)
			}
			absolute, err := filepath.Abs(target)
			if err != nil || p.App["site_name"] != site.DefaultSiteName(absolute) {
				t.Fatalf("folder-derived name: %v, %v", p.App["site_name"], err)
			}
			if kind == "current" {
				if !strings.HasSuffix(out.String(), "Start developing:\n  itsy dev\n\n") {
					t.Fatal(out.String())
				}
			} else if !strings.Contains(out.String(), "  "+wantCD+"\n  itsy dev\n") {
				t.Fatal("incorrect directory instructions", out.String())
			}
			if _, err := os.Stat(filepath.Join(target, ".git")); !os.IsNotExist(err) {
				t.Fatal("initialized Git despite skip flag", err)
			}
		})
	}
}

func TestNewCurrentDirectoryInitializesGit(t *testing.T) {
	configureNewTestGit(t, "[user]\nname = Flea Test\nemail = flea@example.test\n")
	t.Chdir(t.TempDir())
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"new", "."}, &out, &errOut); code != 0 {
		t.Fatal(code, out.String(), errOut.String())
	}
	data, err := exec.Command("git", "log", "-1", "--format=%s").Output()
	if err != nil || strings.TrimSpace(string(data)) != "Initial site" {
		t.Fatal("missing initial commit", string(data), err)
	}
	data, err = exec.Command("git", "status", "--porcelain").Output()
	if err != nil || len(data) != 0 {
		t.Fatal("new site should be clean", string(data), err)
	}
}

func TestNewRefusesNonemptyDirectory(t *testing.T) {
	for _, name := range []string{"notes.txt", ".hidden", ".git", "subdirectory"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			if name == ".git" || name == "subdirectory" {
				if err := os.Mkdir(name, 0755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(name, []byte("keep this"), 0644); err != nil {
				t.Fatal(err)
			}
			var out, errOut bytes.Buffer
			if code := Run(context.Background(), []string{"new", "."}, &out, &errOut); code != 1 || !strings.Contains(errOut.String(), "isn't empty. Choose an empty directory") {
				t.Fatal(code, out.String(), errOut.String())
			}
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 1 || entries[0].Name() != name {
				t.Fatal("modified nonempty directory", entries, err)
			}
			if !entries[0].IsDir() {
				data, err := os.ReadFile(name)
				if err != nil || string(data) != "keep this" {
					t.Fatal("modified existing file", string(data), err)
				}
			}
		})
	}
}

func TestNewRefusesFileAndLinkedDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile("file", []byte("keep this"), 0644); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, "linked"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"file", "linked", "linked/"} {
		var errOut bytes.Buffer
		if code := Run(context.Background(), []string{"new", target}, io.Discard, &errOut); code != 1 || !strings.Contains(errOut.String(), "not a directory") {
			t.Fatal(target, code, errOut.String())
		}
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("wrote through linked directory", entries, err)
	}
	data, err := os.ReadFile("file")
	if err != nil || string(data) != "keep this" {
		t.Fatal("modified existing file", string(data), err)
	}
}
