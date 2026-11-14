// Package release packages only site runtime files and extracts bounded archives
// into fresh release directories. It never accepts links, and admits executable
// modes only for a confined custom-app package file (itsy/apps/<name>/<exe>).
package release

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"itsy/internal/sitefiles"
	"os"
	"path"
	"strings"
)

const MaxBytes int64 = 256 << 20
const MaxFiles = 10000

// allowed includes ordinary runtime files and the static candidate paths of a
// custom-app package. The packer later narrows app packages to the manifest's
// exact declared artifacts.
func allowed(name string) bool {
	return name != "." && (sitefiles.Included(name) || sitefiles.AppPackageFile(name))
}

// Included reports whether a committed path belongs in a deployment snapshot.
func Included(name string) bool { return allowed(name) }

// readBounded opens a confined file without following symlinks and reads it up
// to a hard limit. It is used for package manifests, never request paths.
func readBounded(root *os.Root, name string, limit int64) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", name)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil || !os.SameFile(info, actual) || !actual.Mode().IsRegular() {
		return nil, fmt.Errorf("file changed while reading: %s", name)
	}
	if actual.Size() > limit {
		return nil, fmt.Errorf("%s exceeds its size limit", name)
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s exceeds its size limit", name)
	}
	return data, nil
}

// CheckAppPlatforms verifies that every custom app package in root declares and
// actually contains a native executable for the deployment target. It is the
// portable preflight that stops a package built only for the development
// platform, or one holding a mislabeled file, from being uploaded.
func CheckAppPlatforms(root, goos, goarch string) error {
	r, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer r.Close()
	entries, err := fs.ReadDir(r.FS(), sitefiles.Resolve(r, "itsy/apps"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !sitefiles.AppName.MatchString(entry.Name()) {
			continue
		}
		name := entry.Name()
		base := sitefiles.Resolve(r, "itsy/apps/"+name)
		raw, err := readBounded(r, base+"/app.json", sitefiles.MaxManifestBytes)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("app %q manifest: %w", name, err)
		}
		manifest, err := sitefiles.DecodeAppManifest(raw)
		if err != nil {
			return fmt.Errorf("app %q manifest is invalid", name)
		}
		executables, err := sitefiles.AppExecutables(name, manifest)
		if err != nil {
			return err
		}
		manifestInfo, err := r.Lstat(base + "/app.json")
		if err != nil || !manifestInfo.Mode().IsRegular() || sitefiles.Hardlinked(manifestInfo) {
			return fmt.Errorf("app %q app.json must be a single-link regular file", name)
		}
		for executable := range executables {
			info, err := r.Lstat(base + "/" + executable)
			if err != nil || !info.Mode().IsRegular() || sitefiles.Hardlinked(info) {
				return fmt.Errorf("app %q executable %q is missing or unsafe", name, executable)
			}
		}
		target, ok := manifest.Executable[goos+"-"+goarch]
		if !ok {
			return fmt.Errorf("app %q has no %s-%s executable; build it before deploying", name, goos, goarch)
		}
		info, err := r.Lstat(base + "/" + target)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || sitefiles.Hardlinked(info) {
			return fmt.Errorf("app %q %s-%s executable is missing or not runnable", name, goos, goarch)
		}
		f, err := r.Open(base + "/" + target)
		if err != nil {
			return fmt.Errorf("app %q %s-%s executable is missing", name, goos, goarch)
		}
		header := make([]byte, 1<<12)
		n, readErr := io.ReadFull(f, header)
		f.Close()
		if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
			return fmt.Errorf("app %q %s-%s executable could not be read", name, goos, goarch)
		}
		if err := sitefiles.CheckNativeExecutable(header[:n], goos, goarch); err != nil {
			return fmt.Errorf("app %q %s-%s executable is not a native binary: %w", name, goos, goarch, err)
		}
	}
	return nil
}

// packagedAppFiles reports the exact package files to ship: each valid manifest
// and every platform executable it declares. Unlisted files stay local. A
// manifest that exists but is malformed fails the release rather than shipping
// half a package; a directory with no manifest is ignored.
func packagedAppFiles(root *os.Root) (map[string]bool, error) {
	files := map[string]bool{}
	entries, err := fs.ReadDir(root.FS(), sitefiles.Resolve(root, "itsy/apps"))
	if err != nil {
		if os.IsNotExist(err) {
			return files, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || !sitefiles.AppName.MatchString(name) {
			continue
		}
		base := sitefiles.Resolve(root, "itsy/apps/"+name)
		raw, err := readBounded(root, base+"/app.json", sitefiles.MaxManifestBytes)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("app %q manifest: %w", name, err)
		}
		manifest, err := sitefiles.DecodeAppManifest(raw)
		if err != nil {
			return nil, fmt.Errorf("app %q manifest is invalid", name)
		}
		executables, err := sitefiles.AppExecutables(name, manifest)
		if err != nil {
			return nil, err
		}
		manifestInfo, err := root.Lstat(base + "/app.json")
		if err != nil || !manifestInfo.Mode().IsRegular() || sitefiles.Hardlinked(manifestInfo) {
			return nil, fmt.Errorf("app %q app.json must be a single-link regular file", name)
		}
		for executable := range executables {
			target := base + "/" + executable
			info, err := root.Lstat(target)
			if err != nil || !info.Mode().IsRegular() || sitefiles.Hardlinked(info) {
				return nil, fmt.Errorf("app %q executable %q is missing or unsafe", name, executable)
			}
			files[target] = true
		}
		files[base+"/app.json"] = true
	}
	return files, nil
}

