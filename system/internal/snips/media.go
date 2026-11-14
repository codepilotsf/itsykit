package snips

import "itsy/internal/media"

// ImageAssets includes every retained Blog library photo, including unused covers.
func (s *Store) ImageAssets() []media.Photo {
	out := make([]media.Photo, 0, len(s.Media))
	for _, p := range s.Media {
		out = append(out, p)
	}
	return out
}
