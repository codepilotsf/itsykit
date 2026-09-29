package editstore

import (
	"flea/internal/editor"
	"flea/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func TestCollectionPhotoUpload(t *testing.T) {
	for _, mode := range []string{"save", "invalid", "unreferenced", "hidden", "duplicate", "stale"} {
		t.Run(mode, func(t *testing.T) {
			s, _ := fixture(t, false)
			source := `{#multisnip team}<article><img data-snip-photo="photo" data-photo-id="{snip.photo.id}" src="{snip.photo.src}" alt="{snip.photo.alt}"><h2>{snip.name}</h2></article>{/multisnip}`
			if err := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			d, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			id := "Abcdef1234"
			edit := editor.PageEdit{Version: d.Version, Route: "/", Multisnips: map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Joe", "photo": id}}}}, Media: []editor.MediaEdit{{ID: id, Library: true, Name: "Portrait", Alt: "Joe", Data: qaPhoto(t)}}}
			switch mode {
			case "invalid":
				edit.Media[0].Data = []byte("not a photo")
			case "unreferenced":
				edit.Multisnips["team"][0].Fields["photo"] = ""
			case "hidden":
				edit.Multisnips["hidden"] = edit.Multisnips["team"]
				delete(edit.Multisnips, "team")
			case "duplicate":
				edit.Media = append(edit.Media, edit.Media[0])
			case "stale":
				edit.Version = "stale"
			}
			err = s.Save(editor.Update{Page: &edit})
			if (err == nil) != (mode == "save") {
				t.Fatalf("unexpected result: %v", err)
			}
			after, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if mode != "save" {
				if after.Version != d.Version {
					t.Fatal("failed upload published content")
				}
				return
			}
			photo := after.Store.Media[id]
			if photo.Alt != "Joe" || photo.File == "" || after.Store.Pages["/"].Multisnips["team"][0].Fields["photo"] != id {
				t.Fatal("missing photo metadata or binding")
			}
			if _, err = os.Stat(filepath.Join(s.mediaRoot(), photo.File)); err != nil {
				t.Fatal(err)
			}
			// Reuse the uploaded library photo without re-uploading its bytes.
			edit.Version = after.Version
			edit.Media = nil
			if err = s.Save(editor.Update{Page: &edit}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
