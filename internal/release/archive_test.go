package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveRoundTripAndScope(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"site/index.html": "hello", ".flea.config.json": `{"app":{"site_name":"test"},"admin":{"emails":["owner@example.test"]},"deploy":{"branches":{"main":{"target":"private-target"}}}}`, "flea/data/home.json": "{}", "flea/runtime/sessions.json": "private-session", ".git/config": "secret", "README.md": "local", ".agents/skills/core/SKILL.md": "private agent instructions", ".agents/skills-source.md": "local source provenance"} {
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
	if _, err := os.Stat(filepath.Join(dest, ".flea.config.json")); err != nil {
		t.Fatal("configuration missing from release", err)
	}
	configBytes, err := os.ReadFile(filepath.Join(dest, ".flea.config.json"))
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
	if data, err := os.ReadFile(filepath.Join(dest, "mailers", "sign-in.email.html")); err != nil || string(data) != "{email.code}" {
		t.Fatal("mailer missing from release", err)
	}
	for _, name := range []string{"flea/runtime", ".git/config", "README.md", ".agents"} {
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

func TestRejectUnsafeArchives(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  byte
		size int64
	}{
		{"../escape", tar.TypeReg, 0}, {"../../escape", tar.TypeReg, 0}, {"/tmp/escape", tar.TypeReg, 0}, {"./index", tar.TypeReg, 0}, {"site/link", tar.TypeSymlink, 0}, {"site/hardlink", tar.TypeLink, 0}, {"site/fifo", tar.TypeFifo, 0}, {"flea/runtime/sessions.json", tar.TypeReg, 0}, {"site/huge", tar.TypeReg, MaxBytes + 1}, {"site\\escape", tar.TypeReg, 0},
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
