// Package selfupdate replaces only the local CLI with a verified release.
package selfupdate

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"debug/elf"
	"debug/macho"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const BaseURL = "https://fleacms.com/releases"
const maxBinary = 128 << 20

var stableVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var checksumPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type updater struct {
	base, system, arch, executable string
	client                         *http.Client
	verify                         func(context.Context, string, string) error
}

func Run(ctx context.Context, current string, out io.Writer) error {
	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate this ItsyWeb executable")
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("could not resolve this ItsyWeb executable")
	}
	client := &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if r.URL.Scheme != "https" || len(via) >= 5 {
			return fmt.Errorf("unsafe release redirect")
		}
		return nil
	}}
	u := updater{BaseURL, runtime.GOOS, runtime.GOARCH, path, client, verifyVersion}
	return u.run(ctx, current, out)
}

func (u updater) download(ctx context.Context, path string, dst io.Writer, limit int64) error {
	req, err := http.NewRequestWithContext(ctx, "GET", u.base+"/"+path, nil)
	if err != nil || req.URL.Scheme != "https" {
		return fmt.Errorf("release downloads require HTTPS")
	}
	response, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not download the release; check your connection and try again")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("release download unavailable (HTTP %d); nothing was changed", response.StatusCode)
	}
	if response.ContentLength > limit {
		return fmt.Errorf("release download exceeds its size limit")
	}
	n, err := io.Copy(dst, io.LimitReader(response.Body, limit+1))
	if err != nil {
		return fmt.Errorf("release download was interrupted; nothing was changed")
	}
	if n > limit {
		return fmt.Errorf("release download exceeds its size limit")
	}
	return nil
}

// Stable releases never downgrade a newer local release or prerelease base.
func newer(latest, current string) (bool, error) {
	if len(latest) > 64 || !stableVersion.MatchString(latest) {
		return false, fmt.Errorf("invalid stable release version")
	}
	base := strings.SplitN(strings.SplitN(current, "+", 2)[0], "-", 2)[0]
	if !stableVersion.MatchString(base) {
		return false, fmt.Errorf("this build has an invalid version; reinstall from fleacms.com")
	}
	a, b := strings.Split(latest, "."), strings.Split(base, ".")
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return len(a[i]) > len(b[i]), nil
		}
		if a[i] != b[i] {
			return a[i] > b[i], nil
		}
	}
	return strings.Contains(strings.SplitN(current, "+", 2)[0], "-"), nil
}

func checksum(manifest, name string) (string, error) {
	result := ""
	for _, line := range strings.Split(manifest, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			if result != "" || !checksumPattern.MatchString(fields[0]) {
				return "", fmt.Errorf("invalid release checksum")
			}
			result = fields[0]
		}
	}
	if result == "" {
		return "", fmt.Errorf("release does not include this platform")
	}
	return result, nil
}

