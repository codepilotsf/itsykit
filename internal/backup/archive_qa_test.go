package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQABackupRejectsMaliciousTarHeaders(t *testing.T) {
	for _, header := range []*tar.Header{
		{Name: "../escape", Typeflag: tar.TypeReg, Mode: 0644},
		{Name: "index.html", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"},
		{Name: "index.html", Typeflag: tar.TypeLink, Linkname: ".flea.config.json"},
		{Name: "index.html", Typeflag: tar.TypeFifo, Mode: 0644},
	} {
		t.Run(header.Name+string(header.Typeflag), func(t *testing.T) {
			var payload bytes.Buffer
			gz := gzip.NewWriter(&payload)
			tw := tar.NewWriter(gz)
			if err := tw.WriteHeader(header); err != nil {
				t.Fatal(err)
			}
			if err := tw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(payload.Bytes())
			manifest := Manifest{Name: "demo", Branch: "main", Commit: strings.Repeat("a", 40), Size: int64(payload.Len()), SHA256: hex.EncodeToString(digest[:])}
			var input bytes.Buffer
			input.WriteString(magic)
			if err := json.NewEncoder(&input).Encode(manifest); err != nil {
				t.Fatal(err)
			}
			input.Write(payload.Bytes())
			if _, err := Extract(&input, t.TempDir()); err == nil {
				t.Fatal("accepted dangerous tar header")
			}
		})
	}
}

func TestQABackupBoundsIntegrityAndPrivateExclusions(t *testing.T) {
	root := t.TempDir()
	for name, value := range map[string]string{"site/index.html": "<p>public</p>", ".env": "secret", "flea/runtime/sessions.json": "session", ".flea.config.json": `{"app":{"site_name":"Example"},"deploy":{"branches":{}}}`} {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := Write(root, Manifest{Name: "demo", Branch: "main", Commit: strings.Repeat("a", 40)}, &out); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if _, err := Extract(bytes.NewReader(out.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".env", "flea/runtime/sessions.json"} {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatal("private file included", name, err)
		}
	}
	good := out.Bytes()
	corrupt := bytes.Clone(good)
	corrupt[len(corrupt)-12] ^= 1
	for label, data := range map[string][]byte{"corrupt": corrupt, "truncated": good[:len(good)-1], "trailing": append(bytes.Clone(good), 'x'), "bad magic": []byte("NOT-BACKUP\n"), "huge header": []byte(magic + strings.Repeat("x", 5000) + "\n"), "bad identity": []byte(magic + `{"name":"../escape","branch":"main","commit":"` + strings.Repeat("a", 40) + `","size":1,"sha256":"` + strings.Repeat("0", 64) + `"}` + "\nx")} {
		t.Run(label, func(t *testing.T) {
			if _, err := Extract(bytes.NewReader(data), t.TempDir()); err == nil {
				t.Fatal("accepted invalid backup")
			}
		})
	}
}
