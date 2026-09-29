package snips

import (
	"bytes"
	"encoding/json"
	"flea/internal/media"
	"flea/internal/video"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Entry struct {
	Label   string       `json:"label,omitempty"`
	Content string       `json:"content,omitempty"`
	Type    string       `json:"type,omitempty"`
	Image   *media.Photo `json:"image,omitempty"`
	Video   *video.Video `json:"video,omitempty"`
	Poster  *media.Photo `json:"poster,omitempty"`
}
type Page struct {
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Snips       map[string]Entry       `json:"snips"`
	Multisnips  map[string][]MultiItem `json:"multisnips,omitempty"`
}
type MultiItem struct {
	ID     string            `json:"id"`
	Fields map[string]string `json:"fields"`
}
type Store struct {
	Videos  map[string]video.Video `json:"videos,omitempty"`
	Media   map[string]media.Photo `json:"media,omitempty"`
	Pages   map[string]*Page       `json:"pages"`
	Shared  Page                   `json:"shared"`
	Archive map[string]Entry       `json:"archive,omitempty"`
}

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
func NewStore() *Store {
	return &Store{Pages: map[string]*Page{}, Shared: Page{Snips: map[string]Entry{}}, Archive: map[string]Entry{}}
}
func Decode(data []byte) (*Store, error) {
	if len(data) > 5<<20 {
		return nil, fmt.Errorf("snips.json exceeds 5 MiB")
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		return nil, fmt.Errorf("snips.json must be an object")
	}
	if err := uniqueKeys(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return nil, err
	}
	s := NewStore()
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(s); err != nil {
		return nil, fmt.Errorf("snips.json: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("snips.json must contain one JSON object")
	}
	if s.Pages == nil {
		s.Pages = map[string]*Page{}
	}
	if s.Shared.Snips == nil {
		s.Shared.Snips = map[string]Entry{}
	}
	if s.Archive == nil {
		s.Archive = map[string]Entry{}
	}
	_, err := s.Index()
	return s, err
}

func uniqueKeys(d *json.Decoder) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("invalid JSON key")
			}
			if seen[name] {
				return fmt.Errorf("duplicate JSON key %q in snips.json", name)
			}
			seen[name] = true
			if err := uniqueKeys(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueKeys(d); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
	_, err = d.Token()
	return err
}
func (s *Store) Index() (map[string]Entry, error) {
	if len(s.ImageAssets()) > media.MaxPhotos || len(s.VideoAssets()) > video.MaxVideos {
		return nil, fmt.Errorf("media snip limits exceeded")
	}
	if err := video.Validate(s.Videos); err != nil {
		return nil, err
	}
	for _, v := range s.Videos {
		if v.Poster != "" {
			if _, ok := s.Media[v.Poster]; !ok {
				return nil, fmt.Errorf("unknown video poster photo")
			}
		}
	}
	if err := media.Validate(s.Media); err != nil {
		return nil, err
	}
	result := map[string]Entry{}
	add := func(entries map[string]Entry) error {
		for id, e := range entries {
			if err := validateEntry(e); err != nil {
				return fmt.Errorf("snip %s: %w", id, err)
			}
			if !ValidID(id) {
				return fmt.Errorf("invalid snip ID %q", id)
			}
			if _, ok := result[id]; ok {
				return fmt.Errorf("duplicate snip ID %s", id)
			}
			result[id] = e
		}
		return nil
	}
	if err := add(s.Shared.Snips); err != nil {
		return nil, err
	}
	for route, page := range s.Pages {
		if page == nil || len(route) == 0 || route[0] != '/' {
			return nil, fmt.Errorf("invalid snip page %q", route)
		}
		if err := add(page.Snips); err != nil {
			return nil, err
		}
		if err := ValidateMultisnips(page.Multisnips); err != nil {
			return nil, fmt.Errorf("%s: %w", route, err)
		}
	}
	if len(s.Shared.Multisnips) > 0 {
		return nil, fmt.Errorf("multisnips belong to a page, not shared content")
	}
	for id, entry := range s.Archive {
		if err := validateEntry(entry); err != nil {
			return nil, err
		}
		if !ValidID(id) {
			return nil, fmt.Errorf("invalid archived snip ID %q", id)
		}
	}
	return result, nil
}
func (s *Store) Page(route string) *Page {
	if route == "" {
		return &s.Shared
	}
	p := s.Pages[route]
	if p == nil {
		p = &Page{Snips: map[string]Entry{}}
		s.Pages[route] = p
	}
	if p.Snips == nil {
		p.Snips = map[string]Entry{}
	}
	return p
}
func (s *Store) Bytes() ([]byte, error) {
	b, e := json.MarshalIndent(s, "", "  ")
	return append(b, '\n'), e
}

// Merge preserves owner content and metadata, rehomes referenced IDs, and keeps
// removed entries in an archive. refs are lexical template references, not a render.
func Merge(seed, remote *Store, refs map[string]string) (*Store, error) {
	local, err := seed.Index()
	if err != nil {
		return nil, err
	}
	old, err := remote.Index()
	if err != nil {
		return nil, err
	}
	out := NewStore()
	out.Videos = map[string]video.Video{}
	for id, v := range seed.Videos {
		out.Videos[id] = v
	}
	for id, v := range remote.Videos {
		out.Videos[id] = v
	}
	out.Media = map[string]media.Photo{}
	for id, p := range seed.Media {
		out.Media[id] = p
	}
	for id, p := range remote.Media {
		out.Media[id] = p
	}
	for id, e := range seed.Archive {
		out.Archive[id] = e
	}
	for id, e := range local {
		out.Archive[id] = e
	}
	for id, e := range remote.Archive {
		out.Archive[id] = e
	}
	for id, e := range old {
		out.Archive[id] = e
	}
	for route, p := range seed.Pages {
		q := out.Page(route)
		q.Title = p.Title
		q.Description = p.Description
		q.Multisnips = CloneMultisnips(p.Multisnips)
	}
	for route, p := range remote.Pages {
		q := out.Page(route)
		q.Title = p.Title
		q.Description = p.Description
		if q.Multisnips == nil {
			q.Multisnips = map[string][]MultiItem{}
		}
		for name, items := range CloneMultisnips(p.Multisnips) {
			q.Multisnips[name] = items
		}
	}
	for id, route := range refs {
		e, ok := old[id]
		if !ok {
			e, ok = remote.Archive[id]
		}
		if !ok {
			e, ok = local[id]
		}
		if !ok {
			return nil, fmt.Errorf("missing content for snip %s", id)
		}
		if l, ok := local[id]; ok {
			e.Label = l.Label
		}
		out.Page(route).Snips[id] = e
		delete(out.Archive, id)
	}
	return out, nil
}
