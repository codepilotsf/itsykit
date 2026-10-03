package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreSessionFilesAppendsMissingRulesOnce(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n/.env\n")
	for i := 0; i < 3; i++ {
		if err := IgnoreSessionFiles(dir); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	want := ".env\n/.env\n/flea/runtime/\n/flea/media/*.mp4\n"
	if got := readGitignore(t, dir); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestIgnoreSessionFilesKeepsRulesAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	// The rules are present but not at the end of the file (a trailing /.env is),
	// which the old suffix check missed and would duplicate on every run.
	original := "/flea/runtime/\n/flea/media/*.mp4\n/.env\n"
	writeGitignore(t, dir, original)
	for i := 0; i < 3; i++ {
		if err := IgnoreSessionFiles(dir); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	if got := readGitignore(t, dir); got != original {
		t.Fatalf("got %q, want %q", got, original)
	}
}

func TestIgnoreSessionFilesCreatesGitignore(t *testing.T) {
	dir := t.TempDir()
	if err := IgnoreSessionFiles(dir); err != nil {
		t.Fatal(err)
	}
	want := "/flea/runtime/\n/flea/media/*.mp4\n"
	if got := readGitignore(t, dir); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func writeGitignore(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readGitignore(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
