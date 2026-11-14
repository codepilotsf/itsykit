package serverbundle

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestExtract(t *testing.T) {
	for _, mode := range []string{"amd64", "arm64", "wrong-arch", "checksum", "mismatch", "missing", "missing-checksum", "bad-gzip", "truncated", "not-elf", "dynamic", "cancelled", "existing", "unsupported"} {
		t.Run(mode, func(t *testing.T) {
			arch := "amd64"
			data := make([]byte, 64)
			copy(data, []byte{0x7f, 'E', 'L', 'F', 2, 1, 1})
			binary.LittleEndian.PutUint16(data[16:], 2)
			binary.LittleEndian.PutUint16(data[18:], 62)
			binary.LittleEndian.PutUint32(data[20:], 1)
			binary.LittleEndian.PutUint16(data[52:], 64)
			if mode == "arm64" || mode == "wrong-arch" {
				arch = "arm64"
			}
			if mode == "arm64" {
				binary.LittleEndian.PutUint16(data[18:], 183)
			}
			if mode == "dynamic" {
				binary.LittleEndian.PutUint64(data[32:], 64)
				binary.LittleEndian.PutUint16(data[54:], 56)
				binary.LittleEndian.PutUint16(data[56:], 1)
				data = append(data, make([]byte, 56)...)
				binary.LittleEndian.PutUint32(data[64:], 3) // PT_INTERP
			}
			if mode == "not-elf" {
				data = []byte("not executable")
			}
			var compressed bytes.Buffer
			z := gzip.NewWriter(&compressed)
			_, _ = z.Write(data)
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			body := compressed.Bytes()
			sum := fmt.Sprintf("%x", sha256.Sum256(data))
			if mode == "checksum" {
				sum = "invalid"
			}
			if mode == "mismatch" {
				sum = strings.Repeat("0", 64)
			}
			if mode == "bad-gzip" {
				body = []byte("bad")
			}
			if mode == "truncated" {
				body = body[:len(body)-5]
			}
			name := "assets/itsy-server-linux-" + arch
			source := fstest.MapFS{name + ".gz": &fstest.MapFile{Data: body}, name + ".sha256": &fstest.MapFile{Data: []byte(sum)}}
			if mode == "missing" {
				delete(source, name+".gz")
			}
			if mode == "missing-checksum" {
				delete(source, name+".sha256")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "cancelled" {
				cancel()
			}
			if mode == "unsupported" {
				arch = "../amd64"
			}
			destination := filepath.Join(t.TempDir(), "server")
			if mode == "existing" {
				if err := os.WriteFile(destination, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := extract(ctx, source, arch, destination)
			if mode == "amd64" || mode == "arm64" {
				if err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(destination)
				if err != nil || !bytes.Equal(data, got) {
					t.Fatal("incorrect extraction", err)
				}
				return
			}
			if err == nil {
				t.Fatal("accepted invalid bundle")
			}
			if mode == "existing" {
				got, _ := os.ReadFile(destination)
				if string(got) != "keep" {
					t.Fatal("overwrote existing file")
				}
			} else if _, err := os.Stat(destination); !os.IsNotExist(err) {
				t.Fatal("left failed extraction", err)
			}
		})
	}
}

func TestEmbeddedServers(t *testing.T) {
	// Source-only checkouts have a placeholder; packaged builds require both.
	if _, err := fs.Stat(assets, "assets/itsy-server-linux-amd64.gz"); err != nil {
		if os.Getenv("FLEA_REQUIRE_BUNDLE") == "1" {
			t.Fatal("missing bundled servers")
		}
		t.Skip("run scripts/build.sh to generate bundled servers")
	}
	for _, arch := range []string{"amd64", "arm64"} {
		if err := Extract(context.Background(), arch, filepath.Join(t.TempDir(), "server")); err != nil {
			t.Fatal(arch, err)
		}
	}
}
