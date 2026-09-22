package site

import (
	"context"
	"strings"
	"testing"

	"flea/internal/snips"
)

func TestSnipsRequireUnquotedSyntax(t *testing.T) {
	root := fixture(t)
	store := snips.NewStore()
	store.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "<em>Hello</em>"}
	if err := SaveContent(root, store); err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{`{snip "a7Km2pQ9xB"}`, `{snip 'a7Km2pQ9xB'}`} {
		put(t, root, "index.html", "<p>"+marker+"</p>")
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "unquoted") {
			t.Fatal(marker, err)
		}
	}
	put(t, root, "index.html", `<p>{snip a7Km2pQ9xB}</p><!-- snip:missing123 --><!-- {snip ignored123} -->`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := p.SnipReferences()
	if err != nil || len(refs) != 1 {
		t.Fatal(refs, err)
	}
	body, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(body), "<em>Hello</em>") {
		t.Fatal(string(body), err)
	}
}
