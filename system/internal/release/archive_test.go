package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveRoundTripAndScope(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"site/index.html": "hello", "itsy.config.json": `{"app":{"site_name":"test"},"admin":{"emails":["owner@example.test"]},"redirects":{"/old.html":"/"},"deploy":{"branches":{"main":{"target":"private-target"}}}}`, "itsy/data/home.json": "{}", "itsy/runtime/sessions.json": "private-session", ".git/config": "secret", "README.md": "local", ".agents/skills/core/SKILL.md": "private agent instructions", ".agents/skills-source.md": "local source provenance"} {
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var b bytes.Buffer
	if err := os.MkdirAll(filepath.Join(root, "mailers"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mailers", "sign-in.email.html"), []byte("{email.code}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Pack(root, &b); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Extract(bytes.NewReader(b.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(dest, "site", "index.html")); err != nil || string(data) != "hello" {
		t.Fatal(string(data), err)
	}
	if _, err := os.Stat(filepath.Join(dest, "itsy.config.json")); err != nil {
		t.Fatal("configuration missing from release", err)
	}
	configBytes, err := os.ReadFile(filepath.Join(dest, "itsy.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]json.RawMessage
	if err = json.Unmarshal(configBytes, &config); err != nil {
		t.Fatal(err)
	}
	if _, ok := config["deploy"]; ok || bytes.Contains(configBytes, []byte("private-target")) {
		t.Fatal("deployment settings leaked into runtime archive")
	}
	if config["app"] == nil || config["admin"] == nil {
		t.Fatal("runtime config fields lost")
	}
	var redirects map[string]string
	if err := json.Unmarshal(config["redirects"], &redirects); err != nil || redirects["/old.html"] != "/" {
		t.Fatal("migration redirects lost from runtime archive", err)
	}
	if data, err := os.ReadFile(filepath.Join(dest, "mailers", "sign-in.email.html")); err != nil || string(data) != "{email.code}" {
		t.Fatal("mailer missing from release", err)
	}
	for _, name := range []string{"itsy/runtime", ".git/config", "README.md", ".agents"} {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatal("included local file", name)
		}
	}
	if err := Extract(bytes.NewReader(b.Bytes()), dest); err == nil {
		t.Fatal("overwrote populated release")
	}
	if err := os.Symlink("index.html", filepath.Join(root, "site", "link")); err != nil {
		t.Fatal(err)
	}
	if err := Pack(root, &bytes.Buffer{}); err == nil {
		t.Fatal("pack accepted symlink")
	}
}

func TestArchiveAppPackage(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string, mode os.FileMode) {
		t.Helper()
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("itsy/apps/session/app.json", `{"name":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session"}}`, 0644)
	write("itsy/apps/session/session", "#!/bin/sh\n", 0755)
	write("itsy/apps/session/data/sessions.json", `{"secret":"state"}`, 0600)
	write("itsy/apps/session/notes.txt", "not packaged", 0644)
	write("site/index.html", "hello", 0644)
	var b bytes.Buffer
	if err := Pack(root, &b); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Extract(bytes.NewReader(b.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dest, "itsy", "apps", "session", "session"))
	if err != nil || info.Mode().Perm() != 0555 {
		t.Fatalf("packaged executable mode %v %v", info.Mode(), err)
	}
	// The service account is not the deployment root, so the packaged executable
	// must be readable and executable by every user and writable by none.
	if info.Mode().Perm()&0o444 != 0o444 || info.Mode().Perm()&0o111 != 0o111 || info.Mode().Perm()&0o222 != 0 {
		t.Fatalf("packaged executable is not service-readable: %v", info.Mode())
	}
	if _, err := os.Stat(filepath.Join(dest, "itsy", "apps", "session", "app.json")); err != nil {
		t.Fatal("manifest missing from package", err)
	}
	for _, name := range []string{"itsy/apps/session/data", "itsy/apps/session/notes.txt", "itsy/apps/session/data/sessions.json"} {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatalf("runtime/unlisted app file packaged: %s", name)
		}
	}
	// An executable mode outside the app package is refused.
	var crafted bytes.Buffer
	gz := gzip.NewWriter(&crafted)
	tw := tar.NewWriter(gz)
	tw.WriteHeader(&tar.Header{Name: "site/tool", Typeflag: tar.TypeReg, Mode: 0555, Size: 1})
	tw.Write([]byte("x"))
	tw.Close()
	gz.Close()
	if err := Extract(&crafted, t.TempDir()); err == nil {
		t.Fatal("accepted executable outside a custom app package")
	}
}

func nativeExecutable(goos, goarch string) []byte {
	switch goos {
	case "linux", "freebsd", "netbsd", "openbsd", "dragonfly", "solaris", "illumos", "android":
		class, machine := byte(2), uint16(62)
		switch goarch {
		case "386":
			class, machine = 1, 3
		case "arm":
			class, machine = 1, 40
		case "arm64":
			class, machine = 2, 183
		case "ppc64", "ppc64le":
			class, machine = 2, 21
		case "riscv64":
			class, machine = 2, 243
		}
		h := make([]byte, 64)
		copy(h, "\x7fELF")
		h[4], h[5] = class, 1
		binary.LittleEndian.PutUint16(h[18:], machine)
		return h
	case "darwin":
		cpu := uint32(0x01000007)
		switch goarch {
		case "arm64":
			cpu = 0x0100000c
		}
		h := make([]byte, 32)
		binary.LittleEndian.PutUint32(h[0:], 0xfeedfacf)
		binary.LittleEndian.PutUint32(h[4:], cpu)
		return h
	}
	return []byte("#!/bin/sh\n")
}

func TestAppPlatformPreflight(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "itsy", "apps", "session")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), data, 0755); err != nil {
			t.Fatal(err)
		}
	}
	write("app.json", []byte(`{"name":"session","version":"0.1.0","protocol":2,"executable":{"darwin-arm64":"session-darwin","linux-amd64":"session-linux"}}`))
	write("session-darwin", nativeExecutable("darwin", "arm64"))
	write("session-linux", nativeExecutable("linux", "amd64"))

	if err := CheckAppPlatforms(root, "linux", "amd64"); err != nil {
		t.Fatalf("preflight rejected a matching native package: %v", err)
	}
	if err := CheckAppPlatforms(root, "darwin", "arm64"); err != nil {
		t.Fatalf("preflight rejected a declared platform: %v", err)
	}
	// A target platform the manifest does not declare is refused.
	if err := CheckAppPlatforms(root, "windows", "amd64"); err == nil {
		t.Fatal("preflight accepted an undeclared platform")
	}
	// A mislabeled (non-native) target executable is refused.
	write("session-linux", []byte("#!/bin/sh\n"))
	if err := CheckAppPlatforms(root, "linux", "amd64"); err == nil {
		t.Fatal("preflight accepted a non-native executable")
	}
	// A native binary for the wrong architecture is refused.
	write("session-linux", nativeExecutable("linux", "arm64"))
	if err := CheckAppPlatforms(root, "linux", "amd64"); err == nil {
		t.Fatal("preflight accepted a wrong-architecture executable")
	}
	// A missing target executable is refused.
	if err := os.Remove(filepath.Join(dir, "session-linux")); err != nil {
		t.Fatal(err)
	}
	if err := CheckAppPlatforms(root, "linux", "amd64"); err == nil {
		t.Fatal("preflight accepted a missing executable")
	}
}

