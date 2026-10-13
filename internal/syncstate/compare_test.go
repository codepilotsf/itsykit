package syncstate

import (
	"flea/internal/snips"
	"testing"
)

func TestCompareDirections(t *testing.T) {
	for _, tc := range []struct {
		local, remote string
		want          Comparison
	}{
		{"base", "base", Comparison{Known: true}},
		{"local", "base", Comparison{Different: 1, Local: 1, Known: true}},
		{"base", "remote", Comparison{Different: 1, Remote: 1, Known: true}},
		{"local", "remote", Comparison{Different: 1, Conflicts: 1, Known: true}},
		{"same", "same", Comparison{Known: true}},
	} {
		if got := CompareContent(store("base"), store(tc.local), store(tc.remote)); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	unknown := CompareContent(nil, store("local"), store("remote"))
	if unknown.Different != 1 || unknown.Known || unknown.Remote != 0 || unknown.Local != 0 {
		t.Fatal(unknown)
	}
}
func TestCompareActiveValuesAndMetadata(t *testing.T) {
	b, l, r := store("base"), store("base"), store("base")
	l.Page("/").Title = "New title"
	r.Page("/").Description = "New description"
	r.Page("/").Multisnips = map[string][]snips.MultiItem{"people": {{ID: "a7Km2pQ9xB", Fields: map[string]string{"name": "Sean"}}}}
	r.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Label: "developer label", Content: "base"}
	r.Archive["b8Ln3rR0yC"] = snips.Entry{Content: "retained"}
	if got := CompareContent(b, l, r); got.Different != 3 || got.Local != 1 || got.Remote != 2 || got.Conflicts != 0 {
		t.Fatal(got)
	}
	if l.Page("/").Snips["a7Km2pQ9xB"].Content != "base" {
		t.Fatal("mutated content")
	}
	r.Page("/").Title = ""
	r.Page("/").Description = ""
	r.Page("/").Multisnips = nil
	if got := CompareContent(b, b, r); got.Different != 0 {
		t.Fatal("labels or archives counted", got)
	}
}