// validatePackages checks the extracted tree contains exactly each manifest and
// its declared executables, all runnable regular files.
func validatePackages(root *os.Root) error {
	entries, err := fs.ReadDir(root.FS(), sitefiles.Resolve(root, "itsy/apps"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !sitefiles.AppName.MatchString(entry.Name()) {
			continue
		}
		name := entry.Name()
		base := sitefiles.Resolve(root, "itsy/apps/"+name)
		raw, err := readBounded(root, base+"/app.json", sitefiles.MaxManifestBytes)
		if err != nil {
			return fmt.Errorf("release app %q is missing a valid manifest", name)
		}
		manifest, err := sitefiles.DecodeAppManifest(raw)
		if err != nil {
			return fmt.Errorf("release app %q manifest is invalid", name)
		}
		executables, err := sitefiles.AppExecutables(name, manifest)
		if err != nil {
			return err
		}
		allowed := map[string]bool{"app.json": true}
		for executable := range executables {
			allowed[executable] = true
		}
		dirEntries, err := fs.ReadDir(root.FS(), base)
		if err != nil {
			return err
		}
		for _, f := range dirEntries {
			if f.IsDir() || !allowed[f.Name()] {
				return fmt.Errorf("release app %q contains an unlisted entry %q", name, f.Name())
			}
		}
		manifestInfo, err := root.Lstat(base + "/app.json")
		if err != nil || !manifestInfo.Mode().IsRegular() || sitefiles.Hardlinked(manifestInfo) {
			return fmt.Errorf("release app %q app.json must be a single-link regular file", name)
		}
		for executable := range executables {
			info, err := root.Lstat(base + "/" + executable)
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || sitefiles.Hardlinked(info) {
				return fmt.Errorf("release app %q executable %q is missing or not runnable", name, executable)
			}
		}
	}
	return nil
}

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
	packaged, err := packagedAppFiles(root)
	if err != nil {
		return err
	}
	err = fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		if !allowed(name) {
			if d.IsDir() && sitefiles.AppPackageDir(name) {
				return nil
			}
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		// Ship only the manifest and declared executable of a custom app package.
		if sitefiles.AppPackageFile(name) && !packaged[name] {
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
		} else if sitefiles.AppExecutablePath(name) {
			// A custom app executable is confined, checksummed site code. It is
			// packaged read-only and executable for the service user, never
			// writable.
			hdr.Mode = int64(sitefiles.ExecutableMode)
		}
		var config []byte
		if sitefiles.CanonicalName(name) == "itsy.config.json" && !info.IsDir() {
			if info.Size() > 1<<20 {
				return fmt.Errorf("site config exceeds 1 MiB")
			}
			file, openErr := root.Open(name)
			if openErr != nil {
				return openErr
			}
			actual, statErr := file.Stat()
			if statErr != nil || !os.SameFile(info, actual) || !actual.Mode().IsRegular() {
				file.Close()
				return fmt.Errorf("file changed while packaging: %s", name)
			}
			config, err = io.ReadAll(io.LimitReader(file, (1<<20)+1))
			file.Close()
			if err != nil {
				return err
			}
			if len(config) > 1<<20 {
				return fmt.Errorf("site config exceeds 1 MiB")
			}
			config, err = sitefiles.RuntimeConfig(config)
			if err != nil {
				return err
			}
			hdr.Size = int64(len(config))
		}
		total += hdr.Size
		if total > MaxBytes {
			return fmt.Errorf("release exceeds 256 MiB")
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if config != nil {
			_, err = tw.Write(config)
			return err
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
		mode := os.FileMode(0644)
		if h.Typeflag == tar.TypeReg {
			switch h.Mode {
			case 0644:
			case int64(sitefiles.ExecutableMode):
				// Executable extraction is limited to custom app files.
				if !sitefiles.AppExecutablePath(name) {
					return fmt.Errorf("unexpected executable archive entry %q", name)
				}
				mode = sitefiles.ExecutableMode
			default:
				return fmt.Errorf("unsupported archive entry mode %q", name)
			}
		}
		seen[name] = true
		total += h.Size
		if h.Typeflag == tar.TypeDir {
			if err := makeDirs(root, name, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := makeDirs(root, path.Dir(name), 0o755); err != nil {
			return err
		}
		f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
		if err != nil {
			return err
		}
		// Apply the mode explicitly so a restrictive deploy umask cannot leave
		// the service user unable to read (or execute) a release file.
		if err := f.Chmod(mode); err != nil {
			f.Close()
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
	if err := validatePackages(root); err != nil {
		return err
	}
	return nil
}

// makeDirs creates each missing component under root and sets its mode
// explicitly, so extraction is independent of the caller's umask and the
// service user always has directory traversal.
func makeDirs(root *os.Root, dir string, mode os.FileMode) error {
	if dir == "." || dir == "" {
		return nil
	}
	current := ""
	for _, part := range strings.Split(dir, "/") {
		if part == "" {
			continue
		}
		current = path.Join(current, part)
		if err := root.Mkdir(current, mode); err != nil && !errors.Is(err, fs.ErrExist) {
			return err
		}
		if err := root.Chmod(current, mode); err != nil {
			return err
		}
	}
	return nil
}
