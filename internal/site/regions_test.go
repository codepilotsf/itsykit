package site

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/snips"
)

func TestEditableRegionsExtractRenderAndReload(t *testing.T) {
	root := fixture(t)
	source := `<main class='keep'><h1 snip="title">Hello <em>world</em></h1><section snip="body"><p>Stay <strong>awhile</strong>.</p></section></main>`
	put(t, root, "index.html", source)
	if _, err := Load(root); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if string(before) != source {
		t.Fatal("status rewrote template")
	}
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if !strings.HasPrefix(string(data), `<main class='keep'>`) {
		t.Fatal(string(data))
	}
	ids := richMatches(string(data))
	if len(ids) != 2 || !snips.ValidID(ids[0][1]) {
		t.Fatal(string(data))
	}
	content := p.Content
	entry := content.Pages["/"].Snips[ids[0][1]]
	if entry.Label != "title" || entry.Content != "Hello <em>world</em>" {
		t.Fatal(entry)
	}
	rendered, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(rendered), "Hello <em>world</em>") {
		t.Fatal(string(rendered), err)
	}
	if _, err = LoadEditing(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if string(data) != string(after) {
		t.Fatal("extraction was not idempotent")
	}
	entry.Content = `Owner <strong>edit</strong><script>alert(1)</script><a href="javascript:alert(1)" onclick="bad()">link</a>{app.site_name}[[NW bad NW]]`
	content.Pages["/"].Snips[ids[0][1]] = entry
	content.Pages["/"].Title = "Custom title"
	content.Pages["/"].Description = `<script>metadata</script>`
	b, _ := content.Bytes()
	put(t, root, "itsy/snips.json", string(b))
	put(t, root, "main.layout.html", `<title>{page.title}</title><meta name="description" content="{page.description}"><body>{slot}</body>`)
	p = mustLoad(t, root)
	rendered, err = p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(rendered), "<strong>edit</strong>") || strings.Contains(string(rendered), "onclick") || strings.Contains(string(rendered), "javascript:") || !strings.Contains(string(rendered), "{app.site_name}[[NW bad NW]]") || !strings.Contains(string(rendered), "Custom title") {
		t.Fatal(string(rendered), err)
	}
	if err = p.CheckExtractedSnips(); err != nil {
		t.Fatal(err)
	}
}

func TestRegionValidationAndNoWritesOnFailure(t *testing.T) {
	for _, source := range []string{
		`<div snip><p snip>nested</p></div>`, `<h1 snip>{app.site_name}</h1>`, `<img snip>`, `<script snip>alert(1)</script>`, `<div snip><p>x</div>`,
	} {
		root := fixture(t)
		put(t, root, "index.html", source)
		if err := ExtractSnips(root); err == nil {
			t.Fatal(source)
		}
		after, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
		if string(after) != source {
			t.Fatal("rewrote invalid source")
		}
	}
	for _, source := range []string{`<div title='{snip a7Km2pQ9xB}'></div>`, `<textarea>{snip a7Km2pQ9xB}</textarea>`, `{snip a7Km2pQ9xB}`} {
		if _, _, err := scanRegions("index.html", source, nil); err == nil {
			t.Fatal(source)
		}
	}
}

func TestLegacyTenCharacterNameStillEscapes(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy/snips/abcdefghij.json", `{"text":"<b>legacy</b>"}`)
	put(t, root, "index.html", `<p>{snip abcdefghij}</p><p title='{snip abcdefghij}'>Test</p>`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(got), "&lt;b&gt;legacy&lt;/b&gt;") {
		t.Fatal(string(got), err)
	}
}

func TestSnipExtractionRecovery(t *testing.T) {
	root := fixture(t)
	source := `<h1 snip="title">Original</h1>`
	after := `<h1>{snip a7Km2pQ9xB}</h1>`
	put(t, root, "index.html", source)
	store := snips.NewStore()
	store.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Label: "title", Content: "Original"}
	data, _ := store.Bytes()
	tx := snipTransaction{Content: data, Edits: map[string]snipEdit{"site/index.html": {Before: source, After: after}}}
	journal, _ := json.Marshal(tx)
	put(t, root, transactionFile, string(journal))
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if string(got) != after {
		t.Fatal(string(got))
	}
	if _, err := os.Stat(filepath.Join(root, transactionFile)); !os.IsNotExist(err) {
		t.Fatal("journal not cleared")
	}
}

