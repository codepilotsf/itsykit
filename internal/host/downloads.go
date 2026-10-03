package host

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Downloads are administrator-published immutable files, independent of site
// snapshots. Only this validated catalogue is addressable over HTTP.
const maxDownloadVersions = 128
const maxDownloadBytes int64 = 128 << 20

var downloadVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var downloadName = regexp.MustCompile(`^flea-(darwin|linux)-(amd64|arm64)$`)
var downloadHash = regexp.MustCompile(`^[0-9a-f]{64}$`)

type downloadFile struct {
	file *os.File
	info os.FileInfo
	data []byte
}

type downloads struct{ files map[string]downloadFile }

// Administrators must not change file bytes in place. Pin handles during load,
// hash with bounded streaming, and use independent readers for concurrent ranges.
func loadDownloads(directory string) (_ *downloads, err error) {
	if !filepath.IsAbs(directory) || filepath.Clean(directory) != directory || directory == "/" {
		return nil, fmt.Errorf("download root must be a clean absolute directory")
	}
	for p := directory; p != "/"; p = filepath.Dir(p) {
		info, e := os.Lstat(p)
		if e != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("download root contains a linked or invalid directory")
		}
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	rootInfo, err := root.Stat(".")
	if err != nil || rootInfo.Mode().Perm()&0022 != 0 {
		return nil, fmt.Errorf("download root must not be group/world writable")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	if len(entries) > maxDownloadVersions {
		return nil, fmt.Errorf("too many download versions")
	}
	d := &downloads{files: map[string]downloadFile{}}
	defer func() {
		if err != nil {
			for _, f := range d.files {
				if f.file != nil {
					f.file.Close()
				}
			}
		}
	}()
	for _, entry := range entries {
		version := entry.Name()
		if !downloadVersion.MatchString(version) || !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("invalid download version directory %q", version)
		}
		versionInfo, e := root.Lstat(version)
		if e != nil || !versionInfo.IsDir() || versionInfo.Mode().Perm()&0022 != 0 {
			return nil, fmt.Errorf("download version must be a non-writable directory")
		}
		manifest, info, e := openDownload(root, version+"/SHA256SUMS", 8192)
		if e != nil {
			return nil, e
		}
		data, e := io.ReadAll(io.LimitReader(manifest, 8193))
		manifest.Close()
		if e != nil || int64(len(data)) != info.Size() {
			return nil, fmt.Errorf("download manifest changed")
		}
		seen := map[string]bool{}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 || !downloadHash.MatchString(fields[0]) || !downloadName.MatchString(fields[1]) || seen[fields[1]] {
				return nil, fmt.Errorf("invalid download manifest for %s", version)
			}
			name := fields[1]
			seen[name] = true
			file, stat, e := openDownload(root, version+"/"+name, maxDownloadBytes)
			if e != nil {
				return nil, e
			}
			h := sha256.New()
			n, e := io.Copy(h, io.NewSectionReader(file, 0, stat.Size()))
			if e != nil || n != stat.Size() || hex.EncodeToString(h.Sum(nil)) != fields[0] {
				file.Close()
				return nil, fmt.Errorf("download checksum mismatch: %s/%s", version, name)
			}
			d.files["/releases/"+version+"/"+name] = downloadFile{file: file, info: stat}
		}
		if len(seen) == 0 {
			return nil, fmt.Errorf("empty download manifest")
		}
		children, e := root.Open(version)
		if e != nil {
			return nil, e
		}
		names, e := children.Readdirnames(-1)
		children.Close()
		if e != nil {
			return nil, e
		}
		for _, name := range names {
			if name != "SHA256SUMS" && !seen[name] {
				return nil, fmt.Errorf("unlisted download file: %s/%s", version, name)
			}
		}
		d.files["/releases/"+version+"/SHA256SUMS"] = downloadFile{data: data}
	}
	return d, nil
}

func openDownload(root *os.Root, name string, limit int64) (*os.File, os.FileInfo, error) {
	before, err := root.Lstat(name)
	if err != nil {
		return nil, nil, err
	}
	if !before.Mode().IsRegular() || before.Size() <= 0 || before.Size() > limit || before.Mode().Perm()&0022 != 0 {
		return nil, nil, fmt.Errorf("download must be a bounded, non-writable regular file: %s", name)
	}
	if s, ok := before.Sys().(*syscall.Stat_t); ok && s.Nlink != 1 {
		return nil, nil, fmt.Errorf("linked download: %s", name)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, nil, err
	}
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || before.ModTime() != after.ModTime() {
		f.Close()
		return nil, nil, fmt.Errorf("download changed while opening: %s", name)
	}
	return f, after, nil
}

func (d *downloads) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	f, ok := d.files[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	var reader io.ReadSeeker
	if f.file != nil {
		now, err := f.file.Stat()
		if err != nil || now.Size() != f.info.Size() || now.ModTime() != f.info.ModTime() {
			http.Error(w, "Download unavailable", http.StatusServiceUnavailable)
			return
		}
		reader = io.NewSectionReader(f.file, 0, f.info.Size())
		w.Header().Set("Content-Type", "application/octet-stream")
	} else {
		reader = bytes.NewReader(f.data)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, filepath.Base(r.URL.Path), time.Time{}, reader)
}
