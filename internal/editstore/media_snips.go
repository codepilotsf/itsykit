package editstore

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"unicode/utf8"

	"flea/internal/editor"
	"flea/internal/media"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/video"
)

func (s *Store) mediaRoot() string {
	if s.ConfigPath != "" {
		return site.ManagedMediaRoot(s.ContentPath)
	}
	return filepath.Join(s.Root, "flea", "media")
}

// Files become durable before the content transaction. A rejected draft can
// leave an unreferenced immutable file, never a published missing file.
func (s *Store) applyMedia(p *site.Project, edit editor.PageEdit) error {
	if len(edit.Media) == 0 {
		return nil
	}
	if len(edit.Media) > 100 {
		return fmt.Errorf("too many media changes")
	}
	allowed, err := p.EditableMedia(edit.Route)
	if err != nil {
		return err
	}
	index, err := p.Content.Index()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, change := range edit.Media {
		if len(change.Name) > 200 || len(change.Alt) > 1000 || !utf8.ValidString(change.Name+change.Alt) {
			return &editor.MediaValidationError{Message: "Use a name up to 200 bytes and alternative text up to 1,000 bytes."}
		}
		if change.Library {
			_, existing := index[change.ID]
			_, archived := p.Content.Archive[change.ID]
			if !snips.ValidID(change.ID) || seen[change.ID] || p.Content.Media[change.ID].File != "" || existing || archived || len(change.Data) == 0 || change.ClearPoster || len(change.PosterData) != 0 {
				return fmt.Errorf("invalid new collection photo")
			}
			defs, err := p.VisibleMultisnips(edit.Route)
			if err != nil {
				return err
			}
			referenced := false
			for name, items := range edit.Multisnips {
				for _, item := range items {
					for field, value := range item.Fields {
						if defs[name].Photos[field] && value == change.ID {
							referenced = true
						}
					}
				}
			}
			if !referenced {
				return fmt.Errorf("collection photo must belong to a visible photo field")
			}
			photo, err := s.saveSnipPhoto(change.Data, change.Name, change.Alt)
			if err != nil {
				return err
			}
			if p.Content.Media == nil {
				p.Content.Media = map[string]media.Photo{}
			}
			p.Content.Media[change.ID] = photo
			seen[change.ID] = true
			continue
		}
		kind, ok := allowed[change.ID]
		entry, exists := index[change.ID]
		if !ok || !exists || entry.Type != kind || seen[change.ID] {
			return fmt.Errorf("unknown, hidden or duplicate editable media")
		}
		seen[change.ID] = true
		switch kind {
		case "image":
			if change.ClearPoster || len(change.PosterData) > 0 || entry.Image == nil {
				return fmt.Errorf("invalid image change")
			}
			photo := *entry.Image
			photo.Name, photo.Alt = change.Name, change.Alt
			if len(change.Data) > 0 {
				photo, err = s.saveSnipPhoto(change.Data, change.Name, change.Alt)
				if err != nil {
					return err
				}
			}
			entry.Image = &photo
		case "video":
			if entry.Video == nil || change.Alt != "" || change.ClearPoster && len(change.PosterData) > 0 {
				return fmt.Errorf("invalid video change")
			}
			v := *entry.Video
			v.Name = change.Name
			if len(change.Data) > 0 {
				v.Metadata, err = video.Inspect(bytes.NewReader(change.Data), int64(len(change.Data)))
				if err != nil {
					return &editor.MediaValidationError{Message: err.Error()}
				}
				if err = s.saveSnipBytes(v.File, change.Data, true); err != nil {
					return err
				}
			}
			entry.Video = &v
			if change.ClearPoster {
				entry.Poster = nil
			}
			if len(change.PosterData) > 0 {
				photo, err := s.saveSnipPhoto(change.PosterData, "", "")
				if err != nil {
					return err
				}
				entry.Poster = &photo
			}
		default:
			return fmt.Errorf("unsupported media type")
		}
		if _, ok := p.Content.Shared.Snips[change.ID]; ok {
			p.Content.Shared.Snips[change.ID] = entry
		} else {
			for _, page := range p.Content.Pages {
				if _, ok := page.Snips[change.ID]; ok {
					page.Snips[change.ID] = entry
				}
			}
		}
	}
	_, err = p.Content.Index()
	return err
}

func (s *Store) saveSnipPhoto(input []byte, name, alt string) (media.Photo, error) {
	photo, data, err := media.Process(input, name, alt)
	if err != nil {
		return photo, &editor.MediaValidationError{Message: err.Error()}
	}
	return photo, s.saveSnipBytes(photo.File, data, false)
}

func (s *Store) saveSnipBytes(name string, data []byte, isVideo bool) error {
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
		if (filepath.Ext(file.Name()) == ".mp4") == isVideo {
			total += info.Size()
		}
	}
	var existing []byte
	if isVideo {
		var f *os.File
		f, err = site.OpenVideo(root, name)
		if err == nil {
			existing, err = io.ReadAll(io.LimitReader(f, video.MaxFile+1))
			f.Close()
		}
	} else {
		existing, err = site.ReadPhoto(root, name)
	}
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
	if isVideo {
		limit = video.MaxTotal
	}
	if total+int64(len(data)) > limit {
		return &editor.MediaValidationError{Message: "Media storage is full (64 MiB). Ask your developer to archive unused files."}
	}
	return site.WriteManagedFile(root, name, data)
}
