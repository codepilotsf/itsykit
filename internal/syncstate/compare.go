package syncstate

import (
	"encoding/json"
	"flea/internal/snips"
	"fmt"
)

// Comparison counts differing values, classifying direction only with a baseline.
// Local and Remote mean one side changed; Conflicts means both changed differently.
type Comparison struct {
	Different, Local, Remote, Conflicts int
	Known                               bool
}

func (c Comparison) String() string {
	if c.Different == 0 {
		return "Matches local"
	}
	if !c.Known {
		return fmt.Sprintf("%d differences; direction unknown (no pull baseline)", c.Different)
	}
	return fmt.Sprintf("%d local, %d remote, %d conflicting changes", c.Local, c.Remote, c.Conflicts)
}
func compare(known bool, base, local, remote map[string]string) Comparison {
	c := Comparison{Known: known}
	keys := map[string]bool{}
	for k := range local {
		keys[k] = true
	}
	for k := range remote {
		keys[k] = true
	}
	for k := range keys {
		if local[k] == remote[k] {
			continue
		}
		c.Different++
		if !known {
			continue
		}
		switch {
		case remote[k] == base[k]:
			c.Local++
		case local[k] == base[k]:
			c.Remote++
		default:
			c.Conflicts++
		}
	}
	return c
}

// CompareContent compares active editable values, collections, metadata and media
// libraries. Developer labels, page ownership and retained archives are excluded.
func CompareContent(base, local, remote *snips.Store) Comparison {
	return compare(base != nil, contentValues(base), contentValues(local), contentValues(remote))
}
func contentValues(s *snips.Store) map[string]string {
	values := map[string]string{}
	if s == nil {
		return values
	}
	put := func(key string, value any) { data, _ := json.Marshal(value); values[key] = string(data) }
	for _, p := range s.Blog {
		put("blog:"+p.ID, p)
	}
	entries, _ := s.Index()
	for id, e := range entries {
		e.Label = ""
		put("snip:"+id, e)
	}
	for id, v := range s.Media {
		put("media:"+id, v)
	}
	for id, v := range s.Videos {
		put("video:"+id, v)
	}
	page := func(path string, p *snips.Page) {
		if p == nil {
			return
		}
		if p.Title != "" {
			put("title:"+path, p.Title)
		}
		if p.Description != "" {
			put("description:"+path, p.Description)
		}
		for name, items := range p.Multisnips {
			if len(items) > 0 {
				put("collection:"+path+":"+name, items)
			}
		}
	}
	page("shared", &s.Shared)
	for path, p := range s.Pages {
		page(path, p)
	}
	return values
}
