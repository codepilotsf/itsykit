package editstore

import (
	"bytes"
	"flea/internal/editor"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/video"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func qaVideo(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("../video/testdata/" + name + ".mp4")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestQAVideoSaveReplacePosterAndRange(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, public := fixture(t, remote)
			b := qaVideo(t, "valid")
			id := "Abcdef1234"
			meta, e := video.Inspect(bytes.NewReader(b), int64(len(b)))
			if e != nil {
				t.Fatal(e)
			}
			if e = os.MkdirAll(s.mediaRoot(), 0750); e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(s.mediaRoot(), meta.File), b, 0644); e != nil {
				t.Fatal(e)
			}
			d, e := s.Read()
			if e != nil {
				t.Fatal(e)
			}
			d.Store.Page("/").Snips[id] = snips.Entry{Type: "video", Video: &video.Video{Metadata: meta, Name: "Clip"}}
			if e = snips.WriteFile(s.ContentPath, d.Store); e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(s.Root, "site/index.html"), []byte(`<video controls src="{snip Abcdef1234.src}" poster="{snip Abcdef1234.poster}"></video>`), 0644); e != nil {
				t.Fatal(e)
			}
			d, e = s.Read()
			if e != nil {
				t.Fatal(e)
			}
			edit := editor.PageEdit{Version: d.Version, Route: "/", Media: []editor.MediaEdit{{ID: id, Name: "Clip", PosterData: qaPhoto(t)}}}
			if e = s.Save(editor.Update{Page: &edit}); e != nil {
				t.Fatal(e)
			}
			if e = s.Save(editor.Update{Page: &edit}); e == nil {
				t.Fatal("accepted stale edit")
			}
			d, e = s.Read()
			if e != nil {
				t.Fatal(e)
			}
			entry := d.Store.Pages["/"].Snips[id]
			v := entry.Video
			if entry.Poster == nil || v.Width != 32 || v.Height != 24 {
				t.Fatal("poster or dimensions missing")
			}
			edit.Version = d.Version
			edit.Media = []editor.MediaEdit{{ID: id, Name: "Bad", Data: qaVideo(t, "wrong-codec")}}
			if e = s.Save(editor.Update{Page: &edit}); e == nil {
				t.Fatal("unsupported codec accepted")
			}
			edit.Media = []editor.MediaEdit{{ID: id, PosterData: []byte("invalid image")}}
			if e = s.Save(editor.Update{Page: &edit}); e == nil {
				t.Fatal("invalid poster accepted")
			}
			url := "http://example.test/media/" + v.File
			get := httptest.NewRequest("GET", url, nil)
			get.Header.Set("Range", "bytes=0-15")
			w := httptest.NewRecorder()
			public.ServeHTTP(w, get)
			if w.Code != 206 || !bytes.Equal(w.Body.Bytes(), b[:16]) || !strings.HasPrefix(w.Header().Get("Content-Range"), "bytes 0-15/") {
				t.Fatal("range failed", w.Code, w.Header())
			}
			w = httptest.NewRecorder()
			public.ServeHTTP(w, httptest.NewRequest("HEAD", url, nil))
			if w.Code != 200 || w.Body.Len() != 0 || w.Header().Get("Content-Type") != "video/mp4" {
				t.Fatal("HEAD failed", w.Code)
			}
			p, e := site.LoadWithMedia(s.Root, d.Store, nil, s.mediaRoot())
			if e != nil {
				t.Fatal(e)
			}
			if _, ok := p.AssetData["/media/"+v.File]; ok {
				t.Fatal("video copied into AssetData")
			}
			edit.Media = []editor.MediaEdit{{ID: id, Name: "Replaced", Data: qaVideo(t, "silent")}}
			if e = s.Save(editor.Update{Page: &edit}); e != nil {
				t.Fatal(e)
			}
			next, e := s.Read()
			if e != nil || next.Store.Pages["/"].Snips[id].Video.File == v.File || next.Store.Pages["/"].Snips[id].Video.Name != "Replaced" {
				t.Fatal("replacement did not retain ID", e)
			}
			if _, e = os.Stat(filepath.Join(s.mediaRoot(), v.File)); e != nil {
				t.Fatal("old video was removed", e)
			}
		})
	}
}
