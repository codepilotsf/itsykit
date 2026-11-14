package site

import (
	"bytes"
	"image"
	"image/png"
	"itsy/internal/snips"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypedMediaSnips(t *testing.T) {
	root := fixture(t)
	var img bytes.Buffer
	png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 8, 4)))
	os.MkdirAll(filepath.Join(root, "site/images"), 0755)
	os.WriteFile(filepath.Join(root, "site/images/a.png"), img.Bytes(), 0644)
	clip, err := os.ReadFile("../video/testdata/valid.mp4")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "site/clip.mp4"), clip, 0644)
	put(t, root, "index.html", `<img snip src="/images/a.png" alt="Our &quot;team&quot;" class="rounded-lg"><video snip src="/clip.mp4" poster="/images/a.png" controls></video>`)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	entries := p.Content.Pages["/"].Snips
	if len(entries) != 2 {
		t.Fatalf("entries: %+v", entries)
	}
	source, _ := os.ReadFile(filepath.Join(root, "site/index.html"))
	if strings.Contains(string(source), " snip ") || !strings.Contains(string(source), `class="rounded-lg"`) {
		t.Fatal(string(source))
	}
	visible, err := p.EditableMedia("/")
	if err != nil || len(visible) != 2 {
		t.Fatalf("editable: %v %v", visible, err)
	}
	public, err := p.renderPage(p.Pages["/"], "/", nil, false)
	if err != nil || strings.Contains(string(public), "data-snip-media") || !strings.Contains(string(public), "/media/") {
		t.Fatalf("public %s %v", public, err)
	}
	before, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	if err = ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("not idempotent")
	}
}

func TestTypedMediaWrongTypeDoesNotPublishExtraction(t *testing.T) {
	root := fixture(t)
	store := snips.NewStore()
	store.Page("/").Snips["Abcdefghij"] = snips.Entry{Content: "Text"}
	if err := SaveContent(root, store); err != nil {
		t.Fatal(err)
	}
	source := `<img src="{snip Abcdefghij.src}" alt="{snip Abcdefghij.alt}"><p snip>New content</p>`
	put(t, root, "index.html", source)
	before, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	if err := ExtractSnips(root); err == nil {
		t.Fatal("accepted wrong media type")
	}
	after, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	template, _ := os.ReadFile(filepath.Join(root, "site/index.html"))
	if !bytes.Equal(before, after) || string(template) != source {
		t.Fatal("published invalid extraction")
	}
}
func TestTypedMediaRejectNesting(t *testing.T) {
	for _, source := range []string{`<div snip><img snip src="a.png"></div>`, `<video snip src="a.mp4"><p snip>Text</p></video>`, `<video snip src="a.mp4"><source src="b.mp4"></video>`, `<p>{snip Abcdefghij.src}</p>`, `<img src="{snip Abcdefghij.src}" alt="fixed">`} {
		if _, err := scanMediaBindings("index.html", source); err == nil {
			t.Fatalf("accepted %s", source)
		}
	}
}

func TestTypedMediaReservedMarker(t *testing.T) {
	if _, err := scanMediaBindings("index.html", `<video snip src="/clip.mp4"/>`); err == nil {
		t.Fatal("accepted self-closing video")
	}
	for _, source := range []string{`<img src="/x.png" data-snip-media="Abcdefghij">`, `<div data-snip-media="Abcdefghij">Text</div>`, `<div inert><img snip src="/x.png"></div>`} {
		if _, err := scanMediaBindings("index.html", source); err == nil {
			t.Fatalf("accepted reserved/inert binding: %s", source)
		}
	}
}
