package site

import (
	"bytes"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQAMultisnipPhotoExtractionAndEscaping(t *testing.T) {
	root := fixture(t)
	var b bytes.Buffer
	if e := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 8, 4))); e != nil {
		t.Fatal(e)
	}
	if e := os.MkdirAll(filepath.Join(root, "site", "images"), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "site", "images", "person.png"), b.Bytes(), 0644); e != nil {
		t.Fatal(e)
	}
	put(t, root, "index.html", `<article multisnip="team"><img snip="photo" src="/images/person.png" alt="&quot; onerror=&quot;alert(1)"><h2 snip="name">Joe</h2></article>`)
	if e := ExtractSnips(root); e != nil {
		t.Fatal(e)
	}
	p := mustLoad(t, root)
	items := p.Content.Pages["/"].Multisnips["team"]
	if len(items) != 1 || items[0].Fields["photo"] == "" || len(p.Content.Media) != 1 {
		t.Fatal("photo not extracted")
	}
	defs, e := p.MultisnipDefinitions("/")
	if e != nil || !defs["team"].Photos["photo"] {
		t.Fatal("missing typed photo schema", e)
	}
	response := httptest.NewRecorder()
	NewProductionServer(p, nil).ServeHTTP(response, httptest.NewRequest("GET", "http://example.test/", nil))
	if response.Code != 200 || strings.Contains(response.Body.String(), `alt="" onerror=`) || !strings.Contains(response.Body.String(), `&#34; onerror=&#34;alert(1)`) {
		t.Fatal("photo alt not contextually escaped", response.Code, response.Body.String())
	}
	source, e := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(source), `snip="photo"`) || !strings.Contains(string(source), `src="{snip.photo.src}"`) {
		t.Fatal("incorrect rewritten image", string(source))
	}
	before, e := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = ExtractSnips(root); e != nil {
		t.Fatal(e)
	}
	after, _ := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("photo extraction not idempotent")
	}
}

func TestQAMultisnipPhotoInvalidContexts(t *testing.T) {
	photo := `<img data-snip-photo="photo" data-photo-id="{snip.photo.id}" src="{snip.photo.src}" alt="{snip.photo.alt}">`
	for _, body := range []string{
		`<article><a>` + photo + `</a><h2>{snip.name}</h2></article>`,
		`<article><img src="{snip.photo.src}" alt="{snip.photo.alt}"><h2>{snip.name}</h2></article>`,
		`<article>` + strings.Replace(photo, `alt="{snip.photo.alt}"`, `alt="{snip.other.alt}"`, 1) + `<h2>{snip.name}</h2></article>`,
		`<article>` + strings.Replace(photo, `src="{snip.photo.src}"`, `src="prefix{snip.photo.src}"`, 1) + `<h2>{snip.name}</h2></article>`,
		`<article>` + strings.Replace(photo, `alt="{snip.photo.alt}"`, `onerror="{snip.photo.alt}"`, 1) + `<h2>{snip.name}</h2></article>`,
	} {
		if _, e := definition(body); e == nil {
			t.Fatal("accepted invalid photo context", body)
		}
	}
}

func TestQAMultisnipEmptyPhotoSeed(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<article multisnip="team"><img snip="photo" alt="" width="160" height="160" class="rounded-full object-cover mb-md"><h2 snip="name">Joe</h2></article>`)
	mustLoad(t, root)
	if err := ExtractSnips(root); err != nil {
		t.Fatal(err)
	}
	mustLoad(t, root)
	put(t, root, "index.html", `<img snip="photo" src="/image.png">`)
	if _, err := Load(root); err == nil {
		t.Fatal("standalone image snip accepted")
	}
}
