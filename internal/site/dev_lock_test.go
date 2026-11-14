package site

import (
	"strings"
	"testing"
)

func TestLockDevelopmentReportsConcurrentUse(t *testing.T) {
	root := t.TempDir()
	unlock, err := LockDevelopment(root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = LockDevelopment(root)
	if err == nil {
		t.Fatal("second development lock succeeded")
	}
	message := err.Error()
	if !strings.Contains(message, "another running itsy dev, pull or reset") {
		t.Fatalf("unclear busy message: %q", message)
	}
	if strings.Contains(message, "Content or deployment") {
		t.Fatalf("busy message repeats the generic lock text: %q", message)
	}
	unlock()
	release, err := LockDevelopment(root)
	if err != nil {
		t.Fatal("lock was not released:", err)
	}
	release()
}
