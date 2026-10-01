package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLockSerializesAndRejectsLinks(t *testing.T) {
	name := filepath.Join(t.TempDir(), "lock")
	unlock, err := Lock(name)
	if err != nil {
		t.Fatal(err)
	}
	if release, err := Lock(name); err == nil {
		release()
		t.Fatal("lock was not exclusive")
	} else if !errors.Is(err, ErrBusy) {
		t.Fatalf("busy error = %v, want ErrBusy", err)
	}
	unlock()
	release, err := Lock(name)
	if err != nil {
		t.Fatal(err)
	}
	release()
	link := filepath.Join(t.TempDir(), "linked")
	if err := os.Symlink(name, link); err != nil {
		t.Fatal(err)
	}
	if release, err := Lock(link); err == nil {
		release()
		t.Fatal("followed symlink")
	}
}
