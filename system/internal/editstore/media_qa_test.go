package editstore

import (
	"bytes"
	"image"
	"image/png"
	"itsy/internal/editor"
	"itsy/internal/media"
	"itsy/internal/site"
	"itsy/internal/snips"
	"os"
	"path/filepath"
	"testing"
)

func qaPhoto(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if e := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 8, 4))); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func seedQAPhoto(t *testing.T, s *Store) string {
	t.Helper()
	id := "Abcdef1234"
	photo, data, e := media.Process(qaPhoto(t), "Photo", "Original")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(s.mediaRoot(), 0750); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(s.mediaRoot(), photo.File), data, 0644); e != nil {
		t.Fatal(e)
	}
	d, e := s.Read()
	if e != nil {
		t.Fatal(e)
	}
	d.Store.Page("/").Snips[id] = snips.Entry{Type: "image", Image: &photo}
	if e = snips.WriteFile(s.ContentPath, d.Store); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(s.Root, "site/index.html"), []byte(`<img src="{snip Abcdef1234.src}" alt="{snip Abcdef1234.alt}">`), 0644); e != nil {
		t.Fatal(e)
	}
	return id
}
func TestQAPhotoSaveMetadataVersionsAndLocalReload(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, _ := fixture(t, remote)
			id := seedQAPhoto(t, s)
			d, e := s.Read()
			if e != nil {
				t.Fatal(e)
			}
			edit := editor.PageEdit{Version: d.Version, Route: "/", Media: []editor.MediaEdit{{ID: id, Name: "Changed", Alt: "Updated", Data: qaPhoto(t)}}}
			if e = s.Save(editor.Update{Page: &edit}); e != nil {
				t.Fatal(e)
			}
			if e = s.Save(editor.Update{Page: &edit}); e == nil {
				t.Fatal("accepted stale metadata")
			}
			next, e := s.Read()
			if e != nil || next.Version == d.Version || next.Store.Pages["/"].Snips[id].Image.Alt != "Updated" {
				t.Fatal("photo not versioned", e)
			}
			edit.Version = next.Version
			edit.Media[0].ID = "Unknown000"
			if e = s.Save(editor.Update{Page: &edit}); e == nil {
				t.Fatal("accepted unknown media")
			}
			if !remote {
				p, e := site.Load(s.Root)
				if e != nil || p.Content.Pages["/"].Snips[id].Image.Alt != "Updated" {
					t.Fatal("local reload lost media", e)
				}
			}
		})
	}
}
func TestQAPhotoSaveRejectsSymlinkRoot(t *testing.T) {
	s, _ := fixture(t, false)
	id := seedQAPhoto(t, s)
	d, e := s.Read()
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(s.mediaRoot(), s.mediaRoot()+"-original"); e != nil {
		t.Fatal(e)
	}
	outside := t.TempDir()
	if e = os.Symlink(outside, s.mediaRoot()); e != nil {
		t.Fatal(e)
	}
	edit := editor.PageEdit{Version: d.Version, Route: "/", Media: []editor.MediaEdit{{ID: id, Data: qaPhoto(t)}}}
	if e = s.Save(editor.Update{Page: &edit}); e == nil {
		t.Fatal("followed media directory symlink")
	}
	files, e := os.ReadDir(outside)
	if e != nil || len(files) != 0 {
		t.Fatal("wrote outside project", e)
	}
}
