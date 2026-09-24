package site

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/fleas"
)

func TestEditableRegionsExtractRenderAndReload(t *testing.T) {
	root := fixture(t)
	source := `<main class='keep'><h1 flea="title">Hello <em>world</em></h1><section flea="body"><p>Stay <strong>awhile</strong>.</p></section></main>`
	put(t, root, "index.html", source)
	if _, err := Load(root); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(root, "index.html"))
	if string(before) != source {
		t.Fatal("status rewrote template")
	}
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "index.html"))
	if !strings.HasPrefix(string(data), `<main class='keep'>`) {
		t.Fatal(string(data))
	}
	ids := richMatches(string(data))
	if len(ids) != 2 || !fleas.ValidID(ids[0][1]) {
		t.Fatal(string(data))
	}
	content := p.Content
	entry := content.Pages["/"].Fleas[ids[0][1]]
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
	after, _ := os.ReadFile(filepath.Join(root, "index.html"))
	if string(data) != string(after) {
		t.Fatal("extraction was not idempotent")
	}
	entry.Content = `Owner <strong>edit</strong><script>alert(1)</script><a href="javascript:alert(1)" onclick="bad()">link</a>{site.name}[[NW bad NW]]`
	content.Pages["/"].Fleas[ids[0][1]] = entry
	content.Pages["/"].Title = "Custom title"
	content.Pages["/"].Description = `<script>metadata</script>`
	b, _ := content.Bytes()
	put(t, root, "_/fleas.json", string(b))
	put(t, root, "_layout.html", `<title>{page.title}</title><meta name="description" content="{page.description}"><body>{slot}</body>`)
	p = mustLoad(t, root)
	rendered, err = p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(rendered), "<strong>edit</strong>") || strings.Contains(string(rendered), "onclick") || strings.Contains(string(rendered), "javascript:") || !strings.Contains(string(rendered), "{site.name}[[NW bad NW]]") || !strings.Contains(string(rendered), "Custom title") {
		t.Fatal(string(rendered), err)
	}
	if err = p.CheckExtractedFleas(); err != nil {
		t.Fatal(err)
	}
}

func TestRegionValidationAndNoWritesOnFailure(t *testing.T) {
	for _, source := range []string{
		`<div flea><p flea>nested</p></div>`, `<h1 flea>{site.name}</h1>`, `<img flea>`, `<script flea>alert(1)</script>`, `<div flea><p>x</div>`,
	} {
		root := fixture(t)
		put(t, root, "index.html", source)
		if err := ExtractFleas(root); err == nil {
			t.Fatal(source)
		}
		after, _ := os.ReadFile(filepath.Join(root, "index.html"))
		if string(after) != source {
			t.Fatal("rewrote invalid source")
		}
	}
	for _, source := range []string{`<div title='{flea a7Km2pQ9xB}'></div>`, `<textarea>{flea a7Km2pQ9xB}</textarea>`, `{flea a7Km2pQ9xB}`} {
		if _, _, err := scanRegions(source, nil); err == nil {
			t.Fatal(source)
		}
	}
}

func TestLegacyTenCharacterNameStillEscapes(t *testing.T) {
	root := fixture(t)
	put(t, root, "_/fleas/abcdefghij.json", `{"text":"<b>legacy</b>"}`)
	put(t, root, "index.html", `<p>{flea abcdefghij}</p><p title='{flea abcdefghij}'>Test</p>`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(got), "&lt;b&gt;legacy&lt;/b&gt;") {
		t.Fatal(string(got), err)
	}
}

func TestFleaExtractionRecovery(t *testing.T) {
	root := fixture(t)
	source := `<h1 flea="title">Original</h1>`
	after := `<h1 flea="title">{flea a7Km2pQ9xB}</h1>`
	put(t, root, "index.html", source)
	store := fleas.NewStore()
	store.Page("/").Fleas["a7Km2pQ9xB"] = fleas.Entry{Label: "title", Content: "Original"}
	data, _ := store.Bytes()
	tx := fleaTransaction{Content: data, Edits: map[string]fleaEdit{"index.html": {Before: source, After: after}}}
	journal, _ := json.Marshal(tx)
	put(t, root, transactionFile, string(journal))
	if err := ExtractFleas(root); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "index.html"))
	if string(got) != after {
		t.Fatal(string(got))
	}
	if _, err := os.Stat(filepath.Join(root, transactionFile)); !os.IsNotExist(err) {
		t.Fatal("journal not cleared")
	}
}

func TestFleaReferencesIncludeUnrenderedBranches(t *testing.T) {
	p := &Project{files: map[string]string{"index.html": `<div>{#if flea.absent}{flea a7Km2pQ9xB}{/if}</div>`, "_card.html": `<p>{flea b8Ln3rR0yC}</p>`, "about.html": `<p>{flea a7Km2pQ9xB}</p>`}}
	refs, err := p.FleaReferences()
	if err != nil || len(refs) != 2 || refs["a7Km2pQ9xB"] != "" || refs["b8Ln3rR0yC"] != "" {
		t.Fatal(refs, err)
	}
}

