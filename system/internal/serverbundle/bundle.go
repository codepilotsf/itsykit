// Package serverbundle holds the Linux servers shipped inside the CLI.
// The server entry point must never import this package.
package serverbundle

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"debug/elf"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

//go:embed assets/*
var assets embed.FS

// Extract writes a verified static executable to a new private staging file.
// It has no network or compiler fallback, including in development builds.
func Extract(ctx context.Context, arch, destination string) error {
	return extract(ctx, assets, arch, destination)
}

func extract(ctx context.Context, source fs.FS, arch, destination string) error {
	if arch != "amd64" && arch != "arm64" {
		return fmt.Errorf("Unsupported server architecture: %s", arch)
	}
	name := "assets/itsy-server-linux-" + arch
	sum, err := fs.ReadFile(source, name+".sha256")
	if err != nil {
		return fmt.Errorf("This ItsyKit executable is missing its bundled server.\n\nRebuild from the ItsyKit source with: bash scripts/build.sh\nOr install a complete ItsyKit release.")
	}
	expected := strings.TrimSpace(string(sum))
	decoded, err := hex.DecodeString(expected)
	if err != nil || len(decoded) != sha256.Size {
		return fmt.Errorf("The bundled server checksum is invalid. Rebuild or reinstall ItsyKit.")
	}
	compressed, err := source.Open(name + ".gz")
	if err != nil {
		return fmt.Errorf("The bundled server is missing. Rebuild or reinstall ItsyKit.")
	}
	defer compressed.Close()
	z, err := gzip.NewReader(compressed)
	if err != nil {
		return fmt.Errorf("The bundled server is damaged. Rebuild or reinstall ItsyKit.")
	}
	defer z.Close()
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		f.Close()
		if !keep {
			os.Remove(destination)
		}
	}()
	const limit = 100 << 20
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, hash), io.LimitReader(contextReader{ctx, z}, limit+1))
	if err != nil {
		return fmt.Errorf("Could not unpack the bundled server: %w", err)
	}
	if n > limit || hex.EncodeToString(hash.Sum(nil)) != strings.ToLower(expected) {
		return fmt.Errorf("The bundled server failed its integrity check. Rebuild or reinstall ItsyKit.")
	}
	if err := f.Close(); err != nil {
		return err
	}
	binary, err := elf.Open(destination)
	if err != nil {
		return fmt.Errorf("The bundled server is not a Linux executable.")
	}
	defer binary.Close()
	machine := elf.EM_X86_64
	if arch == "arm64" {
		machine = elf.EM_AARCH64
	}
	if binary.Machine != machine || binary.Class != elf.ELFCLASS64 || binary.Data != elf.ELFDATA2LSB || binary.Type != elf.ET_EXEC {
		return fmt.Errorf("The bundled server does not match the target architecture.")
	}
	for _, p := range binary.Progs {
		if p.Type == elf.PT_INTERP || p.Type == elf.PT_DYNAMIC {
			return fmt.Errorf("The bundled server is not a static executable.")
		}
	}
	keep = true
	return nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
