package site

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"flea/internal/appapi"
)

var testPlatform = runtime.GOOS + "-" + runtime.GOARCH

// nativeExecutable returns a minimal but valid native executable header for the
// running platform so load-time format validation accepts a fixture without
// building or executing anything. Unsupported platforms keep a shell script.
func nativeExecutable() []byte {
	switch runtime.GOOS {
	case "linux", "freebsd", "netbsd", "openbsd", "dragonfly", "solaris", "illumos", "android":
		class, machine := byte(2), uint16(62) // amd64 default
		switch runtime.GOARCH {
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
		case "s390x":
			class, machine = 2, 22
		}
		h := make([]byte, 64)
		copy(h, "\x7fELF")
		h[4], h[5] = class, 1
		binary.LittleEndian.PutUint16(h[18:], machine)
		return h
	case "darwin":
		cpu := uint32(0x01000007) // x86_64
		switch runtime.GOARCH {
		case "arm64":
			cpu = 0x0100000c
		case "386":
			cpu = 7
		case "arm":
			cpu = 12
		}
		h := make([]byte, 32)
		binary.LittleEndian.PutUint32(h[0:], 0xfeedfacf)
		binary.LittleEndian.PutUint32(h[4:], cpu)
		return h
	}
	return []byte("#!/bin/sh\n")
}

func manifest(name, version string, protocol int, executable string) string {
	return fmt.Sprintf(`{"name":%q,"version":%q,"protocol":%d,"executable":{%q:%q}}`, name, version, protocol, testPlatform, executable)
}

func writeExecutable(t *testing.T, root, name string) {
	t.Helper()
	dest := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, nativeExecutable(), 0755); err != nil {
		t.Fatal(err)
	}
}

func TestCustomAppRegistration(t *testing.T) {
	valid := `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`
	validManifest := manifest("demo", "1.0.0", 2, "demo")

	load := func(t *testing.T, settings, man string, executable bool, page bool) error {
		t.Helper()
		root := fixture(t)
		put(t, root, ".flea.config.json", settings)
		if man != "" {
			put(t, root, "flea/apps/demo/app.json", man)
		}
		if executable {
			writeExecutable(t, root, "flea/apps/demo/demo")
		}
		if page {
			put(t, root, "demo/index.html", `<form method="post" action="/demo">{csrf}</form>`)
		}
		_, err := Load(root)
		return err
	}

	if err := load(t, valid, validManifest, true, true); err != nil {
		t.Fatalf("valid custom app rejected: %v", err)
	}
	for _, tc := range []struct {
		name       string
		settings   string
		manifest   string
		executable bool
		page       bool
	}{
		{"name mismatch", valid, manifest("other", "1.0.0", 2, "demo"), true, true},
		{"protocol mismatch", valid, manifest("demo", "1.0.0", 1, "demo"), true, true},
		{"missing version", valid, manifest("demo", "", 2, "demo"), true, true},
		{"unknown manifest key", valid, fmt.Sprintf(`{"name":"demo","version":"1.0.0","protocol":2,"executable":{%q:"demo"},"url":"http://x"}`, testPlatform), true, true},
		{"path executable", valid, manifest("demo", "1.0.0", 2, "../demo"), true, true},
		{"missing platform", valid, fmt.Sprintf(`{"name":"demo","version":"1.0.0","protocol":2,"executable":{"plan9-386":"demo"}}`), true, true},
		{"missing executable", valid, validManifest, false, true},
		{"missing package", valid, "", false, true},
		{"missing page", valid, validManifest, true, false},
		{"duplicate field", valid, fmt.Sprintf(`{"name":"demo","version":"1.0.0","protocol":2,"executable":{%q:"demo"},"protocol":2}`, testPlatform), true, true},
		{"uppercase field", valid, fmt.Sprintf(`{"NAME":"demo","version":"1.0.0","protocol":2,"executable":{%q:"demo"}}`, testPlatform), true, true},
		{"trailing bytes", valid, validManifest + "x", true, true},
		{"email on custom app", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo","email":{"to":"owner@example.test"}}}}`, validManifest, true, true},
		{"reserved name", `{"app":{"site_name":"Test"},"apps":{"data":{"mount":"/demo"}}}`, manifest("data", "1", 2, "demo"), true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := load(t, tc.settings, tc.manifest, tc.executable, tc.page); err == nil {
				t.Fatal("invalid custom app accepted")
			}
		})
	}

	// A non-executable file is not a runnable package.
	root := fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", validManifest)
	put(t, root, "flea/apps/demo/demo", "not executable")
	put(t, root, "demo/index.html", "Form")
	if _, err := Load(root); err == nil {
		t.Fatal("non-executable package accepted")
	}

	// A symlinked executable is refused.
	root = fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", validManifest)
	put(t, root, "demo/index.html", "Form")
	target := filepath.Join(root, "_secret")
	if err := os.WriteFile(target, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "flea", "apps", "demo", "demo")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("symlinked executable accepted")
	}
}

