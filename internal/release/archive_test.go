package release

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveRoundTripAndScope(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"index.html": "hello", "_/flea.config": `{"site":{"name":"test"}}`, "_/fleas/home.json": "{}", "_/deploy.json": "private-target", ".git/config": "secret", "README.md": "local"} {
		file := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var b bytes.Buffer
	if err := Pack(root, &b); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Extract(bytes.NewReader(b.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(dest, "index.html")); err != nil || string(data) != "hello" {
		t.Fatal(string(data), err)
	}
	for _, name := range []string{"_/deploy.json", ".git/config", "README.md"} {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatal("included local file", name)
		}
	}
	if err := Extract(bytes.NewReader(b.Bytes()), dest); err == nil {
		t.Fatal("overwrote populated release")
	}
	if err := os.Symlink("index.html", filepath.Join(root, "link")); err != nil {
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
		{"../escape", tar.TypeReg, 0}, {"../../escape", tar.TypeReg, 0}, {"/tmp/escape", tar.TypeReg, 0}, {"./index", tar.TypeReg, 0}, {"link", tar.TypeSymlink, 0}, {"hardlink", tar.TypeLink, 0}, {"fifo", tar.TypeFifo, 0}, {"_/deploy.json", tar.TypeReg, 0}, {"huge", tar.TypeReg, MaxBytes + 1}, {"html\\escape", tar.TypeReg, 0},
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
			if err := tw.WriteHeader(&tar.Header{Name: "index.html", Mode: 0644, Typeflag: tar.TypeReg}); err != nil {
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
		tw.WriteHeader(&tar.Header{Name: "css", Typeflag: tar.TypeDir, Mode: 0755})
		tw.Close()
		gz.Close()
		if err := Extract(bytes.NewReader(b.Bytes()[:b.Len()-5]), t.TempDir()); err == nil {
			t.Fatal("accepted truncated gzip")
		}
	})
}