func fileHash(f *os.File) (string, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxBinary+1))
	if err != nil || n > maxBinary {
		return "", fmt.Errorf("could not verify executable bytes")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (u updater) run(ctx context.Context, current string, out io.Writer) error {
	if !(u.system == "darwin" && u.arch == "arm64" || u.system == "linux" && u.arch == "amd64") {
		return fmt.Errorf("automatic updates support Apple Silicon macOS and x86-64 Linux")
	}
	root, err := os.OpenRoot(filepath.Dir(u.executable))
	if err != nil {
		return fmt.Errorf("could not open the installation directory")
	}
	defer root.Close()
	// The installer uses the same portable directory lock. Never remove another
	// invocation's lock, including after an interrupted installation.
	const lock = ".flea-install.lock"
	if err := root.Mkdir(lock, 0700); err != nil {
		return fmt.Errorf("installation is busy or not writable; check %s and try again", filepath.Join(filepath.Dir(u.executable), lock))
	}
	defer root.Remove(lock)
	base := filepath.Base(u.executable)
	original, err := root.Lstat(base)
	if err != nil || !original.Mode().IsRegular() || original.Mode()&(os.ModeSetuid|os.ModeSetgid) != 0 {
		return fmt.Errorf("the installed executable is not a regular, unprivileged file")
	}
	old, err := root.Open(base)
	if err != nil {
		return err
	}
	defer old.Close()
	actual, err := old.Stat()
	if err != nil || !os.SameFile(original, actual) {
		return fmt.Errorf("installed executable changed; try again")
	}
	oldHash, err := fileHash(old)
	if err != nil {
		return err
	}
	if base == "flea" {
		if info, err := root.Lstat("itsy"); err == nil {
			if !info.Mode().IsRegular() || !os.SameFile(original, info) {
				return fmt.Errorf("itsy already exists in this directory; use itsy update or choose another installation directory")
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("could not inspect the itsy installation: %w", err)
		}
	}
	var latest, manifest strings.Builder
	if err := u.download(ctx, "latest.txt", &latest, 128); err != nil {
		return err
	}
	version := strings.TrimSpace(latest.String())
	update, err := newer(version, current)
	if err != nil {
		return err
	}
	if !update {
		if err := installItsyName(root, base, base, original, oldHash, original); err != nil {
			return err
		}
		if err := syncInstallation(root); err != nil {
			return err
		}
		fmt.Fprintf(out, "ItsyWeb %s is already up to date.\nUse itsy for all commands from now on.\n", current)
		return nil
	}
	name := "flea-" + u.system + "-" + u.arch
	if err := u.download(ctx, version+"/SHA256SUMS", &manifest, 8192); err != nil {
		return err
	}
	want, err := checksum(manifest.String(), name)
	if err != nil {
		return err
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return err
	}
	temp := ".itsy-update-" + hex.EncodeToString(token[:])
	f, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("installation directory is not writable; reinstall in a user-owned directory")
	}
	defer root.Remove(temp)
	defer f.Close()
	fmt.Fprintf(out, "Downloading ItsyWeb %s…\n", version)
	if err := u.download(ctx, version+"/"+name, f, maxBinary); err != nil {
		return err
	}
	got, err := fileHash(f)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("release checksum did not match; nothing was changed")
	}
	if err := checkPlatform(f, u.system, u.arch); err != nil {
		return err
	}
	if err := f.Chmod(0755); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	candidateInfo, err := f.Stat()
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := u.verify(ctx, filepath.Join(filepath.Dir(u.executable), temp), version); err != nil {
		return err
	}
	candidate, err := root.Open(temp)
	if err != nil {
		return err
	}
	defer candidate.Close()
	candidateNow, err := root.Lstat(temp)
	if err != nil || !os.SameFile(candidateInfo, candidateNow) {
		return fmt.Errorf("downloaded executable changed; nothing was replaced")
	}
	candidateHash, err := fileHash(candidate)
	if err != nil || candidateHash != want {
		return fmt.Errorf("downloaded executable changed; nothing was replaced")
	}
	now, err := root.Lstat(base)
	if err != nil || !os.SameFile(original, now) {
		return fmt.Errorf("installed executable changed during download; nothing was replaced")
	}
	currentHash, err := fileHash(old)
	if err != nil || currentHash != oldHash {
		return fmt.Errorf("installed executable changed during download; nothing was replaced")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if base == "flea" {
		if err := installItsyName(root, base, temp, candidateInfo, want, original); err != nil {
			return err
		}
	} else if err := root.Rename(temp, base); err != nil {
		return fmt.Errorf("could not replace the executable; the previous version is intact")
	}
	if err := syncInstallation(root); err != nil {
		return err
	}
	fmt.Fprintf(out, "Updated to ItsyWeb %s.\n\nRemote servers were not changed. From a site folder, run itsy update-server when ready.\n", version)
	return nil
}

func syncInstallation(root *os.Root) error {
	dir, err := root.Open(".")
	if err == nil {
		err = dir.Sync()
		dir.Close()
	}
	if err != nil {
		return fmt.Errorf("CLI installed, but could not sync the installation directory: %w", err)
	}
	return nil
}

// installItsyName promotes an old regular flea installation without replacing
// an independently installed itsy. A hard link keeps a runnable CLI at both
// names if publishing the retired-command symlink fails partway through.
func installItsyName(root *os.Root, base, source string, expected os.FileInfo, want string, original os.FileInfo) error {
	if base != "flea" {
		return nil
	}
	f, err := root.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !os.SameFile(expected, info) {
		return fmt.Errorf("installed executable changed; run flea update again")
	}
	hash, err := fileHash(f)
	if err != nil || hash != want {
		return fmt.Errorf("installed executable changed; run flea update again")
	}
	if err := root.Link(source, "itsy"); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("could not install the itsy command: %w", err)
		}
		previous, statErr := root.Lstat("itsy")
		if statErr != nil || !previous.Mode().IsRegular() || !os.SameFile(original, previous) {
			return fmt.Errorf("itsy changed during installation; no existing command was replaced")
		}
		if source != base {
			if err := root.Rename(source, "itsy"); err != nil {
				return fmt.Errorf("could not update itsy: %w", err)
			}
		}
	}
	canonical, err := root.Lstat("itsy")
	if err != nil || !canonical.Mode().IsRegular() || !os.SameFile(expected, canonical) {
		return fmt.Errorf("itsy changed during installation; no existing command was replaced")
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return err
	}
	alias := ".itsy-alias-" + hex.EncodeToString(token[:])
	if err := root.Symlink("itsy", alias); err != nil {
		return fmt.Errorf("itsy is installed, but could not prepare the retired flea command: %w", err)
	}
	defer root.Remove(alias)
	now, err := root.Lstat(base)
	if err != nil || !os.SameFile(original, now) {
		return fmt.Errorf("itsy is installed, but flea changed during installation; it was not replaced")
	}
	if err := root.Rename(alias, base); err != nil {
		return fmt.Errorf("itsy is installed, but could not replace the retired flea command: %w", err)
	}
	return nil
}

func checkPlatform(f *os.File, system, arch string) error {
	if system == "linux" {
		binary, err := elf.NewFile(f)
		if err == nil && binary.Type == elf.ET_EXEC && binary.Class == elf.ELFCLASS64 && (arch == "amd64" && binary.Machine == elf.EM_X86_64 || arch == "arm64" && binary.Machine == elf.EM_AARCH64) {
			return nil
		}
	} else {
		binary, err := macho.NewFile(f)
		if err == nil && binary.Type == macho.TypeExec && (arch == "amd64" && binary.Cpu == macho.CpuAmd64 || arch == "arm64" && binary.Cpu == macho.CpuArm64) {
			return nil
		}
	}
	return fmt.Errorf("release executable does not match this platform; nothing was changed")
}

type versionOutput struct{ strings.Builder }

func (w *versionOutput) Write(b []byte) (int, error) {
	if len(b) > 1024-w.Len() {
		return 0, fmt.Errorf("unexpected version output")
	}
	return w.Builder.Write(b)
}
func verifyVersion(ctx context.Context, path, version string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "version")
	cmd.WaitDelay = time.Second
	var output versionOutput
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil || strings.TrimSpace(output.String()) != "ItsyWeb "+version {
		return fmt.Errorf("downloaded CLI could not verify its version; nothing was changed")
	}
	return nil
}
