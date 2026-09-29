package site

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestQAMediaArchiveRestoreAndSharedBinding(t *testing.T) {
	root := fixture(t)
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site/portrait.png"), pngData.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", `<img snip src="/portrait.png" alt="Original">`)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	bound, err := os.ReadFile(filepath.Join(root, "site/index.html"))
	if err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	var id string
	for key := range p.Content.Pages["/"].Snips {
		id = key
	}
	put(t, root, "index.html", `<p>No photo</p>`)
	if err = ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p = mustLoad(t, root)
	if p.Content.Archive[id].Image == nil || len(p.Content.ImageAssets()) != 1 {
		t.Fatal("archived photo or retained asset lost")
	}
	put(t, root, "index.html", string(bound))
	put(t, root, "other.html", string(bound))
	if err = ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p = mustLoad(t, root)
	if p.Content.Shared.Snips[id].Image == nil || p.Content.Archive[id].Image != nil {
		t.Fatal("reintroduced shared image not restored")
	}
	for _, route := range []string{"/", "/other"} {
		editable, err := p.EditableMedia(route)
		if err != nil || editable[id] != "image" {
			t.Fatalf("shared media on %s: %v %v", route, editable, err)
		}
	}
}

func TestQAMediaHiddenConditionalNotEditable(t *testing.T) {
	root := fixture(t)
	var pngData bytes.Buffer
	png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if err := os.WriteFile(filepath.Join(root, "site/portrait.png"), pngData.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", `{#if app.missing}<img snip src="/portrait.png" alt="Hidden">{/if}`)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	editable, err := p.EditableMedia("/")
	if err != nil || len(editable) != 0 {
		t.Fatalf("hidden media editable: %v %v", editable, err)
	}
}
