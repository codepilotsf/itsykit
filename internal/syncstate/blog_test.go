package syncstate

import (
	"itsy/internal/snips"
	"testing"
)

func TestBlogThreeWaySync(t *testing.T) {
	store := func(title string) *snips.Store {
		s := snips.NewStore()
		s.Blog = []snips.BlogPost{{ID: "Abcdef1234", Title: title, Slug: "post", Status: "draft"}}
		return s
	}
	base, local, remote := store("Base"), store("Local"), store("Remote")
	if _, err := MergeSnips(base, local, remote, nil); err == nil {
		t.Fatal("conflict not detected")
	}
	merged, err := MergeSnips(base, local, base, nil)
	if err != nil || merged.Blog[0].Title != "Local" {
		t.Fatal(merged, err)
	}
	merged, err = MergeSnips(base, base, remote, nil)
	if err != nil || merged.Blog[0].Title != "Remote" {
		t.Fatal(merged, err)
	}
	merged, err = MergeSnips(base, snips.NewStore(), remote, nil)
	if err != nil || len(merged.Blog) != 1 {
		t.Fatal("omission deleted a post", err)
	}
	if got := CompareContent(base, local, remote); got.Conflicts != 1 {
		t.Fatal(got)
	}
}
