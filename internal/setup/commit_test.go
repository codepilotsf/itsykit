package setup

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareCommit(t *testing.T) {
	for _, mode := range []string{"accept", "decline", "eof", "hook", "initial"} {
		t.Run(mode, func(t *testing.T) {
			root := repoSite(t)
			if mode == "initial" {
				testGit(t, root, "checkout", "--orphan", "fresh")
			}
			if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("Updated"), 0644); err != nil {
				t.Fatal(err)
			}
			if mode == "hook" {
				if err := os.WriteFile(filepath.Join(root, ".git/hooks/pre-commit"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			input := "y\nMy site update\n"
			if mode == "initial" {
				input = "y\n\n"
			}
			if mode == "accept" {
				input = "y\n\nMy site update\n"
			}
			if mode == "decline" {
				input = "n\n"
			}
			if mode == "eof" {
				input = ""
			}
			var output bytes.Buffer
			err := prepareCommit(context.Background(), root, "", strings.NewReader(input), &output, io.Discard)
			if mode == "accept" || mode == "initial" {
				if err != nil {
					t.Fatal(err)
				}
				if _, err := currentGit(context.Background(), root, true); err != nil {
					t.Fatal(err)
				}
				msg, _ := git(context.Background(), root, "log", "-1", "--format=%s")
				want := "My site update"
				if mode == "initial" {
					want = "init"
				}
				if strings.TrimSpace(msg) != want {
					t.Fatal(msg)
				}
				if mode == "initial" && !strings.Contains(output.String(), "Commit message [init]:") {
					t.Fatal(output.String())
				}
				if mode == "accept" && (strings.Contains(output.String(), "Commit message [") || !strings.Contains(output.String(), "Please enter a commit message.")) {
					t.Fatal(output.String())
				}
			} else if err == nil {
				t.Fatal("expected stopped deployment")
			}
		})
	}
}