func TestRecoveryWithExistingFormattedContent(t *testing.T) {
	for _, saved := range []bool{false, true} {
		t.Run(fmt.Sprint(saved), func(t *testing.T) {
			root := fixture(t)
			source := `<h1 flea="title">Original</h1>`
			after := `<h1 flea="title">{flea a7Km2pQ9xB}</h1>`
			put(t, root, "index.html", source)
			before, _ := fleas.NewStore().Bytes()
			store := fleas.NewStore()
			store.Page("/").Fleas["a7Km2pQ9xB"] = fleas.Entry{Label: "title", Content: "Original"}
			content, _ := store.Bytes()
			tx := fleaTransaction{BeforeContent: before, Content: content, Edits: map[string]fleaEdit{"index.html": {Before: source, After: after}}}
			journal, _ := json.Marshal(tx)
			put(t, root, transactionFile, string(journal))
			if saved {
				put(t, root, "_/fleas.json", string(content))
			} else {
				put(t, root, "_/fleas.json", string(before))
			}
			if err := ExtractFleas(root); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(filepath.Join(root, "index.html"))
			if string(got) != after {
				t.Fatal(string(got))
			}
		})
	}
}

func TestDevSynchronizesFleaAdditionsRemovalsAndRestore(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<h1 flea="title">Original</h1>`)
	p, err := LoadEditing(context.Background(), root)
	if err != nil || !p.FleasUpdated {
		t.Fatal(p, err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "index.html"))
	refs := richMatches(string(data))
	if len(refs) != 1 || strings.Contains(string(data), `{flea "`) {
		t.Fatal(string(data))
	}
	id := refs[0][1]
	p.Content.Page("/").Fleas[id] = fleas.Entry{Label: "title", Content: "Owner content"}
	p.Content.Page("/").Title = "Keep metadata"
	if err := SaveContent(root, p.Content); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", `<h1 flea="renamed">{flea `+id+`}</h1><p flea="new">Added</p>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	entry := p.Content.Page("/").Fleas[id]
	if entry.Content != "Owner content" || entry.Label != "renamed" || len(p.Content.Page("/").Fleas) != 2 {
		t.Fatal(p.Content)
	}
	// A partial's reference keeps this ID active even after removal from the page.
	put(t, root, "_keep.html", `<p>{flea `+id+`}</p>`)
	put(t, root, "index.html", `<h1>No regions</h1>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Content.Page("/").Fleas) != 0 || p.Content.Shared.Fleas[id].Content != "Owner content" || len(p.Content.Archive) != 1 {
		t.Fatal(p.Content)
	}
	if err := os.Remove(filepath.Join(root, "_keep.html")); err != nil {
		t.Fatal(err)
	}
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Content.Shared.Fleas) != 0 || p.Content.Archive[id].Content != "Owner content" {
		t.Fatal(p.Content)
	}
	put(t, root, "index.html", `<h1 flea="restored">{flea `+id+`}</h1>`)
	p, err = LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Content.Page("/").Fleas[id].Content != "Owner content" || p.Content.Page("/").Title != "Keep metadata" {
		t.Fatal(p.Content)
	}
	if _, ok := p.Content.Archive[id]; ok {
		t.Fatal("restored ID is still archived")
	}
	p, err = LoadEditing(context.Background(), root)
	if err != nil || p.FleasUpdated {
		t.Fatal("unchanged reload wrote content", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "_", ".fleas.lock")); !os.IsNotExist(err) {
		t.Fatal("local lock file exists", err)
	}
}

func TestUnquotedFleaContextsAndInvalidIDs(t *testing.T) {
	for _, source := range []string{
		`<p title="{flea a7Km2pQ9xB}">x</p>`, `<textarea>{flea a7Km2pQ9xB}</textarea>`,
		`<title>{flea a7Km2pQ9xB}</title>`, `{flea a7Km2pQ9xB}`,
		`<p>{flea a7Km2pQ9xB extra}</p>`, `<p>{flea a7Km{pQ9xB}</p>`, `<p>{flea}</p>`,
	} {
		root := fixture(t)
		store := fleas.NewStore()
		store.Page("/").Fleas["a7Km2pQ9xB"] = fleas.Entry{Content: "<em>Safe</em>"}
		if err := SaveContent(root, store); err != nil {
			t.Fatal(err)
		}
		put(t, root, "index.html", source)
		if _, err := Load(root); err == nil {
			t.Fatal("accepted", source)
		}
	}
	root := fixture(t)
	put(t, root, "index.html", `<p>{flea team}</p>`)
	p := mustLoad(t, root)
	body, err := p.render(p.Pages["/"], "/")
	if err != nil || !strings.Contains(string(body), "&lt;b&gt;") {
		t.Fatal(string(body), err)
	}
}
