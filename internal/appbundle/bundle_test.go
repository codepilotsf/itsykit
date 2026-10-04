package appbundle

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract(t *testing.T) {
	if !Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("missing Contact bundle")
		}
		t.Skip("requires scripts/build.sh")
	}
	target := filepath.Join(t.TempDir(), "contact")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Extract(ctx, target) == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("partial executable retained")
	}
	if err := Extract(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Size() == 0 || info.Size() > 20<<20 || info.Mode().Perm() != 0500 {
		t.Fatal("invalid executable", err)
	}
	if Extract(context.Background(), target) == nil {
		t.Fatal("overwrote existing file")
	}
	after, _ := os.Stat(target)
	if !os.SameFile(info, after) {
		t.Fatal("existing file replaced")
	}
}
