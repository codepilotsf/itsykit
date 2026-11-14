package site

import (
	"fmt"
	"itsy/internal/media"
	"itsy/internal/sitefiles"
	"itsy/internal/snips"
	"os"
	"path/filepath"
)

// Production keeps uploads beside content, never in administrator-owned code.
func ManagedMediaRoot(contentPath string) string {
	return filepath.Join(filepath.Dir(contentPath), "media")
}

// ManagedRuntimeRoot is the stable per-site private runtime directory, a sibling
// of the content snapshot and its admin sessions. Custom-app storage lives under
// it so it survives immutable code releases; the release tree never holds it.
func ManagedRuntimeRoot(contentPath string) string {
	return sitefiles.Path(filepath.Dir(filepath.Dir(contentPath)), filepath.Join("itsy", "runtime"))
}
func (p *Project) loadMedia() error {
	var total int64
	seen := map[string]fileAsset{}
	for _, photo := range p.Content.ImageAssets() {
		route := "/media/" + photo.File
		if prior, ok := seen[photo.File]; ok {
			if err := media.CheckReader(photo, prior.file, prior.info.Size()); err != nil {
				return err
			}
			continue
		}
		f, err := openPhoto(p.MediaRoot, photo.File)
		if err != nil {
			return fmt.Errorf("photo library: %w", err)
		}
		info, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}
		a := fileAsset{f, info}
		if err := a.validateLinks(); err != nil {
			f.Close()
			return err
		}
		if err = media.CheckReader(photo, f, info.Size()); err != nil {
			f.Close()
			return err
		}
		if !a.unchanged() {
			f.Close()
			return fmt.Errorf("photo changed while loading")
		}
		total += info.Size()
		if total > media.MaxTotal {
			f.Close()
			return fmt.Errorf("photo library exceeds 64 MiB")
		}
		p.Assets[route] = "itsy/media/" + photo.File
		delete(p.AssetData, route)
		if prior, ok := p.assetFiles[route]; ok {
			prior.file.Close()
		}
		p.assetFiles[route] = a
		seen[photo.File] = a
	}
	return nil
}

func openPhoto(root, name string) (*os.File, error) {
	if !media.Filename.MatchString(name) {
		return nil, fmt.Errorf("invalid photo filename")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("photo directory is not a real directory")
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return openSiteFile(r, name, media.MaxFile)
}
func ReadPhoto(root, name string) ([]byte, error) {
	if !media.Filename.MatchString(name) {
		return nil, fmt.Errorf("invalid photo filename")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("photo directory is not a real directory")
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return readFileBounded(r, name, media.MaxFile)
}

// LoadWithMedia loads a code snapshot with managed content and media. runtimeRoot
// is the stable private runtime directory for custom-app storage; empty defaults
// to the site tree (development).
func LoadWithMedia(root string, content *snips.Store, mediaRoot, runtimeRoot string) (*Project, error) {
	return loadProject(root, content, mediaRoot, runtimeRoot, false)
}
