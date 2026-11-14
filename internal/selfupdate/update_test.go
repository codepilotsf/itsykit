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
	"os/exec"
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
	for _, scenario := range []string{"success", "canonical success", "retry linked migration", "checksum mismatch", "invalid executable", "wrong version", "interrupted download", "changed target", "changed candidate", "busy", "already current", "newer local", "duplicate checksum"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "flea")
			if scenario == "canonical success" {
				target = filepath.Join(dir, "itsy")
			}
			if err := os.WriteFile(target, []byte("original CLI"), 0755); err != nil {
				t.Fatal(err)
			}
			if scenario == "retry linked migration" {
				if err := os.Link(target, filepath.Join(dir, "itsy")); err != nil {
					t.Fatal(err)
				}
			}
			payload := binary
			if scenario == "invalid executable" {
				payload = []byte("not an executable")
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(payload))
			if scenario == "checksum mismatch" {
				hash = strings.Repeat("0", 64)
			}
			name := "itsy-" + runtime.GOOS + "-" + runtime.GOARCH
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
			wantSuccess := scenario == "success" || scenario == "canonical success" || scenario == "retry linked migration" || scenario == "already current" || scenario == "newer local"
			if (err == nil) != wantSuccess {
				t.Fatalf("unexpected result: %v %s", err, out.String())
			}
			got, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			switch scenario {
			case "success", "retry linked migration":
				link, linkErr := os.Readlink(target)
				canonical, canonicalErr := os.ReadFile(filepath.Join(dir, "itsy"))
				if linkErr != nil || link != "itsy" || canonicalErr != nil || !bytes.Equal(canonical, payload) {
					t.Fatal("itsy migration failed", link, linkErr, canonicalErr)
				}
				if !verified || !bytes.Equal(got, payload) || !strings.Contains(out.String(), "itsy update-server") {
					t.Fatal("replacement not verified")
				}
			case "canonical success":
				if !verified || !bytes.Equal(got, payload) {
					t.Fatal("canonical update failed")
				}
			case "already current", "newer local":
				if string(got) != "original CLI" {
					t.Fatal("changed current bytes")
				}
				if link, err := os.Readlink(target); err != nil || link != "itsy" {
					t.Fatal("missing already-current conversion", link, err)
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
				if e.Name() != "flea" && !(wantSuccess && e.Name() == "itsy") && !(scenario == "busy" && e.Name() == ".flea-install.lock") {
					t.Fatal("left temporary data", e.Name())
				}
			}
		})
	}
}

func TestMigrationPreservesUnrelatedItsy(t *testing.T) {
	for _, linked := range []bool{false, true} {
		t.Run(fmt.Sprint("symlink=", linked), func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "flea")
			canonical := filepath.Join(dir, "itsy")
			if err := os.WriteFile(target, []byte("old Flea"), 0755); err != nil {
				t.Fatal(err)
			}
			other := canonical
			if linked {
				other = filepath.Join(dir, "other")
			}
			if err := os.WriteFile(other, []byte("independent command"), 0755); err != nil {
				t.Fatal(err)
			}
			if linked {
				if err := os.Symlink("other", canonical); err != nil {
					t.Fatal(err)
				}
			}
			u := updater{system: runtime.GOOS, arch: runtime.GOARCH, executable: target}
			// A nil client proves conflict detection happens before any download.
			err := u.run(context.Background(), "1.0.0", io.Discard)
			if err == nil || !strings.Contains(err.Error(), "itsy already exists") {
				t.Fatal(err)
			}
			for name, want := range map[string]string{target: "old Flea", canonical: "independent command"} {
				got, err := os.ReadFile(name)
				if err != nil || string(got) != want {
					t.Fatal("overwrote independent file", name, err)
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

// This exercises the actual executable's argv handling and production version
// verifier together; fixture-only replacement tests cannot catch that boundary.
func TestRenamedExecutableMigration(t *testing.T) {
	dir := t.TempDir()
	candidate := filepath.Join(dir, "release", "itsy")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", candidate, "../../cmd/itsy")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	payload, err := os.ReadFile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	// Read the recorded version rather than couple this regression to a release.
	versionBytes, err := exec.Command(candidate, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	releaseVersion := strings.TrimSpace(strings.TrimPrefix(string(versionBytes), "ItsyWeb "))
	probe := filepath.Join(dir, ".flea-update-old-client")
	if err := os.WriteFile(probe, payload, 0755); err != nil {
		t.Fatal(err)
	}
	probeOut, err := exec.Command(probe, "version").Output()
	if err != nil || string(probeOut) != "FleaCMS "+releaseVersion+"\n" {
		t.Fatal("released updater cannot verify candidate", string(probeOut), err)
	}
	target := filepath.Join(dir, "flea")
	if err := os.WriteFile(target, payload, 0755); err != nil {
		t.Fatal(err)
	}
	name := "itsy-" + runtime.GOOS + "-" + runtime.GOARCH
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest.txt":
			fmt.Fprintln(w, releaseVersion)
		case "/" + releaseVersion + "/SHA256SUMS":
			fmt.Fprintf(w, "%x  %s\n", sha256.Sum256(payload), name)
		case "/" + releaseVersion + "/" + name:
			w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	u := updater{base: server.URL, system: runtime.GOOS, arch: runtime.GOARCH, executable: target, client: server.Client(), verify: verifyVersion}
	if err := u.run(context.Background(), "0.0.0", io.Discard); err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(dir, "itsy")
	got, err := exec.Command(installed, "version").Output()
	if err != nil || string(got) != string(versionBytes) {
		t.Fatal("installed command failed", string(got), err)
	}
	stopped, err := exec.Command(target, "deploy").CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 || !strings.Contains(string(stopped), "From now on, please use itsy.") {
		t.Fatal("retired command dispatched", string(stopped), err)
	}
}
