package selfupdate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerOffline(t *testing.T) {
	installer, err := filepath.Abs("../../../website/site/install.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name, system, machine, artifact string
		bad, missing, unsupported       bool
	}{
		{"mac arm64", "Darwin", "arm64", "itsy-darwin-arm64", false, false, false},
		{"linux amd64", "Linux", "x86_64", "itsy-linux-amd64", false, false, false},
		{"linux arm64", "Linux", "aarch64", "itsy-linux-arm64", false, false, true},
		{"linux arm64 alias", "Linux", "arm64", "itsy-linux-arm64", false, false, true},
		{"wrong version", "Linux", "x86_64", "itsy-linux-amd64", false, false, false},
		{"bad checksum", "Linux", "x86_64", "itsy-linux-amd64", true, false, false},
		{"missing linux amd64", "Linux", "x86_64", "itsy-linux-amd64", false, true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			fake := filepath.Join(dir, "tools")
			fixtures := filepath.Join(dir, "fixtures")
			destination := filepath.Join(dir, "with spaces", "bin")
			for _, p := range []string{fake, fixtures, destination} {
				if err := os.MkdirAll(p, 0755); err != nil {
					t.Fatal(err)
				}
			}
			payload := []byte("#!/bin/sh\nprintf 'ItsyKit 1.0.1\\n'\n")
			if scenario.name == "wrong version" {
				payload = []byte("#!/bin/sh\nprintf 'ItsyKit 0.0.0\\n'\n")
			}
			name := scenario.artifact
			hash := fmt.Sprintf("%x", sha256.Sum256(payload))
			if scenario.bad {
				hash = strings.Repeat("0", 64)
			}
			manifest := hash + "  " + name + "\n"
			if scenario.missing {
				manifest = hash + "  itsy-darwin-arm64\n"
			}
			uname := fmt.Sprintf("#!/bin/sh\ncase \"$1\" in -s) echo %s;; -m) echo %s;; *) exit 1;; esac\n", scenario.system, scenario.machine)
			if err := os.WriteFile(filepath.Join(fake, "uname"), []byte(uname), 0755); err != nil {
				t.Fatal(err)
			}
			for name, body := range map[string][]byte{"latest.txt": []byte("1.0.1\n"), "SHA256SUMS": []byte(manifest), name: payload} {
				if err := os.WriteFile(filepath.Join(fixtures, name), body, 0644); err != nil {
					t.Fatal(err)
				}
			}
			curl := `#!/bin/bash
set -e
output=''
for ((i=1;i<=$#;i++)); do
  if [[ "${!i}" == --output ]]; then j=$((i+1)); output=${!j}; fi
done
url=${!#}
[[ "$url" == https://itsykit.dev/releases/* ]] || exit 1
name=${url##*/}
if [[ -n "$output" ]]; then cp "$FIXTURES/$name" "$output"; else cat "$FIXTURES/$name"; fi
`
			if err := os.WriteFile(filepath.Join(fake, "curl"), []byte(curl), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(destination, "flea")
			if err := os.WriteFile(target, []byte("old installation"), 0755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", installer)
			cmd.Env = append(os.Environ(), "HOME="+dir, "SHELL=/bin/bash", "PATH="+fake+":"+os.Getenv("PATH"), "FIXTURES="+fixtures, "FLEA_INSTALL_DIR="+destination)
			out, err := cmd.CombinedOutput()
			if (err != nil) != (scenario.bad || scenario.missing || scenario.unsupported || scenario.name == "wrong version") {
				t.Fatal(err, string(out))
			}
			got, _ := os.ReadFile(target)
			if scenario.bad || scenario.missing || scenario.unsupported || scenario.name == "wrong version" {
				if string(got) != "old installation" {
					t.Fatal("installer replaced before checksum validation")
				}
				if scenario.unsupported && !strings.Contains(string(out), "Linux ARM64 CLI downloads are no longer supported") {
					t.Fatal("missing platform diagnostic", string(out))
				}
				if scenario.missing && !strings.Contains(string(out), "Itsy 1.0.1 has no verified download for linux/amd64") {
					t.Fatal("missing platform diagnostic", string(out))
				}
			} else {
				if string(got) != string(payload) || !strings.Contains(string(out), "export PATH=") {
					t.Fatal("install/PATH instructions failed", string(out))
				}
				link, linkErr := os.Readlink(target)
				canonical, canonicalErr := os.ReadFile(filepath.Join(destination, "itsy"))
				if linkErr != nil || link != "itsy" || canonicalErr != nil || string(canonical) != string(payload) {
					t.Fatal("missing canonical itsy installation", link, linkErr, canonicalErr)
				}
				// Reinstalling is also safe and leaves no staging directories.
				cmd = exec.Command("bash", installer)
				cmd.Env = append(os.Environ(), "HOME="+dir, "SHELL=/bin/bash", "PATH="+fake+":"+os.Getenv("PATH"), "FIXTURES="+fixtures, "FLEA_INSTALL_DIR="+destination)
				if out, err = cmd.CombinedOutput(); err != nil {
					t.Fatal(err, string(out))
				}
			}
			entries, _ := os.ReadDir(destination)
			wantEntries := 2
			if scenario.bad || scenario.missing || scenario.unsupported || scenario.name == "wrong version" {
				wantEntries = 1
			}
			if len(entries) != wantEntries {
				t.Fatal("installer left staging files", entries)
			}
		})
	}
}
