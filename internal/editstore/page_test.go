package editstore

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"flea/internal/editor"
	"flea/internal/snips"
)

func TestAtomicPageSave(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, server := fixture(t, remote)
			put := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(s.Root, "site", name), []byte(body), 0644); err != nil {
					t.Fatal(err)
				}
			}
			put("index.html", `<article>{snip abcdefghij}</article><h2>{snip klmnopqrst}</h2><article>{snip uvwxyzABCD}</article><p>{page.title}</p>`)
			put("about.html", `<article>{snip uvwxyzABCD}</article><article>{snip EFGHIJKLMN}</article>`)
			store := snips.NewStore()
			store.Page("/").Snips["abcdefghij"] = snips.Entry{Content: "<p>Original</p>"}
			store.Page("/").Snips["klmnopqrst"] = snips.Entry{Content: "Heading"}
			store.Shared.Snips["uvwxyzABCD"] = snips.Entry{Content: "<p>Shared</p>"}
			store.Page("/about").Snips["EFGHIJKLMN"] = snips.Entry{Content: "<p>Other page</p>"}
			data, _ := store.Bytes()
			if err := os.WriteFile(s.ContentPath, data, 0644); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			edit := editor.PageEdit{Version: doc.Version, Route: "/", Title: "Published title", Description: "Published description", Changes: []editor.ContentEdit{{ID: "abcdefghij", Content: `<h1 style="text-align: center">Title</h1><p><a href="https://example.com" target="_blank">Link</a></p><script>bad</script>`}, {ID: "klmnopqrst", Content: "New heading"}, {ID: "uvwxyzABCD", Content: "<p>New shared text</p>"}}}
			before, _ := os.ReadFile(s.ContentPath)
			edit.Changes[1].Content = `<em>New</em> <a href="https://example.com">heading</a>`
			bad := edit
			bad.Changes = append(append([]editor.ContentEdit{}, edit.Changes...), editor.ContentEdit{ID: "EFGHIJKLMN", Content: "cross-page"})
			if err := s.Save(editor.Update{Page: &bad}); err == nil {
				t.Fatal("cross-page save accepted")
			}
			after, _ := os.ReadFile(s.ContentPath)
			if string(before) != string(after) {
				t.Fatal("failed batch partially persisted")
			}
			bad = edit
			bad.Changes = []editor.ContentEdit{{ID: "klmnopqrst", Content: "<h1>Nested block</h1>"}}
			if err := s.Save(editor.Update{Page: &bad}); err == nil {
				t.Fatal("block inside heading accepted")
			}
			bad = edit
			bad.Changes = append(append([]editor.ContentEdit{}, edit.Changes...), edit.Changes[0])
			if err := s.Save(editor.Update{Page: &bad}); err == nil {
				t.Fatal("duplicate accepted")
			}
			var receipt editor.PageResult
			if err := s.Save(editor.Update{Page: &edit, Result: &receipt}); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(editor.Update{Page: &edit}); err == nil {
				t.Fatal("stale batch accepted")
			}
			got, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if got.Store.Pages["/"].Title != edit.Title || got.Store.Pages["/"].Description != edit.Description || got.Store.Shared.Snips["uvwxyzABCD"].Content != "<p>New shared text</p>" {
				t.Fatal("incomplete save")
			}
			for _, route := range []string{"/", "/about"} {
				w := httptest.NewRecorder()
				server.ServeHTTP(w, httptest.NewRequest("GET", route, nil))
				if !strings.Contains(w.Body.String(), "New shared text") || strings.Contains(w.Body.String(), "<script>bad") {
					t.Fatal(w.Body.String())
				}
			}
			content := got.Store.Pages["/"].Snips["abcdefghij"].Content
			if receipt.Page.Version != got.Version || receipt.Page.Version == edit.Version || receipt.Page.Changes[0].Content != content {
				t.Fatal("receipt does not describe the committed canonical snapshot", receipt)
			}
			if !strings.Contains(content, `<h1 style="text-align: center;">`) || !strings.Contains(content, `target="_blank" rel="noopener noreferrer"`) {
				t.Fatal(content)
			}
			// A template-only edit must invalidate a draft even if content is unchanged.
			edit.Version = got.Version
			put("index.html", `<article>{snip abcdefghij}</article>`)
			if err := s.Save(editor.Update{Page: &edit}); err == nil {
				t.Fatal("template-stale batch accepted")
			}
			// Local restart / remote reread retains the entire publication.
			persisted, _ := os.ReadFile(s.ContentPath)
			var saved snips.Store
			if json.Unmarshal(persisted, &saved) != nil || saved.Pages["/"].Title != edit.Title {
				t.Fatal("save not durable")
			}
		})
	}
}

func TestPageDetailsWithoutRegions(t *testing.T) {
	s, _ := fixture(t, false)
	if err := os.WriteFile(filepath.Join(s.Root, "site", "plain.html"), []byte("<p>Developer content</p>"), 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	edit := editor.PageEdit{Version: doc.Version, Route: "/plain", Title: "Plain page", Description: "Details only"}
	if err := s.Save(editor.Update{Page: &edit}); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentPageSavesDoNotOverwrite(t *testing.T) {
	for _, remote := range []bool{false, true} {
		s, _ := fixture(t, remote)
		doc, err := s.Read()
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for _, title := range []string{"First", "Second"} {
			wg.Add(1)
			go func(title string) {
				defer wg.Done()
				results <- s.Save(editor.Update{Page: &editor.PageEdit{Version: doc.Version, Route: "/", Title: title}})
			}(title)
		}
		wg.Wait()
		close(results)
		success := 0
		for err := range results {
			if err == nil {
				success++
			}
		}
		if success != 1 {
			t.Fatal("concurrent draft overwritten", success)
		}
	}
}

func TestPageSaveRejectsNoneditableReferences(t *testing.T) {
	for _, source := range []string{
		`<nav><span>{snip abcdefghij}</span></nav>`,
		`<section>Fixed text {snip abcdefghij}</section>`,
		`{#if site.missing}<article>{snip abcdefghij}</article>{/if}`,
		`<button>{snip abcdefghij}</button>`,
	} {
		t.Run(source, func(t *testing.T) {
			s, _ := fixture(t, false)
			if err := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			edit := editor.PageEdit{Version: doc.Version, Route: "/", Title: "Must not save", Changes: []editor.ContentEdit{{ID: "abcdefghij", Content: "Changed"}}}
			if err := s.Save(editor.Update{Page: &edit}); err == nil {
				t.Fatal("noneditable region accepted")
			}
			after, err := s.Read()
			if err != nil || after.Version != doc.Version {
				t.Fatal("failed save changed data", err)
			}
		})
	}
}
