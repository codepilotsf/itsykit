//go:build unix

package release

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestExtractIgnoresUmask proves extraction sets directory and file modes
// explicitly, so a restrictive deploy umask cannot strip service-user access.
func TestExtractIgnoresUmask(t *testing.T) {
	previous := syscall.Umask(0o077)
	defer syscall.Umask(previous)
	root := t.TempDir()
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if err := makeDirs(r, "site/deep/nested", 0755); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "site", "deep", "nested"))
	if err != nil || info.Mode().Perm() != 0755 {
		t.Fatalf("directory mode %v %v", info, err)
	}
	f, err := r.OpenFile("site/deep/nested/file", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Chmod(0644); err != nil {
		t.Fatal(err)
	}
	f.Close()
	info, err = os.Stat(filepath.Join(root, "site", "deep", "nested", "file"))
	if err != nil || info.Mode().Perm() != 0644 {
		t.Fatalf("file mode %v %v", info, err)
	}
}
