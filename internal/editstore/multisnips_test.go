package editstore

import (
	"flea/internal/editor"
	"flea/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func TestMultisnipPageSave(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, _ := fixture(t, remote)
			if err := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(`{#multisnip team}<article><h2>{snip.name}</h2><div>{snip.bio}</div></article>{/multisnip}`), 0644); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			edit := editor.PageEdit{Version: doc.Version, Route: "/", Multisnips: map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "New", "bio": "<p>Hello</p><script>bad</script>"}}}}}
			if err := s.Save(editor.Update{Page: &edit}); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(editor.Update{Page: &edit}); err == nil {
				t.Fatal("stale version accepted")
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if doc.Store.Pages["/"].Multisnips["team"][0].Fields["bio"] != "<p>Hello</p>" {
				t.Fatal("not sanitized")
			}
			edit.Version = doc.Version
			for _, bad := range []map[string][]snips.MultiItem{
				{"other": {}},
				{"team": {{ID: "a123456789", Fields: map[string]string{"name": "N", "bio": "B", "unknown": "X"}}}},
				{"team": {{ID: "a123456789", Fields: map[string]string{"name": "<p>Block</p>", "bio": "B"}}}},
				{"team": {{ID: "a123456789", Fields: map[string]string{"name": "N", "bio": "B"}}, {ID: "a123456789", Fields: map[string]string{"name": "N", "bio": "B"}}}},
			} {
				edit.Multisnips = bad
				if err := s.Save(editor.Update{Page: &edit}); err == nil {
					t.Fatal("invalid change accepted", bad)
				}
			}
			edit.Multisnips = map[string][]snips.MultiItem{"team": {}}
			if err := s.Save(editor.Update{Page: &edit}); err != nil {
				t.Fatal(err)
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if items, exists := doc.Store.Pages["/"].Multisnips["team"]; !exists || len(items) != 0 {
				t.Fatal("empty group lost")
			}
		})
	}
}
