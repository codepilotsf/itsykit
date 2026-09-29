package site

import (
	"flea/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func TestQAMultisnipExtractionIsRepeatableAndPreservesEmpty(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<article multisnip="team"><h2 snip="name">Joe</h2><div snip="bio"><p>Hello</p></div></article><p snip>Other</p>`)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if string(first) != string(second) {
		t.Fatal("second extraction changed content")
	}
	store, err := snips.Decode(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Pages["/"].Multisnips["team"]) != 1 || len(store.Pages["/"].Snips) != 1 {
		t.Fatal("lost seed or ordinary snip")
	}
	store.Pages["/"].Multisnips["team"] = []snips.MultiItem{}
	data, _ := store.Bytes()
	put(t, root, "flea/snips.json", string(data))
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	if items, ok := p.Content.Pages["/"].Multisnips["team"]; !ok || len(items) != 0 {
		t.Fatal("empty collection reseeded or lost")
	}
	defs, err := p.MultisnipDefinitions("/")
	if err != nil || len(defs["team"].Fields) != 2 {
		t.Fatal("empty collection lost schema", err)
	}
}

func TestQAMultisnipRejectsNoncontentAncestors(t *testing.T) {
	for _, tag := range []string{"nav", "button", "form", "a", "template"} {
		t.Run(tag, func(t *testing.T) {
			_, err := definition("<" + tag + "><span>{snip.name}</span></" + tag + ">")
			if err == nil {
				t.Errorf("accepted editable field inside <%s>", tag)
			}
		})
	}
}

func TestQAMultisnipRejectsTemplateAncestor(t *testing.T) {
	_, err := multisnipBlocks("index.html", `<template>{#multisnip team}<article><h2>{snip.name}</h2></article>{/multisnip}</template>`)
	if err == nil {
		t.Fatal("multisnip inside inert template accepted")
	}
}

func TestQAMultisnipDuplicateAuthoringDoesNotRewriteSource(t *testing.T) {
	root := fixture(t)
	source := `{#multisnip team}<article><h2>{snip.name}</h2></article>{/multisnip}<article multisnip="team"><h2 snip="name">New content must not be discarded</h2></article>`
	put(t, root, "index.html", source)
	store := snips.NewStore()
	store.Page("/").Multisnips = map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Existing"}}}}
	data, _ := store.Bytes()
	put(t, root, "flea/snips.json", string(data))
	if err := ExtractSnips(root); err == nil {
		t.Error("duplicate compiled and authoring collection accepted")
	}
	after, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if string(after) != source {
		t.Fatal("invalid extraction consumed new seed text")
	}
}
