package editstore

import (
	"bytes"
	"fmt"
	"itsy/internal/editor"
	"itsy/internal/media"
	"itsy/internal/site"
	"itsy/internal/sitefiles"
	"os"
	"path/filepath"
	"syscall"
)

func (s *Store) mediaRoot() string {
	if s.ConfigPath != "" {
		return site.ManagedMediaRoot(s.ContentPath)
	}
	return sitefiles.Path(s.Root, filepath.Join("itsy", "media"))
}

// runtimeRoot keeps custom-app storage stable and private beside admin sessions
// in production, and inside the dev site otherwise.
func (s *Store) runtimeRoot() string {
	if s.ConfigPath != "" {
		return site.ManagedRuntimeRoot(s.ContentPath)
	}
	return sitefiles.Path(s.Root, filepath.Join("itsy", "runtime"))
}

func (s *Store) saveCoverPhoto(input []byte, name, alt string) (media.Photo, error) {
	photo, data, err := media.Process(input, name, alt)
	if err != nil {
		return photo, &editor.MediaValidationError{Message: err.Error()}
	}
	return photo, s.savePhotoBytes(photo.File, data)
}

func (s *Store) savePhotoBytes(name string, data []byte) error {
	root := s.mediaRoot()
	if s.ConfigPath == "" {
		if err := site.IgnoreSessionFiles(s.Root); err != nil {
			return err
		}
	}
	if err := os.Mkdir(root, 0750); err != nil && !os.IsExist(err) {
		return err
	}
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid media directory")
	}
	// Reject nonregular entries, and count each media class independently.
	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	var total int64
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("invalid media file")
		}
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink != 1 {
			return fmt.Errorf("hard-linked media files are unsupported")
		}
		if media.Filename.MatchString(file.Name()) {
			total += info.Size()
		}
	}
	existing, err := site.ReadPhoto(root, name)
	if err == nil {
		if !bytes.Equal(existing, data) {
			return fmt.Errorf("immutable media file differs")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	limit := int64(media.MaxTotal)
	if total+int64(len(data)) > limit {
		return &editor.MediaValidationError{Message: "Media storage is full (64 MiB). Ask your developer to archive unused files."}
	}
	return site.WriteManagedFile(root, name, data)
}
