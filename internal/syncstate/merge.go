// Package syncstate holds the explicit three-way synchronization contract.
// It knows nothing about SSH, files, or authentication.
package syncstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"flea/internal/media"
	"flea/internal/snips"
	"flea/internal/video"
)

type Snapshot struct {
	Exists   bool         `json:"exists"`
	Snips    *snips.Store `json:"snips"`
	Revision string       `json:"revision"`
}

func (s Snapshot) Version() string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Baselines belong to one exact destination, not just a branch name.
type Baseline struct {
	Destination string       `json:"destination"`
	Snips       *snips.Store `json:"snips,omitempty"`
}

type Plan struct {
	Expected string       `json:"expected"`
	Snips    *snips.Store `json:"snips,omitempty"`
}

type Conflict struct{ Key, Local, Remote string }
type Resolve func(Conflict) (bool, error) // true chooses local; false keeps remote.

func choose(key, base, local, remote string, resolve Resolve) (string, error) {
	if local == remote || remote == base {
		return local, nil
	}
	if local == base {
		return remote, nil
	}
	if resolve == nil {
		return "", fmt.Errorf("conflicting changes to %s; review locally and retry", key)
	}
	useLocal, err := resolve(Conflict{key, local, remote})
	if err != nil {
		return "", err
	}
	if useLocal {
		return local, nil
	}
	return remote, nil
}

// MergeSnips compares individual content and metadata values. Labels and region
// ownership belong to templates. Missing local IDs are not deletion requests.
func MergeSnips(base, local, remote *snips.Store, resolve Resolve) (*snips.Store, error) {
	if base == nil || local == nil || remote == nil {
		return nil, fmt.Errorf("run flea pull before publishing existing snips")
	}
	bytes, err := remote.Bytes()
	if err != nil {
		return nil, err
	}
	out, err := snips.Decode(bytes)
	if err != nil {
		return nil, err
	}
	if out.Media == nil {
		out.Media = map[string]media.Photo{}
	}
	if out.Videos == nil {
		out.Videos = map[string]video.Video{}
	}
	videoIDs := make([]string, 0, len(local.Videos))
	for id := range local.Videos {
		videoIDs = append(videoIDs, id)
	}
	sort.Strings(videoIDs)
	for _, id := range videoIDs {
		encode := func(s *snips.Store) string {
			v, ok := s.Videos[id]
			if !ok {
				return "null"
			}
			b, _ := json.Marshal(v)
			return string(b)
		}
		value, e := choose("video "+id, encode(base), encode(local), encode(remote), resolve)
		if e != nil {
			return nil, e
		}
		if value != "null" {
			var v video.Video
			if e = json.Unmarshal([]byte(value), &v); e != nil {
				return nil, e
			}
			out.Videos[id] = v
		}
	}
	photoIDs := make([]string, 0, len(local.Media))
	for id := range local.Media {
		photoIDs = append(photoIDs, id)
	}
	sort.Strings(photoIDs)
	for _, id := range photoIDs {
		encode := func(s *snips.Store) string {
			p, ok := s.Media[id]
			if !ok {
				return "null"
			}
			b, _ := json.Marshal(p)
			return string(b)
		}
		value, e := choose("photo "+id, encode(base), encode(local), encode(remote), resolve)
		if e != nil {
			return nil, e
		}
		if value != "null" {
			var p media.Photo
			if e = json.Unmarshal([]byte(value), &p); e != nil {
				return nil, e
			}
			out.Media[id] = p
		}
	}
	index := func(s *snips.Store) (map[string]snips.Entry, error) {
		m, e := s.Index()
		if e != nil {
			return nil, e
		}
		for id, entry := range s.Archive {
			if _, ok := m[id]; !ok {
				m[id] = entry
			}
		}
		return m, nil
	}
	b, err := index(base)
	if err != nil {
		return nil, err
	}
	l, err := index(local)
	if err != nil {
		return nil, err
	}
	r, err := index(remote)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(l))
	for id := range l {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		entry := l[id]
		label := entry.Label
		encode := func(value snips.Entry) string {
			value.Label = ""
			data, _ := json.Marshal(value)
			return string(data)
		}
		re, exists := r[id]
		if exists {
			// Distinguish a new ID collision from an empty baseline value.
			be, known := b[id]
			localValue, remoteValue := encode(entry), encode(re)
			if !known && localValue != remoteValue {
				if resolve == nil {
					return nil, fmt.Errorf("new snip ID collision: %s", id)
				}
				useLocal, e := resolve(Conflict{"snip " + id, localValue, remoteValue})
				if e != nil {
					return nil, e
				}
				if !useLocal {
					entry = re
				}
			} else {
				var selected string
				selected, err = choose("snip "+id, encode(be), localValue, remoteValue, resolve)
				if err != nil {
					return nil, err
				}
				entry = snips.Entry{}
				if err = json.Unmarshal([]byte(selected), &entry); err != nil {
					return nil, err
				}
			}
		}
		entry.Label = label
		delete(out.Archive, id)
		delete(out.Shared.Snips, id)
		for _, page := range out.Pages {
			delete(page.Snips, id)
		}
		out.Archive[id] = entry
	}
	routes := make([]string, 0, len(local.Pages))
	for route := range local.Pages {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	for _, route := range routes {
		lp := local.Pages[route]
		bp, rp := base.Pages[route], remote.Pages[route]
		if bp == nil {
			bp = &snips.Page{}
		}
		if rp == nil {
			rp = &snips.Page{}
		}
		op := out.Page(route)
		if op.Multisnips == nil {
			op.Multisnips = map[string][]snips.MultiItem{}
		}
		groupNames := make([]string, 0, len(lp.Multisnips))
		for name := range lp.Multisnips {
			groupNames = append(groupNames, name)
		}
		sort.Strings(groupNames)
		for _, name := range groupNames {
			encode := func(items []snips.MultiItem) string { b, _ := json.Marshal(items); return string(b) }
			value, e := choose(route+" multisnip "+name, encode(bp.Multisnips[name]), encode(lp.Multisnips[name]), encode(rp.Multisnips[name]), resolve)
			if e != nil {
				return nil, e
			}
			var items []snips.MultiItem
			if e = json.Unmarshal([]byte(value), &items); e != nil {
				return nil, e
			}
			if items == nil {
				delete(op.Multisnips, name)
			} else {
				op.Multisnips[name] = items
			}
		}
		op.Title, err = choose(route+" title", bp.Title, lp.Title, rp.Title, resolve)
		if err != nil {
			return nil, err
		}
		op.Description, err = choose(route+" description", bp.Description, lp.Description, rp.Description, resolve)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
