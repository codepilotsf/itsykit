package site

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"itsy/internal/appapi"
	"itsy/internal/sitefiles"
)

// appPackage is a resolved, confined custom-app package. The manifest describes
// the package; the authenticated handshake remains the compatibility authority.
type appPackage struct {
	Name       string
	Version    string
	Executable string // absolute path to the executable selected for this platform
	Dir        string // absolute package directory
	Storage    string // absolute app-owned private storage directory
	Hash       [32]byte
	Size       int64
}

// headerBytes bounds the native-format probe: enough for an ELF or Mach-O
// header and any fat architecture table.
const headerBytes = 4 << 10

// loadAppPackage reads and validates itsy/apps/<name>. It never follows a
// symlink and keeps every declared executable confined to the package directory.
// The executable for the running platform is selected here; a package without
// one is rejected with the required platform named.
func loadAppPackage(root *os.Root, siteRoot, runtimeRoot, name string) (appPackage, error) {
	var pkg appPackage
	if name != filepath.Base(name) || !appapi.ValidName(name) {
		return pkg, fmt.Errorf("invalid app name")
	}
	base := sitefiles.Resolve(root, "itsy/apps/"+name)
	raw, err := readFileBounded(root, base+"/app.json", sitefiles.MaxManifestBytes)
	if err != nil {
		return pkg, fmt.Errorf("app %q needs a package at itsy/apps/%s with app.json and an executable", name, name)
	}
	manifest, err := sitefiles.DecodeAppManifest(raw)
	if err != nil {
		return pkg, fmt.Errorf("itsy/apps/%s/app.json %w", name, err)
	}
	executables, err := sitefiles.ValidateAppManifest(manifest, name, appapi.Version)
	if err != nil {
		return pkg, err
	}
	if err := scanPackage(root, base); err != nil {
		return pkg, fmt.Errorf("app %q package is unsafe: %w", name, err)
	}
	// The manifest and every declared artifact must be single-link regular files
	// so a non-selected hardlink cannot smuggle private bytes into a release and
	// a swapped file cannot be reused after validation.
	manifestInfo, err := root.Lstat(base + "/app.json")
	if err != nil || !manifestInfo.Mode().IsRegular() || sitefiles.Hardlinked(manifestInfo) {
		return pkg, fmt.Errorf("app %q app.json must be a single-link regular file", name)
	}
	for file := range executables {
		info, err := root.Lstat(base + "/" + file)
		if err != nil || !info.Mode().IsRegular() || sitefiles.Hardlinked(info) {
			return pkg, fmt.Errorf("app %q executable %q is missing or unsafe", name, file)
		}
	}
	platform := runtime.GOOS + "-" + runtime.GOARCH
	selected, ok := manifest.Executable[platform]
	if !ok {
		return pkg, fmt.Errorf("app %q does not provide a %s executable; declare it in app.json", name, platform)
	}
	f, err := openSiteFile(root, base+"/"+selected, sitefiles.MaxAppExecutable)
	if err != nil {
		return pkg, fmt.Errorf("app %q executable is missing or unsafe", name)
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || sitefiles.Hardlinked(info) {
		f.Close()
		return pkg, fmt.Errorf("app %q executable must be a runnable regular file with a single link", name)
	}
	header := make([]byte, headerBytes)
	n, readErr := io.ReadFull(f, header)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		f.Close()
		return pkg, fmt.Errorf("app %q executable could not be read", name)
	}
	if err := sitefiles.CheckNativeExecutable(header[:n], runtime.GOOS, runtime.GOARCH); err != nil {
		f.Close()
		return pkg, fmt.Errorf("app %q executable %q is not a %s binary: %w", name, selected, platform, err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return pkg, fmt.Errorf("app %q executable could not be read", name)
	}
	h := sha256.New()
	size, err := io.Copy(h, io.LimitReader(f, sitefiles.MaxAppExecutable+1))
	f.Close()
	if err != nil || size > sitefiles.MaxAppExecutable {
		return pkg, fmt.Errorf("app %q executable could not be read", name)
	}
	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return appPackage{
		Name:       name,
		Version:    manifest.Version,
		Executable: sitefiles.Path(siteRoot, filepath.Join("itsy", "apps", name, selected)),
		Dir:        sitefiles.Path(siteRoot, filepath.Join("itsy", "apps", name)),
		Storage:    filepath.Join(runtimeRoot, "apps", name),
		Hash:       sum,
		Size:       size,
	}, nil
}

// scanPackage rejects links, irregular files and excessive entries anywhere in a
// package, so nothing outside the confined manifest+executables is admitted.
func scanPackage(root *os.Root, base string) error {
	count := 0
	return fs.WalkDir(root.FS(), base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		count++
		if count > sitefiles.MaxAppFiles {
			return fmt.Errorf("package contains too many entries")
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink is unsupported: %s", path)
		}
		if !d.IsDir() && !d.Type().IsRegular() {
			return fmt.Errorf("unsupported file: %s", path)
		}
		return nil
	})
}