// TestPackRejectsHardlinkedAppArtifact proves the packer refuses a hardlinked
// manifest or declared executable, so private linked bytes cannot be shipped.
func TestPackRejectsHardlinkedAppArtifact(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		write := func(name string, body []byte, mode os.FileMode) {
			t.Helper()
			file := filepath.Join(root, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, body, mode); err != nil {
				t.Fatal(err)
			}
		}
		write("itsy/apps/session/app.json", []byte(`{"name":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session","darwin-arm64":"session-darwin"}}`), 0644)
		write("itsy/apps/session/session", []byte("#!/bin/sh\n"), 0755)
		write("itsy/apps/session/session-darwin", []byte("#!/bin/sh\n"), 0755)
		return root
	}
	root := setup(t)
	if err := os.Link(filepath.Join(root, "itsy", "apps", "session", "session-darwin"), filepath.Join(root, "_other")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	if err := Pack(root, &bytes.Buffer{}); err == nil {
		t.Fatal("pack accepted a hardlinked non-selected executable")
	}

	root = setup(t)
	if err := os.Link(filepath.Join(root, "itsy", "apps", "session", "app.json"), filepath.Join(root, "_manifest")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	if err := Pack(root, &bytes.Buffer{}); err == nil {
		t.Fatal("pack accepted a hardlinked manifest")
	}
}

// TestPackRejectsNonStrictManifest proves the packer uses the same strict
// manifest decoder as the site loader: case-variant or duplicate keys fail.
func TestPackRejectsNonStrictManifest(t *testing.T) {
	for _, manifest := range []string{
		`{"NAME":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session"}}`,
		`{"name":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session"},"protocol":2}`,
		`{"name":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session"},"url":"x"}`,
	} {
		root := t.TempDir()
		file := filepath.Join(root, "itsy", "apps", "session", "app.json")
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(manifest), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "itsy", "apps", "session", "session"), []byte("#!/bin/sh\n"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := Pack(root, &bytes.Buffer{}); err == nil {
			t.Fatalf("pack accepted non-strict manifest %s", manifest)
		}
	}
}

// TestExtractRejectsUnlistedAppFile proves the extractor refuses a package file
// the manifest does not declare, so nothing unlisted is admitted.
func TestExtractRejectsUnlistedAppFile(t *testing.T) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	write := func(name string, mode int64, body string) {
		t.Helper()
		tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: mode, Size: int64(len(body))})
		io.WriteString(tw, body)
	}
	write("itsy/apps/session/app.json", 0644, `{"name":"session","version":"0.1.0","protocol":2,"executable":{"linux-amd64":"session"}}`)
	write("itsy/apps/session/session", 0555, "#!/bin/sh\n")
	write("itsy/apps/session/extra", 0644, "unlisted")
	tw.Close()
	gz.Close()
	if err := Extract(&b, t.TempDir()); err == nil {
		t.Fatal("extracted an unlisted app file")
	}
}

