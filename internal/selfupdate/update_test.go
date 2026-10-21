package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUpdateVerifiedReplacementAndFailures(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"success", "checksum mismatch", "invalid executable", "wrong version", "interrupted download", "changed target", "changed candidate", "busy", "already current", "newer local", "duplicate checksum"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "flea")
			if err := os.WriteFile(target, []byte("original CLI"), 0755); err != nil {
				t.Fatal(err)
			}
			payload := binary
			if scenario == "invalid executable" {
				payload = []byte("not an executable")
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(payload))
			if scenario == "checksum mismatch" {
				hash = strings.Repeat("0", 64)
			}
			name := "flea-" + runtime.GOOS + "-" + runtime.GOARCH
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/latest.txt":
					fmt.Fprintln(w, "1.0.1")
				case "/1.0.1/SHA256SUMS":
					fmt.Fprintf(w, "%s  %s\n", hash, name)
					if scenario == "duplicate checksum" {
						fmt.Fprintf(w, "%s  %s\n", hash, name)
					}
				case "/1.0.1/" + name:
					if scenario == "interrupted download" {
						w.Header().Set("Content-Length", fmt.Sprint(len(payload)+100))
					}
					w.Write(payload)
				default:
					http.NotFound(w, r)
				}
			}))
			defer s.Close()
			verified := false
			u := updater{base: s.URL, system: runtime.GOOS, arch: runtime.GOARCH, executable: target, client: s.Client(), verify: func(ctx context.Context, p, version string) error {
				verified = true
				if version != "1.0.1" {
					t.Error(version)
				}
				if scenario == "wrong version" {
					return fmt.Errorf("unexpected version")
				}
				if scenario == "changed target" {
					return os.WriteFile(target, []byte("independent change"), 0755)
				}
				if scenario == "changed candidate" {
					return os.WriteFile(p, []byte("independent change"), 0755)
				}
				return nil
			}}
			if scenario == "busy" {
				if err := os.Mkdir(filepath.Join(dir, ".flea-install.lock"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			current := "1.0.0"
			if scenario == "already current" {
				current = "1.0.1"
			}
			if scenario == "newer local" {
				current = "1.1.0-rc.1"
			}
			var out bytes.Buffer
			err := u.run(context.Background(), current, &out)
			wantSuccess := scenario == "success" || scenario == "already current" || scenario == "newer local"
			if (err == nil) != wantSuccess {
				t.Fatalf("unexpected result: %v %s", err, out.String())
			}
			got, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			switch scenario {
			case "success":
				if !verified || !bytes.Equal(got, payload) || !strings.Contains(out.String(), "flea update-server") {
					t.Fatal("replacement not verified")
				}
			case "changed target":
				if string(got) != "independent change" {
					t.Fatal("overwrote independent edit")
				}
			default:
				if string(got) != "original CLI" {
					t.Fatal("lost previous CLI")
				}
			}
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if e.Name() != "flea" && !(scenario == "busy" && e.Name() == ".flea-install.lock") {
					t.Fatal("left temporary data", e.Name())
				}
			}
		})
	}
}

func TestStableVersionSelection(t *testing.T) {
	for _, tc := range []struct {
		latest, current string
		want, invalid   bool
	}{
		{"1.0.0", "0.9.0", true, false}, {"1.0.0", "1.0.0-rc.1", true, false},
		{"1.0.0", "1.0.0+build", false, false}, {"1.0.0", "2.0.0", false, false},
		{"1.10.0", "1.9.0", true, false}, {"../../bad", "1.0.0", false, true},
		{"1.0.0-rc.1", "1.0.0", false, true}, {"01.0.0", "1.0.0", false, true},
	} {
		got, err := newer(tc.latest, tc.current)
		if got != tc.want || (err != nil) != tc.invalid {
			t.Fatal(tc, got, err)
		}
	}
}

func TestDownloadRequiresHTTPSAndBounds(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, strings.Repeat("x", 129)) }))
	defer s.Close()
	u := updater{base: s.URL, client: s.Client()}
	var b bytes.Buffer
	if err := u.download(context.Background(), "latest.txt", &b, 128); err == nil {
		t.Fatal("accepted oversized metadata")
	}
	u.base = "http://example.test"
	if err := u.download(context.Background(), "latest.txt", &b, 128); err == nil {
		t.Fatal("accepted HTTP")
	}
}

func TestUpdateRejectsUnsupportedPlatformsBeforeDownloading(t *testing.T) {
	for _, platform := range [][2]string{{"linux", "arm64"}, {"darwin", "amd64"}, {"windows", "amd64"}} {
		t.Run(strings.Join(platform[:], "/"), func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "flea")
			if err := os.WriteFile(target, []byte("existing CLI"), 0755); err != nil {
				t.Fatal(err)
			}
			u := updater{system: platform[0], arch: platform[1], executable: target}
			err := u.run(context.Background(), "0.17.0", io.Discard)
			if err == nil || !strings.Contains(err.Error(), "Apple Silicon macOS and x86-64 Linux") {
				t.Fatal("missing unsupported-platform diagnostic", err)
			}
			data, err := os.ReadFile(target)
			if err != nil || string(data) != "existing CLI" {
				t.Fatal("changed existing CLI", string(data), err)
			}
		})
	}
}
