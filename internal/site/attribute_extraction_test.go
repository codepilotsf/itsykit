package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/snips"
)

func TestExtractionRemovesOnlyAuthoringAttribute(t *testing.T) {
	for _, tc := range []struct{ before, after string }{
		{`<p class="text-lg" snip>`, `<p class="text-lg">`},
		{`<p snip class='keep' title="a > flea b">`, `<p class='keep' title="a > flea b">`},
		{`<P SNIP = 'Label &amp; more' CLASS=keep>`, `<P CLASS=keep>`},
		{"<p\n  class=keep\n  snip\n  id='copy'>", "<p\n  class=keep\n  id='copy'>"},
		{`<p data-snip="keep" title='snip="not an attribute"' snip=label >`, `<p data-snip="keep" title='snip="not an attribute"' >`},
	} {
		t.Run(tc.before, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "index.html", tc.before+"Hello</p>")
			p, err := LoadEditing(context.Background(), root)
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(root, "site", "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			ids := richMatches(string(got))
			if len(ids) != 1 {
				t.Fatal(string(got))
			}
			want := tc.after + "{snip " + ids[0][1] + "}</p>"
			if string(got) != want {
				t.Fatalf("got %s; want %s", got, want)
			}
			if err := p.CheckExtractedSnips(); err != nil {
				t.Fatal(err)
			}
			p, err = LoadEditing(context.Background(), root)
			if err != nil || p.SnipsUpdated {
				t.Fatal("not idempotent", err)
			}
		})
	}
}

func TestRedundantAttributeRemovalPreservesStoredContent(t *testing.T) {
	root := fixture(t)
	store := snips.NewStore()
	const id = "b8Ln3rR0yC"
	store.Page("/").Snips[id] = snips.Entry{Label: "Existing label", Content: "Owner content"}
	store.Page("/").Title = "Owner title"
	if err := SaveContent(root, store); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", `<p class="text-lg text-dark m-0" snip>{snip `+id+`}</p>`)
	p := mustLoad(t, root)
	if err := p.CheckExtractedSnips(); err == nil {
		t.Fatal("authoring marker accepted for deploy")
	}
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if string(got) != `<p class="text-lg text-dark m-0">{snip `+id+`}</p>` {
		t.Fatal(string(got))
	}
	if entry := p.Content.Page("/").Snips[id]; entry.Content != "Owner content" || entry.Label != "Existing label" || p.Content.Page("/").Title != "Owner title" {
		t.Fatal("stored content changed")
	}
	// Move the directive to a partial: identity and content no longer need an attribute.
	put(t, root, "index.html", `{render copy}`)
	put(t, root, "_copy.html", string(got))
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Content.Shared.Snips[id].Content != "Owner content" {
		t.Fatal("shared scope lost")
	}
	body, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(body), "Owner content") {
		t.Fatal("directive no longer renders", err)
	}
}
