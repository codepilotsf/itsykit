// Package snips retains the existing content-file location for Blog storage.
package snips

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"itsy/internal/media"
	"regexp"
	"strings"
)

type Store struct {
	Blog  []BlogPost             `json:"blog,omitempty"`
	Media map[string]media.Photo `json:"media,omitempty"`
}

var idPattern = regexp.MustCompile(`^[a-zA-Z0-9]{10}$`)

func ValidID(id string) bool { return idPattern.MatchString(id) }
func NewStore() *Store       { return &Store{} }
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
		return nil, fmt.Errorf("snips.json (Blog and media only; page CMS fields are unsupported): %w", err)
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("snips.json must contain one JSON object")
	}
	return s, s.Validate()
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

func (s *Store) Validate() error {
	if err := media.Validate(s.Media); err != nil {
		return err
	}
	return s.ValidateBlog()
}
func (s *Store) Bytes() ([]byte, error) {
	b, e := json.MarshalIndent(s, "", "  ")
	return append(b, '\n'), e
}

// Merge preserves every live post and photo; local-only IDs seed new records.
func Merge(seed, remote *Store) (*Store, error) {
	if err := seed.Validate(); err != nil {
		return nil, err
	}
	if err := remote.Validate(); err != nil {
		return nil, err
	}
	out := NewStore()
	out.Blog = PreserveBlog(seed.Blog, remote.Blog)
	out.Media = map[string]media.Photo{}
	for id, p := range seed.Media {
		out.Media[id] = p
	}
	for id, p := range remote.Media {
		out.Media[id] = p
	}
	return out, out.Validate()
}
