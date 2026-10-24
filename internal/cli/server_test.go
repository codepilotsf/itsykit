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
	if code := Run(context.Background(), []string{"server"}, &out, &errOut); code != 0 || !strings.Contains(out.String(), "Usage: flea server list [ssh-target]") || !strings.HasSuffix(out.String(), "\n\n") || errOut.Len() != 0 {
		t.Fatalf("server help: code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
	// Help for each subcommand prints usage without connecting or prompting.
	for _, args := range [][]string{{"server", "indexing", "--help"}, {"server", "indexing", "-h"}, {"server", "list", "--help"}, {"server", "list", "-h"}, {"server", "remove", "--help"}, {"server", "remove", "-h"}} {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), args, &stdout, &stderr); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "Usage: flea server list [ssh-target]") {
			t.Fatalf("%v: code=%d out=%q err=%q", args, code, stdout.String(), stderr.String())
		}
	}
	// Usage errors never reach setup and never prompt.
	for _, args := range [][]string{
		{"server", "indexing"},
		{"server", "indexing", "server", "example.test", "noindex", "extra"},
		{"server", "bogus"},
		{"server", "sites"},
		{"server", "sites", "-x"},
		{"server", "list", "-x"},
		{"server", "list", "a", "b"},
		{"server", "remove", "-x"},
		{"server", "remove", "a", "b", "c"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Fatalf("%v returned %d (out=%q err=%q)", args, code, stdout.String(), stderr.String())
		}
	}
	// Explicit positional target still validates without a local site.
	if code := Run(context.Background(), []string{"server", "remove", "target", "bad name"}, io.Discard, io.Discard); code != 1 {
		t.Fatalf("invalid site name returned %d", code)
	}
	if code := Run(context.Background(), []string{"server", "list", "bad target!"}, io.Discard, io.Discard); code != 1 {
		t.Fatalf("invalid target returned %d", code)
	}
}