// TestAppPackageRejectsUnsafeArtifacts proves the reserved app.json name, a
// wrong native format, and hardlinks on any declared artifact or on the manifest
// itself all fail closed at load.
func TestAppPackageRejectsUnsafeArtifacts(t *testing.T) {
	valid := `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`
	load := func(t *testing.T) error {
		root := fixture(t)
		put(t, root, ".flea.config.json", valid)
		put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
		put(t, root, "demo/index.html", "Form")
		writeExecutable(t, root, "flea/apps/demo/demo")
		_, err := Load(root)
		return err
	}
	if err := load(t); err != nil {
		t.Fatalf("valid package rejected: %v", err)
	}

	// app.json is reserved and is never a valid executable filename.
	root := fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "app.json"))
	put(t, root, "demo/index.html", "Form")
	if _, err := Load(root); err == nil {
		t.Fatal("reserved app.json accepted as an executable")
	}

	// A declaration whose bytes are not the platform's native format is refused.
	root = fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	put(t, root, "demo/index.html", "Form")
	if err := os.WriteFile(filepath.Join(root, "flea", "apps", "demo", "demo"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("non-native executable accepted")
	}

	// A hardlink on a non-selected declared artifact is still refused.
	root = fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", fmt.Sprintf(`{"name":"demo","version":"1.0.0","protocol":2,"executable":{%q:"demo",%q:"demo-plan9"}}`, testPlatform, "plan9-386"))
	put(t, root, "demo/index.html", "Form")
	writeExecutable(t, root, "flea/apps/demo/demo")
	if err := os.WriteFile(filepath.Join(root, "flea", "apps", "demo", "demo-plan9"), []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(filepath.Join(root, "flea", "apps", "demo", "demo-plan9"), filepath.Join(root, "_linked")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("hardlinked non-selected executable accepted")
	}

	// A hardlinked manifest is refused.
	root = fixture(t)
	put(t, root, ".flea.config.json", valid)
	put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	put(t, root, "demo/index.html", "Form")
	writeExecutable(t, root, "flea/apps/demo/demo")
	if err := os.Link(filepath.Join(root, "flea", "apps", "demo", "app.json"), filepath.Join(root, "_manifest-link")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("hardlinked manifest accepted")
	}
}

// TestAppStorageIsStableRuntimeState proves custom-app storage is derived from
// the managed runtime root beside content, never from the immutable release
// directory, and that two releases share one storage location.
func TestAppStorageIsStableRuntimeState(t *testing.T) {
	releaseA := fixture(t)
	releaseB := fixture(t)
	for _, root := range []string{releaseA, releaseB} {
		put(t, root, ".flea.config.json", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`)
		put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
		put(t, root, "demo/index.html", "Form")
		writeExecutable(t, root, "flea/apps/demo/demo")
	}
	contentPath := filepath.Join(t.TempDir(), "sites", "demo", "content", "rev.json")
	runtimeRoot := ManagedRuntimeRoot(contentPath)
	want := filepath.Join(runtimeRoot, "apps", "demo")
	for _, root := range []string{releaseA, releaseB} {
		p, err := LoadWithMedia(root, nil, ManagedMediaRoot(contentPath), runtimeRoot)
		if err != nil {
			t.Fatal(err)
		}
		pkg := p.appPackages["demo"]
		if pkg.Storage != want {
			t.Fatalf("storage %q, want %q", pkg.Storage, want)
		}
		if strings.HasPrefix(pkg.Storage, root) {
			t.Fatalf("storage inside the immutable release: %q", pkg.Storage)
		}
	}
	// Development keeps storage inside the writable site tree.
	dev, err := Load(releaseA)
	if err != nil {
		t.Fatal(err)
	}
	if got := dev.appPackages["demo"].Storage; got != filepath.Join(releaseA, "flea", "runtime", "apps", "demo") {
		t.Fatalf("dev storage %q", got)
	}
}

// TestChangedPackageFailsClosed proves a reload that changes the package
// identity makes the app unavailable until restart instead of reusing the old
// process.
func TestChangedPackageFailsClosed(t *testing.T) {
	root := fixture(t)
	put(t, root, ".flea.config.json", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`)
	put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	put(t, root, "demo/index.html", "Form")
	writeExecutable(t, root, "flea/apps/demo/demo")
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	s.apps.Lock()
	s.apps.running = map[string]*runningApp{"demo": {mount: appapi.Mount{Name: "demo", Path: "/demo"}, config: p.Apps["demo"], pkg: p.appPackages["demo"]}}
	s.apps.Unlock()
	if _, running, _ := s.appFor(p, "/demo"); running == nil {
		t.Fatal("matching package marked unavailable")
	}
	changed := p.appPackages["demo"]
	changed.Hash[0] ^= 0xff
	p.appPackages["demo"] = changed
	if _, running, _ := s.appFor(p, "/demo"); running != nil {
		t.Fatal("changed package reused its running process")
	}
}

// TestAppPackageHardlinkRejected proves a hardlinked executable, which could be
// swapped through its other name after validation, fails closed at load.
func TestAppPackageHardlinkRejected(t *testing.T) {
	root := fixture(t)
	put(t, root, ".flea.config.json", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`)
	put(t, root, "flea/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	put(t, root, "demo/index.html", "Form")
	writeExecutable(t, root, "flea/apps/demo/demo")
	if err := os.Link(filepath.Join(root, "flea", "apps", "demo", "demo"), filepath.Join(root, "_second-link")); err != nil {
		t.Skipf("hardlinks unavailable: %v", err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("hardlinked executable accepted")
	}
}