func TestSnipReferencesIncludeUnrenderedBranches(t *testing.T) {
	p := &Project{files: map[string]string{"index.html": `<div>{#if snip.absent}{snip a7Km2pQ9xB}{/if}</div>`, "_card.html": `<p>{snip b8Ln3rR0yC}</p>`, "about.html": `<p>{snip a7Km2pQ9xB}</p>`}}
	refs, err := p.SnipReferences()
	if err != nil || len(refs) != 2 || refs["a7Km2pQ9xB"] != "" || refs["b8Ln3rR0yC"] != "" {
		t.Fatal(refs, err)
	}
}

func TestRecoveryWithExistingFormattedContent(t *testing.T) {
	for _, saved := range []bool{false, true} {
		t.Run(fmt.Sprint(saved), func(t *testing.T) {
			root := fixture(t)
			source := `<h1 snip="title">Original</h1>`
			after := `<h1>{snip a7Km2pQ9xB}</h1>`
			put(t, root, "index.html", source)
			before, _ := snips.NewStore().Bytes()
			store := snips.NewStore()
			store.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Label: "title", Content: "Original"}
			content, _ := store.Bytes()
			tx := snipTransaction{BeforeContent: before, Content: content, Edits: map[string]snipEdit{"site/index.html": {Before: source, After: after}}}
			journal, _ := json.Marshal(tx)
			put(t, root, transactionFile, string(journal))
			if saved {
				put(t, root, "itsy/snips.json", string(content))
			} else {
				put(t, root, "itsy/snips.json", string(before))
			}
			if err := ExtractSnips(root); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
			if string(got) != after {
				t.Fatal(string(got))
			}
		})
	}
}

func TestDevSynchronizesSnipAdditionsRemovalsAndRestore(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<h1 snip="title">Original</h1>`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil || !p.SnipsUpdated {
		t.Fatal(p, err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	refs := richMatches(string(data))
	if len(refs) != 1 || strings.Contains(string(data), `{snip "`) {
		t.Fatal(string(data))
	}
	id := refs[0][1]
	p.Content.Page("/").Snips[id] = snips.Entry{Label: "title", Content: "Owner content"}
	p.Content.Page("/").Title = "Keep metadata"
	if err := SaveContent(root, p.Content); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", `<h1 snip="renamed">{snip `+id+`}</h1><p snip="new">Added</p>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	entry := p.Content.Page("/").Snips[id]
	if entry.Content != "Owner content" || entry.Label != "renamed" || len(p.Content.Page("/").Snips) != 2 {
		t.Fatal(p.Content)
	}
	// A partial's reference keeps this ID active even after removal from the page.
	put(t, root, "_keep.html", `<p>{snip `+id+`}</p>`)
	put(t, root, "index.html", `<h1>No regions</h1>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Content.Page("/").Snips) != 0 || p.Content.Shared.Snips[id].Content != "Owner content" || len(p.Content.Archive) != 1 {
		t.Fatal(p.Content)
	}
	if err := os.Remove(filepath.Join(root, "site", "_keep.html")); err != nil {
		t.Fatal(err)
	}
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Content.Shared.Snips) != 0 || p.Content.Archive[id].Content != "Owner content" {
		t.Fatal(p.Content)
	}
	put(t, root, "index.html", `<h1 snip="restored">{snip `+id+`}</h1>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Content.Page("/").Snips[id].Content != "Owner content" || p.Content.Page("/").Title != "Keep metadata" {
		t.Fatal(p.Content)
	}
	if _, ok := p.Content.Archive[id]; ok {
		t.Fatal("restored ID is still archived")
	}
	p, err = LoadEditing(context.Background(), root)
	if err != nil || p.SnipsUpdated {
		t.Fatal("unchanged reload wrote content", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "itsy", ".snips.lock")); !os.IsNotExist(err) {
		t.Fatal("local lock file exists", err)
	}
}

func TestUnquotedSnipContextsAndInvalidIDs(t *testing.T) {
	for _, source := range []string{
		`<p title="{snip a7Km2pQ9xB}">x</p>`, `<textarea>{snip a7Km2pQ9xB}</textarea>`,
		`<title>{snip a7Km2pQ9xB}</title>`, `{snip a7Km2pQ9xB}`,
		`<p>{snip a7Km2pQ9xB extra}</p>`, `<p>{snip a7Km{pQ9xB}</p>`, `<p>{snip}</p>`,
	} {
		root := fixture(t)
		store := snips.NewStore()
		store.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "<em>Safe</em>"}
		if err := SaveContent(root, store); err != nil {
			t.Fatal(err)
		}
		put(t, root, "index.html", source)
		if _, err := Load(root); err == nil {
			t.Fatal("accepted", source)
		}
	}
	root := fixture(t)
	put(t, root, "index.html", `<p>{snip team}</p>`)
	p := mustLoad(t, root)
	body, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(body), "&lt;b&gt;") {
		t.Fatal(string(body), err)
	}
}
