package editstore

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/editor"
	"flea/internal/media"
	"flea/internal/snips"
	"flea/internal/video"
)

func TestAtomicMediaSnipSave(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, _ := fixture(t, remote)
			im := image.NewRGBA(image.Rect(0, 0, 2, 2))
			im.Set(0, 0, color.RGBA{R: 255, A: 255})
			var input bytes.Buffer
			png.Encode(&input, im)
			photo, data, err := media.Process(input.Bytes(), "Photo", "Original alt")
			if err != nil {
				t.Fatal(err)
			}
			if err = os.MkdirAll(s.mediaRoot(), 0750); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(s.mediaRoot(), photo.File), data, 0644); err != nil {
				t.Fatal(err)
			}
			content := snips.NewStore()
			content.Page("/").Snips["abcdefghij"] = snips.Entry{Content: "Original"}
			content.Page("/").Snips["Abcdef1234"] = snips.Entry{Type: "image", Image: &photo}
			content.Page("/other").Snips["Zbcdef1234"] = snips.Entry{Type: "image", Image: &photo}
			if err = snips.WriteFile(s.ContentPath, content); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(s.Root, "site/index.html"), []byte(`<p>{snip abcdefghij}</p><img src="{snip Abcdef1234.src}" alt="{snip Abcdef1234.alt}">`), 0644); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(s.Root, "site/other.html"), []byte(`<img src="{snip Zbcdef1234.src}" alt="{snip Zbcdef1234.alt}">`), 0644); err != nil {
				t.Fatal(err)
			}
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			edit := editor.PageEdit{Version: doc.Version, Route: "/", Title: "Changed", Changes: []editor.ContentEdit{{ID: "abcdefghij", Content: "Changed"}}, Media: []editor.MediaEdit{{ID: "Abcdef1234", Name: "Photo", Alt: "Updated"}}}
			before, _ := os.ReadFile(s.ContentPath)
			bad := edit
			bad.Media = append(append([]editor.MediaEdit{}, edit.Media...), editor.MediaEdit{ID: "Zbcdef1234"})
			if err = s.Save(editor.Update{Page: &bad}); err == nil {
				t.Fatal("cross-page media accepted")
			}
			after, _ := os.ReadFile(s.ContentPath)
			if !bytes.Equal(before, after) {
				t.Fatal("failed save changed content")
			}
			bad = edit
			bad.Media = []editor.MediaEdit{{ID: "Abcdef1234", Data: []byte("not an image")}}
			if err = s.Save(editor.Update{Page: &bad}); err == nil {
				t.Fatal("invalid image accepted")
			}
			if err = s.Save(editor.Update{Page: &edit}); err != nil {
				t.Fatal(err)
			}
			if err = s.Save(editor.Update{Page: &edit}); err == nil {
				t.Fatal("stale media accepted")
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if doc.Store.Pages["/"].Snips["Abcdef1234"].Image.Alt != "Updated" || doc.Store.Pages["/"].Snips["abcdefghij"].Content != "Changed" {
				t.Fatal("incomplete atomic save")
			}
		})
	}
}

func TestVideoSnipSaveAndPoster(t *testing.T) {
	s, _ := fixture(t, false)
	data, err := os.ReadFile("../video/testdata/valid.mp4")
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := video.Inspect(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(s.mediaRoot(), 0750); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(s.mediaRoot(), metadata.File), data, 0644); err != nil {
		t.Fatal(err)
	}
	content := snips.NewStore()
	content.Page("/").Snips["Abcdef1234"] = snips.Entry{Type: "video", Video: &video.Video{Metadata: metadata, Name: "Intro"}}
	if err = snips.WriteFile(s.ContentPath, content); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(s.Root, "site/index.html"), []byte(`<video controls src="{snip Abcdef1234.src}" poster="{snip Abcdef1234.poster}"></video>`), 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	var poster bytes.Buffer
	png.Encode(&poster, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	edit := editor.PageEdit{Version: doc.Version, Route: "/", Media: []editor.MediaEdit{{ID: "Abcdef1234", Name: "Updated", PosterData: poster.Bytes(), Data: data}}}
	if err = s.Save(editor.Update{Page: &edit}); err != nil {
		t.Fatal(err)
	}
	doc, err = s.Read()
	if err != nil {
		t.Fatal(err)
	}
	entry := doc.Store.Pages["/"].Snips["Abcdef1234"]
	if entry.Video.Name != "Updated" || entry.Poster == nil {
		t.Fatal("video/poster not saved")
	}
	edit.Version = doc.Version
	edit.Media = []editor.MediaEdit{{ID: "Abcdef1234", Name: "Updated", ClearPoster: true}}
	if err = s.Save(editor.Update{Page: &edit}); err != nil {
		t.Fatal(err)
	}
	doc, err = s.Read()
	if err != nil || doc.Store.Pages["/"].Snips["Abcdef1234"].Poster != nil {
		t.Fatal("poster not cleared", err)
	}
}

func TestLocalExtractionRecoveryBlocksSaves(t *testing.T) {
	s, _ := fixture(t, false)
	doc, err := s.Read()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(s.Root, "flea/runtime/.snips-pending.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.Save(editor.Update{Version: doc.Version, Route: "/", Title: "Blocked"}); err == nil {
		t.Fatal("pending extraction did not block save")
	}
}

func TestMediaSnipRetainedQuotaAndHardlinks(t *testing.T) {
	for _, kind := range []string{"quota", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			s, _ := fixture(t, false)
			if err := os.MkdirAll(s.mediaRoot(), 0750); err != nil {
				t.Fatal(err)
			}
			photo, data, err := media.Process(qaPhoto(t), "", "")
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(s.mediaRoot(), strings.Repeat("a", 64)+".png")
			if kind == "quota" {
				f, err := os.Create(name)
				if err != nil {
					t.Fatal(err)
				}
				err = f.Truncate(media.MaxTotal)
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				outside := filepath.Join(t.TempDir(), "photo.png")
				if err = os.WriteFile(outside, data, 0644); err != nil {
					t.Fatal(err)
				}
				if err = os.Link(outside, name); err != nil {
					t.Fatal(err)
				}
			}
			if err = s.saveSnipBytes(photo.File, data, false); err == nil {
				t.Fatal("unsafe or over-quota write accepted")
			}
			if _, err = os.Stat(filepath.Join(s.mediaRoot(), photo.File)); !os.IsNotExist(err) {
				t.Fatal("rejected bytes were written", err)
			}
		})
	}
}
