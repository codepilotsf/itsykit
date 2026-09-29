package site

import (
	"flea/internal/snips"
	"strings"
	"testing"
)

func TestRichContentCannotMoveIntoHeading(t *testing.T) {
	root := fixture(t)
	content := snips.NewStore()
	content.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "<p>Owner paragraph</p>"}
	put(t, root, "index.html", `<article>{snip a7Km2pQ9xB}</article>`)
	if _, e := LoadWithSnips(root, content); e != nil {
		t.Fatal(e)
	}
	put(t, root, "index.html", `<h1>{snip a7Km2pQ9xB}</h1>`)
	if _, e := LoadWithSnips(root, content); e == nil || !strings.Contains(e.Error(), "cannot fit") {
		t.Fatal(e)
	}
	if content.Page("/").Snips["a7Km2pQ9xB"].Content != "<p>Owner paragraph</p>" {
		t.Fatal("content changed")
	}
}
