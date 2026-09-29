package site

import (
	"flea/internal/media"
	"fmt"
	"os"
	"path/filepath"
)

// Production keeps uploads beside content, never in administrator-owned code.
func ManagedMediaRoot(contentPath string) string {
	return filepath.Join(filepath.Dir(contentPath), "media")
}
func (p *Project) loadMedia() error {
	total := 0
	seen := map[string][]byte{}
	for _, photo := range p.Content.ImageAssets() {
		if data, ok := seen[photo.File]; ok {
			if err := media.CheckBytes(photo, data); err != nil {
				return err
			}
			continue
		}
		data, err := ReadPhoto(p.MediaRoot, photo.File)
		if err != nil {
			return fmt.Errorf("photo library: %w", err)
		}
		if err = media.CheckBytes(photo, data); err != nil {
			return err
		}
		seen[photo.File] = data
		total += len(data)
		if total > media.MaxTotal {
			return fmt.Errorf("photo library exceeds 64 MiB")
		}
		route := "/media/" + photo.File
		p.Assets[route] = "flea/media/" + photo.File
		p.AssetData[route] = data
	}
	return nil
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
