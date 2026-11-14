package editstore

import (
	"itsy/internal/editor"
	"itsy/internal/media"
	"itsy/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func TestQAMultisnipPhotoSelectionValidation(t *testing.T) {
	s, _ := fixture(t, false)
	source := `{#multisnip team}<article><img data-snip-photo="photo" data-photo-id="{snip.photo.id}" src="{snip.photo.src}" alt="{snip.photo.alt}"><h2>{snip.name}</h2></article>{/multisnip}`
	if e := os.WriteFile(filepath.Join(s.Root, "site", "index.html"), []byte(source), 0644); e != nil {
		t.Fatal(e)
	}
	d, e := s.Read()
	if e != nil {
		t.Fatal(e)
	}
	photo, data, e := media.Process(qaPhoto(t), "Photo", "")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(s.mediaRoot(), 0750); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(s.mediaRoot(), photo.File), data, 0644); e != nil {
		t.Fatal(e)
	}
	d.Store.Media = map[string]media.Photo{"Abcdef1234": photo}
	if e = snips.WriteFile(s.ContentPath, d.Store); e != nil {
		t.Fatal(e)
	}
	d, e = s.Read()
	if e != nil {
		t.Fatal(e)
	}
	id := ""
	for key := range d.Store.Media {
		id = key
	}
	for _, value := range []string{"unknown000", "<img src=x onerror=alert(1)>", "javascript:alert(1)", id, ""} {
		d, e = s.Read()
		if e != nil {
			t.Fatal(e)
		}
		edit := editor.PageEdit{Version: d.Version, Route: "/", Multisnips: map[string][]snips.MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Joe", "photo": value}}}}}
		e = s.Save(editor.Update{Page: &edit})
		valid := value == id || value == ""
		if (e == nil) != valid {
			t.Fatalf("photo value %q: %v", value, e)
		}
	}
}
