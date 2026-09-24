package fleas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Entry struct {
	Label   string `json:"label,omitempty"`
	Content string `json:"content"`
}
type Page struct {
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Fleas       map[string]Entry `json:"fleas"`
}
type Store struct {
	Pages   map[string]*Page `json:"pages"`
	Shared  Page             `json:"shared"`
	Archive map[string]Entry `json:"archive,omitempty"`
}

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
func NewStore() *Store {
	return &Store{Pages: map[string]*Page{}, Shared: Page{Fleas: map[string]Entry{}}, Archive: map[string]Entry{}}
}
func Decode(data []byte) (*Store, error) {
	if len(data) > 5<<20 {
		return nil, fmt.Errorf("fleas.json exceeds 5 MiB")
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "{") {
		return nil, fmt.Errorf("fleas.json must be an object")
	}
	if err := uniqueKeys(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return nil, err
	}
	s := NewStore()
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(s); err != nil {
		return nil, fmt.Errorf("fleas.json: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("fleas.json must contain one JSON object")
	}
	if s.Pages == nil {
		s.Pages = map[string]*Page{}
	}
	if s.Shared.Fleas == nil {
		s.Shared.Fleas = map[string]Entry{}
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
				return fmt.Errorf("duplicate JSON key %q in fleas.json", name)
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
	result := map[string]Entry{}
	add := func(entries map[string]Entry) error {
		for id, e := range entries {
			if !ValidID(id) {
				return fmt.Errorf("invalid flea ID %q", id)
			}
			if _, ok := result[id]; ok {
				return fmt.Errorf("duplicate flea ID %s", id)
			}
			result[id] = e
		}
		return nil
	}
	if err := add(s.Shared.Fleas); err != nil {
		return nil, err
	}
	for route, page := range s.Pages {
		if page == nil || len(route) == 0 || route[0] != '/' {
			return nil, fmt.Errorf("invalid flea page %q", route)
		}
		if err := add(page.Fleas); err != nil {
			return nil, err
		}
	}
	for id := range s.Archive {
		if !ValidID(id) {
			return nil, fmt.Errorf("invalid archived flea ID %q", id)
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
		p = &Page{Fleas: map[string]Entry{}}
		s.Pages[route] = p
	}
	if p.Fleas == nil {
		p.Fleas = map[string]Entry{}
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
	}
	for route, p := range remote.Pages {
		q := out.Page(route)
		q.Title = p.Title
		q.Description = p.Description
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
			return nil, fmt.Errorf("missing content for flea %s", id)
		}
		if l, ok := local[id]; ok {
			e.Label = l.Label
		}
		out.Page(route).Fleas[id] = e
		delete(out.Archive, id)
	}
	return out, nil
}
