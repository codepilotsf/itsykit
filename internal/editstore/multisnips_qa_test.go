package editstore

import (
	"flea/internal/editor"
	"flea/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func TestQAMultisnipReorderRetainsRemovedField(t *testing.T) {
	s, _ := fixture(t, false)
	if err := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(`{#multisnip team}<article><h2>{snip.name}</h2></article>{/multisnip}`), 0644); err != nil {
		t.Fatal(err)
	}
	store := snips.NewStore()
	store.Page("/").Multisnips = map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Joe", "bio": "Keep this removed field"}}}}
	data, _ := store.Bytes()
	if err := os.WriteFile(s.ContentPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	edit := editor.PageEdit{Route: "/", Version: doc.Version, Multisnips: map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Updated"}}}}}
	if err := s.Save(editor.Update{Page: &edit}); err != nil {
		t.Fatal(err)
	}
	after, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	if after.Store.Pages["/"].Multisnips["team"][0].Fields["bio"] != "Keep this removed field" {
		t.Fatal("collection edit silently discarded developer-removed field")
	}
}

func TestQAMultisnipRejectsHiddenGroupEdit(t *testing.T) {
	s, _ := fixture(t, false)
	source := `{#if app.never}{#multisnip team}<article><h2>{snip.name}</h2></article>{/multisnip}{/if}`
	if err := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	store := snips.NewStore()
	store.Page("/").Multisnips = map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Joe"}}}}
	data, _ := store.Bytes()
	if err := os.WriteFile(s.ContentPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	edit := editor.PageEdit{Route: "/", Version: doc.Version, Multisnips: map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Hidden edit"}}}}}
	if err := s.Save(editor.Update{Page: &edit}); err == nil {
		t.Fatal("hidden multisnip accepted browser edits")
	}
}
