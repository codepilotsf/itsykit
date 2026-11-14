package snips

import (
	"fmt"
	"itsy/internal/media"
	"itsy/internal/video"
)

func validateEntry(e Entry) error {
	switch e.Type {
	case "":
		if e.Image != nil || e.Video != nil || e.Poster != nil {
			return fmt.Errorf("media requires an explicit snip type")
		}
	case "image":
		if e.Image == nil || e.Video != nil || e.Poster != nil || e.Content != "" {
			return fmt.Errorf("invalid image snip")
		}
		return media.Validate(map[string]media.Photo{"Snip123456": *e.Image})
	case "video":
		if e.Video == nil || e.Image != nil || e.Content != "" || e.Video.Poster != "" {
			return fmt.Errorf("invalid video snip")
		}
		if err := video.Validate(map[string]video.Video{"Snip123456": *e.Video}); err != nil {
			return err
		}
		if e.Poster != nil {
			return media.Validate(map[string]media.Photo{"Snip123456": *e.Poster})
		}
	default:
		return fmt.Errorf("unknown snip type")
	}
	return nil
}

func (s *Store) entries() []Entry {
	var out []Entry
	for _, e := range s.Shared.Snips {
		out = append(out, e)
	}
	for _, p := range s.Pages {
		if p != nil {
			for _, e := range p.Snips {
				out = append(out, e)
			}
		}
	}
	for _, e := range s.Archive {
		out = append(out, e)
	}
	return out
}

// ImageAssets includes retained archived media, so backup and sync remain recoverable.
func (s *Store) ImageAssets() []media.Photo {
	var out []media.Photo
	for _, p := range s.Media {
		out = append(out, p)
	}
	for _, e := range s.entries() {
		if e.Image != nil {
			out = append(out, *e.Image)
		}
		if e.Poster != nil {
			out = append(out, *e.Poster)
		}
	}
	return out
}
func (s *Store) VideoAssets() []video.Video {
	var out []video.Video
	for _, v := range s.Videos {
		out = append(out, v)
	}
	for _, e := range s.entries() {
		if e.Video != nil {
			out = append(out, *e.Video)
		}
	}
	return out
}
