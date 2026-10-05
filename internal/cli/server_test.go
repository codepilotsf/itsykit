package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestServerCommandOutsideSite(t *testing.T) {
	t.Chdir(t.TempDir())
	var out, errOut bytes.Buffer
	if code := Run(context.Background(), []string{"server"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "Usage: flea server sites") || !strings.HasSuffix(out.String(), "\n\n") || errOut.Len() != 0 {
		t.Fatalf("server help: code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
	for _, args := range [][]string{{"server", "bogus"}, {"server", "sites"}, {"server", "sites", "-x"}, {"server", "remove"}, {"server", "remove", "target"}} {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Fatalf("%v returned %d (out=%q err=%q)", args, code, stdout.String(), stderr.String())
		}
	}
	// An invalid site name fails before any SSH connection is attempted.
	if code := Run(context.Background(), []string{"server", "remove", "target", "bad name"}, io.Discard, io.Discard); code != 1 {
		t.Fatalf("invalid site name returned %d", code)
	}
	// An invalid target fails during validation, still without a local site.
	if code := Run(context.Background(), []string{"server", "sites", "bad target!"}, io.Discard, io.Discard); code != 1 {
		t.Fatalf("invalid target returned %d", code)
	}
}
