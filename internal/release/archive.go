// Package release packages only site runtime files and extracts bounded archives
// into fresh release directories. It never accepts links or executable modes.
package release

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"nanoweb/internal/sitefiles"
	"os"
	"path"
)

const MaxBytes int64 = 100 << 20
const MaxFiles = 10000

func allowed(name string) bool {
	return name != "." && sitefiles.Included(name)
}

// Included reports whether a committed path belongs in a deployment snapshot.
func Included(name string) bool { return allowed(name) }

func Pack(directory string, output io.Writer) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	gz := gzip.NewWriter(output)
	tw := tar.NewWriter(gz)
	defer gz.Close()
	defer tw.Close()
	var total int64
	count := 0
	err = fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		if !allowed(name) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in release: %s", name)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported release file: %s", name)
		}
		count++
		if count > MaxFiles {
			return fmt.Errorf("release exceeds %d entries", MaxFiles)
		}
		hdr := &tar.Header{Name: name, Mode: 0644, Size: info.Size(), Typeflag: tar.TypeReg}
		if info.IsDir() {
			hdr.Typeflag = tar.TypeDir
			hdr.Mode = 0755
			hdr.Size = 0
		}
		total += hdr.Size
		if total > MaxBytes {
			return fmt.Errorf("release exceeds 100 MiB")
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := root.Open(name)
		if err != nil {
			return err
		}
		defer file.Close()
		actual, err := file.Stat()
		if err != nil {
			return err
		}
		if !os.SameFile(info, actual) || !actual.Mode().IsRegular() {
			return fmt.Errorf("file changed while packaging: %s", name)
		}
		_, err = io.CopyN(tw, file, hdr.Size)
		return err
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// Extract requires an empty, private destination owned by the caller.
func Extract(input io.Reader, directory string) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("release destination must be empty")
	}
	gz, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	var total int64
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := h.Name
		if !allowed(name) || seen[name] || len(seen) >= MaxFiles || h.Size < 0 || h.Size > MaxBytes-total {
			return fmt.Errorf("invalid, duplicate, or oversized archive entry %q", name)
		}
		if h.Typeflag != tar.TypeDir && h.Typeflag != tar.TypeReg {
			return fmt.Errorf("unsupported archive entry %q", name)
		}
		if h.Typeflag == tar.TypeDir && h.Size != 0 {
			return fmt.Errorf("directory contains data")
		}
		seen[name] = true
		total += h.Size
		if h.Typeflag == tar.TypeDir {
			if err := root.MkdirAll(name, 0755); err != nil {
				return err
			}
			continue
		}
		if err := root.MkdirAll(path.Dir(name), 0755); err != nil {
			return err
		}
		f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyN(f, tr, h.Size)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	// Consume the gzip footer too so truncated/corrupt uploads cannot activate.
	if n, err := io.Copy(io.Discard, io.LimitReader(gz, 1)); err != nil {
		return err
	} else if n != 0 {
		return fmt.Errorf("unexpected data after tar archive")
	}
	return nil
}
