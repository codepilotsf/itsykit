package syncstate

import (
	"encoding/json"
	"fmt"
	"itsy/internal/snips"
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

// CompareContent compares Blog posts and library photos as whole records.
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
	for id, v := range s.Media {
		put("media:"+id, v)
	}
	return values
}
