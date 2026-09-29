package selfupdate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallerOffline(t *testing.T) {
	installer, err := filepath.Abs("../../website/site/install.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []bool{false, true} {
		t.Run(fmt.Sprint(bad), func(t *testing.T) {
			dir := t.TempDir()
			fake := filepath.Join(dir, "tools")
			fixtures := filepath.Join(dir, "fixtures")
			destination := filepath.Join(dir, "with spaces", "bin")
			for _, p := range []string{fake, fixtures, destination} {
				if err := os.MkdirAll(p, 0755); err != nil {
					t.Fatal(err)
				}
			}
			payload := []byte("#!/bin/sh\nprintf 'FleaCMS 1.0.1\\n'\n")
			name := "flea-" + runtime.GOOS + "-" + runtime.GOARCH
			hash := fmt.Sprintf("%x", sha256.Sum256(payload))
			if bad {
				hash = strings.Repeat("0", 64)
			}
			for name, body := range map[string][]byte{"latest.txt": []byte("1.0.1\n"), "SHA256SUMS": []byte(hash + "  " + name + "\n"), name: payload} {
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
[[ "$url" == https://fleacms.com/releases/* ]] || exit 1
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
			cmd.Env = append(os.Environ(), "PATH="+fake+":"+os.Getenv("PATH"), "FIXTURES="+fixtures, "FLEA_INSTALL_DIR="+destination)
			out, err := cmd.CombinedOutput()
			if (err != nil) != bad {
				t.Fatal(err, string(out))
			}
			got, _ := os.ReadFile(target)
			if bad {
				if string(got) != "old installation" {
					t.Fatal("installer replaced before checksum validation")
				}
			} else {
				if string(got) != string(payload) || !strings.Contains(string(out), "export PATH=") {
					t.Fatal("install/PATH instructions failed", string(out))
				}
				// Reinstalling is also safe and leaves no staging directories.
				cmd = exec.Command("bash", installer)
				cmd.Env = append(os.Environ(), "PATH="+fake+":"+os.Getenv("PATH"), "FIXTURES="+fixtures, "FLEA_INSTALL_DIR="+destination)
				if out, err = cmd.CombinedOutput(); err != nil {
					t.Fatal(err, string(out))
				}
			}
			entries, _ := os.ReadDir(destination)
			if len(entries) != 1 || entries[0].Name() != "flea" {
				t.Fatal("installer left staging files", entries)
			}
		})
	}
}
