package syncstate

import (
	"itsy/internal/snips"
	"testing"
)

func store(value string) *snips.Store {
	s := snips.NewStore()
	s.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: value}
	return s
}
func TestThreeWay(t *testing.T) {
	for _, c := range []struct {
		local, remote, want string
		conflict            bool
	}{
		{"local", "base", "local", false}, {"base", "owner", "owner", false}, {"same", "same", "same", false}, {"local", "owner", "", true},
	} {
		result, err := MergeSnips(store("base"), store(c.local), store(c.remote), nil)
		if (err != nil) != c.conflict {
			t.Fatal(c, err)
		}
		if err == nil && result.Archive["a7Km2pQ9xB"].Content != c.want {
			t.Fatal(result)
		}
	}
}
func TestResolutionAndMetadata(t *testing.T) {
	b, l, r := store("base"), store("local"), store("owner")
	l.Page("/").Title = "Local title"
	r.Page("/").Description = "Owner description"
	merged, err := MergeSnips(b, l, r, func(c Conflict) (bool, error) { return false, nil })
	if err != nil || merged.Archive["a7Km2pQ9xB"].Content != "owner" || merged.Page("/").Title != "Local title" || merged.Page("/").Description != "Owner description" {
		t.Fatal(merged, err)
	}
	if l.Page("/").Snips["a7Km2pQ9xB"].Content != "local" {
		t.Fatal("mutated input")
	}
}