func TestRejectUnsafeArchives(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  byte
		size int64
	}{
		{"../escape", tar.TypeReg, 0}, {"../../escape", tar.TypeReg, 0}, {"/tmp/escape", tar.TypeReg, 0}, {"./index", tar.TypeReg, 0}, {"site/link", tar.TypeSymlink, 0}, {"site/hardlink", tar.TypeLink, 0}, {"site/fifo", tar.TypeFifo, 0}, {"itsy/runtime/sessions.json", tar.TypeReg, 0}, {"site/huge", tar.TypeReg, MaxBytes + 1}, {"site\\escape", tar.TypeReg, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			tw := tar.NewWriter(gz)
			if err := tw.WriteHeader(&tar.Header{Name: tc.name, Typeflag: tc.typ, Mode: 0777, Size: tc.size, Linkname: "../../outside"}); err != nil {
				t.Fatal(err)
			}
			_ = tw.Close()
			_ = gz.Close()
			if err := Extract(&b, t.TempDir()); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
	t.Run("duplicate", func(t *testing.T) {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		tw := tar.NewWriter(gz)
		for i := 0; i < 2; i++ {
			if err := tw.WriteHeader(&tar.Header{Name: "site/index.html", Mode: 0644, Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
		}
		tw.Close()
		gz.Close()
		if err := Extract(&b, t.TempDir()); err == nil {
			t.Fatal("accepted duplicate")
		}
	})
	t.Run("truncated", func(t *testing.T) {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		tw := tar.NewWriter(gz)
		tw.WriteHeader(&tar.Header{Name: "site/css", Typeflag: tar.TypeDir, Mode: 0755})
		tw.Close()
		gz.Close()
		if err := Extract(bytes.NewReader(b.Bytes()[:b.Len()-5]), t.TempDir()); err == nil {
			t.Fatal("accepted truncated gzip")
		}
	})
}
