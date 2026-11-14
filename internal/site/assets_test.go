package site

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestStreamedAssetRequests(t *testing.T) {
	root := fixture(t)
	body := bytes.Repeat([]byte("0123456789"), 10000)
	put(t, root, "large.txt", string(body))
	p := mustLoad(t, root)
	defer p.closeAssetFiles()
	if _, ok := p.AssetData["/large.txt"]; ok || len(p.assetFiles) != 1 {
		t.Fatal("large asset retained in heap or missing handle")
	}
	s := NewProductionServer(p, nil)
	for _, tc := range []struct {
		method, span string
		status       int
		body         []byte
	}{
		{"GET", "", 200, body},
		{"HEAD", "", 200, nil},
		{"GET", "bytes=7-29", 206, body[7:30]},
		{"GET", "bytes=999999-", 416, nil},
	} {
		r := httptest.NewRequest(tc.method, "/large.txt", nil)
		r.Header.Set("Range", tc.span)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != tc.status || (tc.status != 416 && !bytes.Equal(w.Body.Bytes(), tc.body)) {
			t.Fatalf("%s %s: status %d, bytes %d", tc.method, tc.span, w.Code, w.Body.Len())
		}
		if tc.status == 200 && w.Header().Get("Content-Length") != fmt.Sprint(len(body)) {
			t.Fatal("missing content length")
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := httptest.NewRequest("GET", "/large.txt", nil)
			r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", i, i+100))
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != 206 || !bytes.Equal(w.Body.Bytes(), body[i:i+101]) {
				t.Errorf("concurrent range %d failed", i)
			}
		}(i)
	}
	wg.Wait()
}

func TestStreamedAssetSnapshotAndMutation(t *testing.T) {
	root := fixture(t)
	old := bytes.Repeat([]byte("a"), memoryAssetLimit+1)
	put(t, root, "large.bin", string(old))
	p := mustLoad(t, root)
	defer p.closeAssetFiles()
	s := NewProductionServer(p, nil)
	put(t, root, "replacement.bin", string(bytes.Repeat([]byte("b"), len(old))))
	if err := os.Rename(filepath.Join(root, "site/replacement.bin"), filepath.Join(root, "site/large.bin")); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", "/large.bin"); !bytes.Equal(w.Body.Bytes(), old) {
		t.Fatal("replacement changed published bytes")
	}
	next := mustLoad(t, root)
	defer next.closeAssetFiles()
	s.Publish(next)
	if w := request(s, "GET", "/large.bin"); w.Code != 200 || w.Body.Bytes()[0] != 'b' {
		t.Fatal("new snapshot did not publish replacement")
	}
	put(t, root, "large.bin", string(bytes.Repeat([]byte("c"), len(old))))
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(filepath.Join(root, "site/large.bin"), later, later); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", "/large.bin"); w.Code != 503 {
		t.Fatal("modified file did not fail closed", w.Code)
	}
	if w := request(s, "HEAD", "/large.bin"); w.Code != 503 || w.Body.Len() != 0 {
		t.Fatal("HEAD mutation response", w.Code, w.Body.Len())
	}
	if w := request(s, "GET", "/"); w.Code != 200 {
		t.Fatal("unrelated page unavailable")
	}
}

func TestStreamedAssetReaderStopsAfterMutation(t *testing.T) {
	root := fixture(t)
	put(t, root, "large.bin", string(bytes.Repeat([]byte("a"), memoryAssetLimit+1)))
	p := mustLoad(t, root)
	defer p.closeAssetFiles()
	a := p.assetFiles["/large.bin"]
	b := make([]byte, 32)
	if n, err := a.ReadAt(b, 0); err != nil || n != len(b) {
		t.Fatal(n, err)
	}
	put(t, root, "large.bin", "new bytes")
	if n, err := a.ReadAt(b, 0); err == nil || n != 0 {
		t.Fatal("active reader accepted mutated bytes", n, err)
	}
}

func TestStreamedAssetDoesNotFollowReplacementSymlink(t *testing.T) {
	root := fixture(t)
	body := string(bytes.Repeat([]byte("a"), memoryAssetLimit+1))
	put(t, root, "large.bin", body)
	p := mustLoad(t, root)
	defer p.closeAssetFiles()
	file := filepath.Join(root, "site/large.bin")
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "itsy.config.json"), file); err != nil {
		t.Fatal(err)
	}
	if w := request(NewProductionServer(p, nil), "GET", "/large.bin"); w.Code != 200 || w.Body.String() != body {
		t.Fatal("request followed replacement path")
	}
	if _, err := Load(root); err == nil {
		t.Fatal("new snapshot accepted symlink")
	}
}

func TestStreamedAssetRejectsLinks(t *testing.T) {
	for _, link := range []func(string, string) error{os.Link, os.Symlink} {
		root := fixture(t)
		put(t, root, "source.bin", string(bytes.Repeat([]byte("a"), memoryAssetLimit+1)))
		if err := link(filepath.Join(root, "site/source.bin"), filepath.Join(root, "site/linked.bin")); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Fatal("linked asset accepted")
		}
	}
}

func TestManagedPhotoStreamsValidatedFile(t *testing.T) {
	root := fixture(t)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	put(t, root, "portrait.png", b.String())
	put(t, root, "index.html", `<img snip src="/portrait.png" alt="Photo">`)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	defer p.closeAssetFiles()
	for _, photo := range p.Content.ImageAssets() {
		route := "/media/" + photo.File
		if _, ok := p.AssetData[route]; ok || p.assetFiles[route].file == nil {
			t.Fatal("photo was not disk backed")
		}
		w := request(NewProductionServer(p, nil), "GET", route)
		if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
			t.Fatal("photo not served", w.Code)
		}
		if err := os.WriteFile(filepath.Join(root, "itsy/media", photo.File), []byte("corrupt"), 0644); err != nil {
			t.Fatal(err)
		}
		if w := request(NewProductionServer(p, nil), "GET", route); w.Code != 503 {
			t.Fatal("corrupt photo served")
		}
		if _, err := Load(root); err == nil {
			t.Fatal("corrupt photo loaded")
		}
	}
}
