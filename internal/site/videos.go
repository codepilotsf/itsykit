package site

import (
	"crypto/sha256"
	"fmt"
	"io"
	"itsy/internal/sitefiles"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"itsy/internal/snips"
	"itsy/internal/video"
)

// Snapshot holds an already-open, checksum-validated file, never a URL-derived
// path. os.File's cleanup closes it once old snapshots and readers are collected.
type videoAsset struct {
	file *os.File
	info os.FileInfo
	name string
}

func OpenVideo(mediaRoot, name string) (*os.File, error) {
	if !video.Filename.MatchString(name) {
		return nil, fmt.Errorf("invalid video filename")
	}
	info, e := os.Lstat(mediaRoot)
	if e != nil {
		return nil, e
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("invalid media directory")
	}
	r, e := os.OpenRoot(mediaRoot)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	d, e := pinnedDir(r, ".")
	if e != nil {
		return nil, e
	}
	defer d.Close()
	before, e := d.Lstat(name)
	if e != nil {
		return nil, e
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("video must be a regular file")
	}
	if stat, ok := before.Sys().(*syscall.Stat_t); ok && stat.Nlink != 1 {
		return nil, fmt.Errorf("hard-linked videos are unsupported")
	}
	f, e := d.Open(name)
	if e != nil {
		return nil, e
	}
	after, e := f.Stat()
	if e != nil || !os.SameFile(before, after) || after.Size() > video.MaxFile {
		f.Close()
		return nil, fmt.Errorf("invalid or changed video file")
	}
	return f, nil
}

func CheckVideo(f *os.File, v video.Video) error {
	info, e := f.Stat()
	if e != nil {
		return e
	}
	meta, e := video.Inspect(f, info.Size())
	if e != nil {
		return e
	}
	if meta != v.Metadata {
		return fmt.Errorf("video metadata/checksum mismatch")
	}
	return nil
}

func (p *Project) loadVideos() error {
	p.videos = map[string]videoAsset{}
	success := false
	defer func() {
		if !success {
			for _, v := range p.videos {
				v.file.Close()
			}
		}
	}()
	var total int64
	for _, v := range p.Content.VideoAssets() {
		route := "/media/" + v.File
		if prior, ok := p.videos[route]; ok {
			if err := CheckVideo(prior.file, v); err != nil {
				return err
			}
			continue
		}
		f, e := OpenVideo(p.MediaRoot, v.File)
		if e != nil {
			return e
		}
		if e = CheckVideo(f, v); e != nil {
			f.Close()
			return e
		}
		info, e := f.Stat()
		if e != nil {
			f.Close()
			return e
		}
		total += info.Size()
		if total > video.MaxTotal {
			f.Close()
			return fmt.Errorf("video library exceeds 64 MiB")
		}
		p.videos[route] = videoAsset{f, info, v.File}
		p.Assets[route] = "itsy/media/" + v.File
	}
	success = true
	return nil
}
func (p *Project) serveVideo(w http.ResponseWriter, r *http.Request) bool {
	v, ok := p.videos[r.URL.Path]
	if !ok {
		return false
	}
	info, e := v.file.Stat()
	if e != nil || info.Size() != v.info.Size() || !info.ModTime().Equal(v.info.ModTime()) {
		http.Error(w, "Video unavailable.", http.StatusServiceUnavailable)
		return true
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("ETag", `"`+v.name[:64]+`"`)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(3 * time.Minute))
	http.ServeContent(w, r, v.name, time.Time{}, io.NewSectionReader(v.file, 0, info.Size()))
	return true
}

// InstallVideos copies only catalogued immutable files. Existing matching bytes
// are retained; symlinks, corruption and absent source files stop the operation.
func InstallVideos(source, dest string, store *snips.Store) error {
	if len(store.VideoAssets()) == 0 {
		return nil
	}
	if e := os.Mkdir(dest, 0750); e != nil && !os.IsExist(e) {
		return e
	}
	r, e := os.OpenRoot(dest)
	if e != nil {
		return e
	}
	defer r.Close()
	info, e := os.Lstat(dest)
	if e != nil || !info.IsDir() {
		return fmt.Errorf("invalid media directory")
	}
	d, e := pinnedDir(r, ".")
	if e != nil {
		return e
	}
	defer d.Close()
	dir, e := d.Open(".")
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(-1)
	dir.Close()
	if e != nil {
		return e
	}
	var total int64
	for _, entry := range entries {
		if entry.Name()[0] == '.' {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("invalid video file")
		}
		if video.Filename.MatchString(entry.Name()) {
			total += info.Size()
		}
	}
	for _, v := range store.VideoAssets() {
		f, e := OpenVideo(dest, v.File)
		if e == nil {
			e = CheckVideo(f, v)
			f.Close()
			if e != nil {
				return e
			}
			continue
		}
		if !os.IsNotExist(e) {
			return e
		}
		f, e = OpenVideo(source, v.File)
		if e != nil {
			return e
		}
		if e = CheckVideo(f, v); e != nil {
			f.Close()
			return e
		}
		b, e := io.ReadAll(io.NewSectionReader(f, 0, v.Size))
		f.Close()
		if e != nil {
			return e
		}
		if total+v.Size > video.MaxTotal {
			return fmt.Errorf("video storage exceeds 64 MiB including retained files")
		}
		if int64(len(b)) != v.Size || fmt.Sprintf("%x.mp4", sha256.Sum256(b)) != v.File {
			return fmt.Errorf("video changed while copying")
		}
		if e = writeSiteFile(r, v.File, b); e != nil {
			return e
		}
		total += v.Size
	}
	return nil
}

// Seed files stay out of Git. Hashes and metadata in committed snips pin exactly
// which local immutable bytes may be added to a deployment snapshot.
func StageVideos(root, dest string) error {
	r, e := os.OpenRoot(dest)
	if e != nil {
		return e
	}
	defer r.Close()
	s, e := readSnipStore(r)
	if e != nil {
		return e
	}
	return InstallVideos(sitefiles.Path(root, filepath.Join("itsy", "media")), sitefiles.Path(dest, filepath.Join("itsy", "media")), s)
}
