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
	// The development tree records a prerelease version, which the public
	// update channel deliberately refuses. Exercise the staging pipeline with
	// the stable form of the same version so the script's real behavior stays
	// covered during development.
	stable := Version
	if i := strings.IndexAny(stable, "-+"); i >= 0 {
		stable = stable[:i]
	}
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprintf("missing-amd64=%t", missing), func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "system")
			for _, name := range []string{"scripts/stage-release.sh", "scripts/version.sh", "internal/buildinfo/version.go", "../website/site/install.sh"} {
				body, err := os.ReadFile(filepath.Join("../..", name))
				if err != nil {
					t.Fatal(err)
				}
				if name == "internal/buildinfo/version.go" {
					body = []byte(strings.Replace(string(body), `var Version = "`+Version+`"`, `var Version = "`+stable+`"`, 1))
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
			files := []string{"itsy-darwin-arm64", "itsy-linux-amd64"}
			for _, name := range files {
				if missing && name == "itsy-linux-amd64" {
					continue
				}
				if err := os.WriteFile(filepath.Join(root, "dist", name), []byte(name), 0755); err != nil {
					t.Fatal(err)
				}
			}
			// A leftover artifact from an older build must not enter this release.
			if err := os.WriteFile(filepath.Join(root, "dist/flea-linux-arm64"), []byte("old ARM64 CLI"), 0755); err != nil {
				t.Fatal(err)
			}
			run := func() ([]byte, error) {
				cmd := exec.Command("bash", "scripts/stage-release.sh")
				cmd.Dir = root
				cmd.Env = append(os.Environ(), "CI_COMMIT_TAG=v"+stable)
				return cmd.CombinedOutput()
			}
			out, err := run()
			if missing {
				if err == nil || !strings.Contains(string(out), "Missing native-verified artifact: itsy-linux-amd64") {
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
			release := filepath.Join(root, "dist/public/releases", stable)
			if _, err := os.Stat(filepath.Join(release, "flea-linux-arm64")); !os.IsNotExist(err) {
				t.Fatal("staged unsupported Linux ARM64 CLI", err)
			}
			manifest, err := os.ReadFile(filepath.Join(release, "SHA256SUMS"))
			if err != nil {
				t.Fatal(err)
			}
			if len(strings.Fields(string(manifest))) != 4*len(files) {
				t.Fatal("unexpected checksum entries", string(manifest))
			}
			for _, name := range files {
				body, err := os.ReadFile(filepath.Join(release, name))
				if err != nil || string(body) != name {
					t.Fatal("staged artifact differs", name, err)
				}
				legacy := strings.Replace(name, "itsy-", "flea-", 1)
				alias, err := os.ReadFile(filepath.Join(release, legacy))
				if err != nil || string(alias) != string(body) {
					t.Fatal("legacy artifact alias differs", legacy, err)
				}
				legacySum := fmt.Sprintf("%x  %s\n", sha256.Sum256(alias), legacy)
				if !strings.Contains(string(manifest), legacySum) {
					t.Fatal("missing legacy checksum", legacy)
				}
				want := fmt.Sprintf("%x  %s", sha256.Sum256(body), name)
				if !strings.Contains(string(manifest), want+"\n") {
					t.Fatal("missing checksum", want)
				}
			}
			latest, err := os.ReadFile(filepath.Join(root, "dist/public/releases/latest.txt"))
			if err != nil || string(latest) != stable+"\n" {
				t.Fatal("wrong update channel", err, string(latest))
			}
			installer, err := os.ReadFile(filepath.Join(root, "dist/public/install.sh"))
			original, readErr := os.ReadFile("../../../website/site/install.sh")
			if err != nil || readErr != nil || string(installer) != string(original) {
				t.Fatal("installer differs", err, readErr)
			}
			if out, err := run(); err == nil || !strings.Contains(string(out), "Public staging already exists") {
				t.Fatal("overwrote release", err, string(out))
			}
		})
	}
}
