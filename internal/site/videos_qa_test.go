package site

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"flea/internal/snips"
	"flea/internal/video"
)

func qaVideoSite(t *testing.T) (string, video.Video, []byte) {
	t.Helper()
	root := fixture(t)
	b, e := os.ReadFile("../video/testdata/valid.mp4")
	if e != nil {
		t.Fatal(e)
	}
	m, e := video.Inspect(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "flea", "media")
	if e = os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, m.File), b, 0644); e != nil {
		t.Fatal(e)
	}
	s := snips.NewStore()
	s.Videos = map[string]video.Video{"Abcdefghij": {Metadata: m, Name: "Clip"}}
	if e = SaveContent(root, s); e != nil {
		t.Fatal(e)
	}
	return root, s.Videos["Abcdefghij"], b
}
func TestQAVideoCatalogueOnlyAndSafeURLs(t *testing.T) {
	root, v, _ := qaVideoSite(t)
	other, e := os.ReadFile("../video/testdata/silent.mp4")
	if e != nil {
		t.Fatal(e)
	}
	m, e := video.Inspect(bytes.NewReader(other), int64(len(other)))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(root, "flea", "media", m.File), other, 0644); e != nil {
		t.Fatal(e)
	}
	p := mustLoad(t, root)
	server := NewProductionServer(p, nil)
	for _, path := range []string{"/media/" + m.File, "/media/../" + v.File, "/media/" + v.File + "/", "/media/%2e%2e/" + v.File} {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest("GET", "http://example.test"+path, nil))
		if w.Code != 404 {
			t.Fatal("served uncatalogued or alias URL", path, w.Code)
		}
	}
	if len(p.videos) != 1 {
		t.Fatal("uncatalogued files entered snapshot")
	}
	for route := range p.AssetData {
		if bytes.Contains([]byte(route), []byte("/media/")) {
			t.Fatal("video buffered in assets")
		}
	}
}
func TestQAVideoRejectsSymlinkAndHardlink(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "directory symlink"} {
		t.Run(kind, func(t *testing.T) {
			root, v, b := qaVideoSite(t)
			dir := filepath.Join(root, "flea", "media")
			file := filepath.Join(dir, v.File)
			outside := filepath.Join(t.TempDir(), v.File)
			if e := os.WriteFile(outside, b, 0644); e != nil {
				t.Fatal(e)
			}
			if e := os.Remove(file); e != nil {
				t.Fatal(e)
			}
			var e error
			switch kind {
			case "symlink":
				e = os.Symlink(outside, file)
			case "hardlink":
				e = os.Link(outside, file)
			case "directory symlink":
				if e = os.Remove(dir); e == nil {
					e = os.Symlink(filepath.Dir(outside), dir)
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			if f, e := OpenVideo(filepath.Join(root, "flea", "media"), v.File); e == nil {
				f.Close()
				t.Fatal("accepted linked video")
			}
			if _, e := Load(root); e == nil {
				t.Fatal("loaded linked catalogue video")
			}
		})
	}
}

func TestQASyncSourceMissingVideosCannotServe(t *testing.T) {
	root, v, _ := qaVideoSite(t)
	if err := os.Remove(filepath.Join(root, "flea", "media", v.File)); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("regular loader accepted missing video")
	}
	p, err := LoadSyncSource(root)
	if err != nil {
		t.Fatal(err)
	}
	server := NewProductionServer(p, nil)
	for _, path := range []string{"/", "/media/" + v.File} {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest("GET", "http://example.test"+path, nil))
		if w.Code != 503 {
			t.Fatal("sync-only snapshot served", path, w.Code)
		}
	}
}

func TestQAInstallVideosCountsRetainedStorage(t *testing.T) {
	root, v, _ := qaVideoSite(t)
	p := mustLoad(t, root)
	dest := filepath.Join(t.TempDir(), "media")
	if err := os.MkdirAll(filepath.Join(dest, "videos"), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dest, "videos", "retained.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(video.MaxTotal); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	if err = InstallVideos(filepath.Join(root, "flea", "media"), dest, p.Content); err == nil {
		t.Fatal("ignored retained video quota")
	}
	if _, err = os.Stat(filepath.Join(dest, "videos", v.File)); !os.IsNotExist(err) {
		t.Fatal("wrote video beyond quota", err)
	}
}
