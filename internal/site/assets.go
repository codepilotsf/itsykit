package site

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"syscall"
	"time"
)

// Small assets retain byte snapshots; large assets must remain immutable on disk.
// Keeping small files in memory also reduces the number of retained file handles.
const memoryAssetLimit = 64 << 10

type fileAsset struct {
	file *os.File
	info os.FileInfo
}

func (a fileAsset) unchanged() bool {
	info, err := a.file.Stat()
	return err == nil && info.Size() == a.info.Size() && info.ModTime().Equal(a.info.ModTime())
}

// Check around every read as well as before headers. If a developer edits an
// immutable file during a response, stop the transfer rather than send new bytes.
// As with videos, this relies on filesystem metadata and is not an OS sandbox
// against a privileged writer deliberately restoring metadata.
func (a fileAsset) ReadAt(b []byte, offset int64) (int, error) {
	if !a.unchanged() {
		return 0, fmt.Errorf("asset changed while serving")
	}
	n, err := a.file.ReadAt(b, offset)
	if !a.unchanged() {
		clear(b[:n])
		return 0, fmt.Errorf("asset changed while serving")
	}
	return n, err
}

func (p *Project) loadAsset(root *os.Root, route, name string) error {
	f, err := openSiteFile(root, name, 100<<20)
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		if !keep {
			f.Close()
		}
	}()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	a := fileAsset{f, info}
	// Theme editing/versioning consumes the CSS snapshot, including larger themes.
	if info.Size() <= memoryAssetLimit || route == "/css/theme.css" {
		data, err := io.ReadAll(io.LimitReader(f, (100<<20)+1))
		if err != nil {
			return err
		}
		if int64(len(data)) != info.Size() || !a.unchanged() {
			return fmt.Errorf("asset changed while loading: %s", name)
		}
		p.AssetData[route] = data
		return nil
	}
	if err := a.validateLinks(); err != nil {
		return err
	}
	p.assetFiles[route] = a
	keep = true
	return nil
}

func (a fileAsset) validateLinks() error {
	if stat, ok := a.info.Sys().(*syscall.Stat_t); ok && stat.Nlink != 1 {
		return fmt.Errorf("hard-linked streamed assets are unsupported")
	}
	return nil
}

func (p *Project) serveAssetFile(w http.ResponseWriter, r *http.Request) bool {
	a, ok := p.assetFiles[r.URL.Path]
	if !ok {
		return false
	}
	if !a.unchanged() {
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusServiceUnavailable)
			return true
		}
		http.Error(w, "Asset unavailable.", http.StatusServiceUnavailable)
		return true
	}
	// SectionReaders have independent offsets. Requests never reopen a pathname.
	// os.File cleanup closes handles once snapshots and active readers are collected.
	http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, io.NewSectionReader(a, 0, a.info.Size()))
	return true
}

// Only unpublished/failed loads may be closed explicitly. Published readers own
// their handles through reachability, like the existing video snapshots.
func (p *Project) closeAssetFiles() {
	for _, a := range p.assetFiles {
		a.file.Close()
	}
	for _, v := range p.videos {
		v.file.Close()
	}
}
