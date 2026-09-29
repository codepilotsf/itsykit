package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/snips"
)

func TestBareSnipExtraction(t *testing.T) {
	for _, attribute := range []string{`snip`, `snip=""`} {
		root := fixture(t)
		put(t, root, "index.html", `<section class="keep" `+attribute+`><p>Hello <b>world</b>.</p></section>`)
		p, err := LoadEditing(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		entries := p.Content.Page("/").Snips
		if len(entries) != 1 {
			t.Fatal(entries)
		}
		for id, entry := range entries {
			if !snips.ValidID(id) || entry.Label != "" || entry.Content != `<p>Hello <b>world</b>.</p>` {
				t.Fatal(id, entry)
			}
			want := `<section class="keep">{snip ` + id + `}</section>`
			got, err := os.ReadFile(filepath.Join(root, "site", "index.html"))
			if err != nil || string(got) != want {
				t.Fatal(string(got), err)
			}
			if _, err := LoadEditing(context.Background(), root); err != nil {
				t.Fatal(err)
			}
			again, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
			if string(again) != want {
				t.Fatal("unstable ID", string(again))
			}
		}
		if _, err := os.Stat(filepath.Join(root, "flea/snips.json")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOldContentIsRejectedWithoutWrites(t *testing.T) {
	for _, name := range []string{"flea/fleas.json", "flea/fleas/old.json", "flea/.fleas-pending.json"} {
		root := fixture(t)
		put(t, root, name, `{}`)
		const source = `<section snip>Keep this</section>`
		put(t, root, "index.html", source)
		if _, err := LoadEditing(context.Background(), root); err == nil {
			t.Fatal("accepted", name)
		}
		got, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
		if string(got) != source {
			t.Fatal("modified source", name)
		}
		if _, err := os.Stat(filepath.Join(root, "flea/snips.json")); !os.IsNotExist(err) {
			t.Fatal("created content", err)
		}
	}
	for _, source := range []string{`<p flea>Hello</p>`, `<p>{flea a7Km2pQ9xB}</p>`} {
		root := fixture(t)
		put(t, root, "index.html", source)
		if _, err := LoadEditing(context.Background(), root); err == nil {
			t.Fatal("accepted old syntax", source)
		}
	}
}

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
