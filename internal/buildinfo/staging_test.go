package buildinfo

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicReleaseStaging(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprintf("missing-arm64=%t", missing), func(t *testing.T) {
			root := t.TempDir()
			for _, name := range []string{"scripts/stage-release.sh", "scripts/version.sh", "internal/buildinfo/version.go", "website/site/install.sh"} {
				body, err := os.ReadFile(filepath.Join("../..", name))
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, body, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(filepath.Join(root, "dist"), 0755); err != nil {
				t.Fatal(err)
			}
			files := []string{"flea-darwin-arm64", "flea-linux-amd64", "flea-linux-arm64"}
			for _, name := range files {
				if missing && name == "flea-linux-arm64" {
					continue
				}
				if err := os.WriteFile(filepath.Join(root, "dist", name), []byte(name), 0755); err != nil {
					t.Fatal(err)
				}
			}
			run := func() ([]byte, error) {
				cmd := exec.Command("bash", "scripts/stage-release.sh")
				cmd.Dir = root
				cmd.Env = append(os.Environ(), "CI_COMMIT_TAG=v"+Version)
				return cmd.CombinedOutput()
			}
			out, err := run()
			if missing {
				if err == nil || !strings.Contains(string(out), "Missing native-verified artifact: flea-linux-arm64") {
					t.Fatal(err, string(out))
				}
				if _, err := os.Stat(filepath.Join(root, "dist/public")); !os.IsNotExist(err) {
					t.Fatal("staged an incomplete release", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err, string(out))
			}
			release := filepath.Join(root, "dist/public/releases", Version)
			manifest, err := os.ReadFile(filepath.Join(release, "SHA256SUMS"))
			if err != nil {
				t.Fatal(err)
			}
			if len(strings.Fields(string(manifest))) != 2*len(files) {
				t.Fatal("unexpected checksum entries", string(manifest))
			}
			for _, name := range files {
				body, err := os.ReadFile(filepath.Join(release, name))
				if err != nil || string(body) != name {
					t.Fatal("staged artifact differs", name, err)
				}
				want := fmt.Sprintf("%x  %s", sha256.Sum256(body), name)
				if !strings.Contains(string(manifest), want+"\n") {
					t.Fatal("missing checksum", want)
				}
			}
			latest, err := os.ReadFile(filepath.Join(root, "dist/public/releases/latest.txt"))
			if err != nil || string(latest) != Version+"\n" {
				t.Fatal("wrong update channel", err, string(latest))
			}
			installer, err := os.ReadFile(filepath.Join(root, "dist/public/install.sh"))
			original, readErr := os.ReadFile("../../website/site/install.sh")
			if err != nil || readErr != nil || string(installer) != string(original) {
				t.Fatal("installer differs", err, readErr)
			}
			if out, err := run(); err == nil || !strings.Contains(string(out), "Public staging already exists") {
				t.Fatal("overwrote release", err, string(out))
			}
		})
	}
}
