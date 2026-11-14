// Package syncstate holds the explicit three-way synchronization contract.
// It knows nothing about SSH, files, or authentication.
package syncstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"itsy/internal/media"
	"itsy/internal/snips"
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

// MergeContent compares Blog posts and cover photos as whole records.
// Missing local IDs are not deletion requests.
func MergeContent(base, local, remote *snips.Store, resolve Resolve) (*snips.Store, error) {
	if base == nil || local == nil || remote == nil {
		return nil, fmt.Errorf("run itsy pull before publishing existing Blog content")
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
	// Posts conflict as whole records; omitted IDs never delete live posts.
	postMap := func(s *snips.Store) map[string]snips.BlogPost {
		m := map[string]snips.BlogPost{}
		for _, p := range s.Blog {
			m[p.ID] = p
		}
		return m
	}
	bp, lp, rp := postMap(base), postMap(local), postMap(remote)
	ids := make([]string, 0, len(lp))
	for id := range lp {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		encode := func(m map[string]snips.BlogPost) string {
			p, ok := m[id]
			if !ok {
				return "null"
			}
			b, _ := json.Marshal(p)
			return string(b)
		}
		v, e := choose("blog post "+id, encode(bp), encode(lp), encode(rp), resolve)
		if e != nil {
			return nil, e
		}
		if v != "null" {
			var p snips.BlogPost
			if e = json.Unmarshal([]byte(v), &p); e != nil {
				return nil, e
			}
			rp[id] = p
		}
	}
	out.Blog = nil
	for _, p := range remote.Blog {
		out.Blog = append(out.Blog, rp[p.ID])
		delete(rp, p.ID)
	}
	for _, id := range ids {
		if p, ok := rp[id]; ok {
			out.Blog = append(out.Blog, p)
			delete(rp, id)
		}
	}
	if err := out.ValidateBlog(); err != nil {
		return nil, err
	}
	return out, nil
}
