package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/site"
)

func TestCleanAndStarterCreation(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"new", "clean", "--skip-git"}, {"new", "--starter", "full", "--skip-git"}, {"new", "short", "-s", "--skip-git"}} {
		if code := Run(context.Background(), args, io.Discard, io.Discard); code != 0 {
			t.Fatal(args, code)
		}
	}
	for _, name := range []string{"clean", "full", "short"} {
		p, err := site.LoadEditing(context.Background(), name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = os.Stat(filepath.Join(name, "site", "team.html"))
		if name == "clean" {
			if !os.IsNotExist(err) {
				t.Fatal("clean has team page")
			}
			if len(p.Content.Pages) != 1 {
				t.Fatal("clean has demo content")
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Content.Pages["/team"].Multisnips["team"]) != 6 {
				t.Fatal("starter must have six fictional team members")
			}
		}
	}
}

func TestResetKeepsGitAndRecoverableFiles(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	if err := createSite(root, true); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{".git/HEAD": "history", ".env": "secret", "itsy/media/upload.jpg": "photo", "custom.txt": "custom"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := resetSite(root, strings.NewReader("\n"), &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "custom.txt")); err != nil {
		t.Fatal("cancel changed files", err)
	}
	unlock, err := site.LockDevelopment(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := resetSite(root, strings.NewReader("y\n"), &out); err == nil {
		t.Fatal("reset during dev")
	}
	unlock()
	if err := resetSite(root, strings.NewReader("yes\n"), &out); err != nil {
		t.Fatal(err)
	}
	backups, err := filepath.Glob(filepath.Join(parent, ".itsy-reset-backup-*"))
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	info, _ := os.Stat(backups[0])
	if info.Mode().Perm() != 0700 {
		t.Fatal("backup permissions")
	}
	for _, name := range []string{".env", "itsy/media/upload.jpg", "custom.txt", "site/team.html"} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("reset retained %s", name)
		}
		if _, err := os.Stat(filepath.Join(backups[0], name)); err != nil {
			t.Fatal(name, err)
		}
	}
	if b, err := os.ReadFile(filepath.Join(root, ".git", "HEAD")); err != nil || string(b) != "history" {
		t.Fatal("git changed", err)
	}
	if _, err := site.Load(root); err != nil {
		t.Fatal(err)
	}
}

func TestResetRejectsNonSitesAndLinks(t *testing.T) {
	root := t.TempDir()
	if err := resetSite(root, strings.NewReader("y\n"), io.Discard); err == nil {
		t.Fatal("reset non-site")
	}
	real := filepath.Join(root, "real")
	if err := Create(real); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if err := resetSite(link, strings.NewReader("y\n"), io.Discard); err == nil {
		t.Fatal("reset linked root")
	}
	if err := os.WriteFile(filepath.Join(real, ".git"), []byte("gitdir: elsewhere"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := resetSite(real, strings.NewReader("y\n"), io.Discard); err == nil {
		t.Fatal("reset worktree")
	}
}
