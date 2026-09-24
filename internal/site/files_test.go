package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfinedWritesAndReads(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "nested/page.html", "before")
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, body := range []string{"after", "short", ""} {
		if err := writeSiteFile(root, "nested/page.html", []byte(body)); err != nil {
			t.Fatal(err)
		}
		got, err := readFile(root, "nested/page.html")
		if err != nil || string(got) != body {
			t.Fatalf("read %q: %v", got, err)
		}
	}
	if _, err := readFile(root, "missing/file"); !os.IsNotExist(err) {
		t.Fatalf("missing directory lost its error type: %v", err)
	}
	for _, name := range []string{"../outside", "/absolute", "nested/../page.html", "nested//page.html", `nested\page.html`} {
		if _, err := readFile(root, name); err == nil {
			t.Errorf("read accepted %q", name)
		}
		if err := writeSiteFile(root, name, []byte("unsafe")); err == nil {
			t.Errorf("write accepted %q", name)
		}
	}
	outside := t.TempDir()
	put(t, outside, "secret", "untouched")
	for name, target := range map[string]string{"linked-dir": outside, "linked-file": filepath.Join(outside, "secret")} {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"linked-dir/secret", "linked-file", "nested"} {
		if _, err := readFile(root, name); err == nil {
			t.Errorf("read accepted %q", name)
		}
		if err := writeSiteFile(root, name, []byte("unsafe")); err == nil {
			t.Errorf("write accepted %q", name)
		}
	}
	got, err := os.ReadFile(filepath.Join(outside, "secret"))
	if err != nil || string(got) != "untouched" {
		t.Fatal("external file changed", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "nested"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".flea-") {
			t.Fatal("temporary file retained")
		}
	}
}

func TestDataDirectoriesShareValidation(t *testing.T) {
	for _, directory := range []string{"_/fleas", "_/data"} {
		t.Run(directory, func(t *testing.T) {
			root := fixture(t)
			put(t, root, directory+"/extra.json", `{"title":"Hello"}`)
			if _, err := Load(root); err != nil {
				t.Fatal(err)
			}
			put(t, root, directory+"/extra.json", `null`)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "JSON object") {
				t.Fatal(err)
			}
		})
	}
	root := fixture(t)
	put(t, root, "_/fleas/duplicate.json", `{}`)
	put(t, root, "_/data/duplicate.json", `{}`)
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "duplicate legacy flea/data") {
		t.Fatal(err)
	}
}

func TestPageRoutesMatchFleaScopes(t *testing.T) {
	for file, want := range map[string]string{"index.html": "/", "about.html": "/about", "about/index.html": "/about", "about/history.html": "/about/history"} {
		if got := pageRoute(file); got != want {
			t.Errorf("%s: %s", file, got)
		}
		if got := fleaScope(file); got != want {
			t.Errorf("scope %s: %s", file, got)
		}
	}
	for _, file := range []string{"_layout.html", "about/_card.html", "_shared/card.html"} {
		if got := fleaScope(file); got != "" {
			t.Errorf("shared scope %s: %s", file, got)
		}
	}
}
