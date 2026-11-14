package site

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomLayout(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<html><body>Default {slot}</body></html>`)
	put(t, root, "shell/_brand.html", `<header>Custom header</header>`)
	put(t, root, "shell/admin.layout.html", `<html><body>{render brand}{slot}</body></html>`)
	put(t, root, "custom.html", "<!-- flea\nlayout: /shell/admin.layout.html\n-->\n<h1>{app.site_name}</h1>")
	s := NewServer(mustLoad(t, root))
	got := request(s, "GET", "/custom").Body.String()
	if !strings.Contains(got, "Custom header") || strings.Contains(got, "Default") || strings.Contains(got, "<!-- flea") {
		t.Fatal(got)
	}
	if request(s, "GET", "/shell/admin.layout.html").Code != 404 {
		t.Fatal("layout publicly accessible")
	}
	put(t, root, "shell/admin.layout.html", `<html><body>Changed {slot}</body></html>`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(request(s, "GET", "/custom").Body.String(), "Changed") {
		t.Fatal("layout edit not loaded")
	}
}

func TestCustomLayoutExtractionPreservesSettings(t *testing.T) {
	root := fixture(t)
	put(t, root, "custom.layout.html", `<html><body>{slot}</body></html>`)
	const settings = "<!-- flea\nlayout: /custom.layout.html\n-->\n"
	put(t, root, "custom.html", settings+`<h1 snip="Heading">Editable</h1>`)
	if _, err := LoadEditing(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "site", "custom.html"))
	if err != nil || !strings.HasPrefix(string(data), settings) || !strings.Contains(string(data), "{snip ") {
		t.Fatal(string(data), err)
	}
}

func TestInvalidPageSettings(t *testing.T) {
	for _, settings := range []string{
		"layout: /missing.html", "layout: ../../outside.html", "layout: /../_outside.html",
		"layout: /custom.layout.html\nlayout: /custom.layout.html", "unknown: value", "layout: /public.html",
		"layout: /custom.layout.html\n-->\n<!-- flea\nlayout: /custom.layout.html",
	} {
		t.Run(settings, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "custom.layout.html", `<html>{slot}</html>`)
			put(t, root, "custom.html", "<!-- flea\n"+settings+"\n-->\nBody")
			if _, err := Load(root); err == nil {
				t.Fatal("accepted", settings)
			}
		})
	}
}

func TestLayoutsCannotNestOrBeIncluded(t *testing.T) {
	root := fixture(t)
	put(t, root, "custom.html", "<!-- flea\nlayout: /custom.layout.html\n-->Body")
	put(t, root, "custom.layout.html", "<!-- flea\nlayout: /main.layout.html\n-->{slot}")
	put(t, root, "main.layout.html", `<html>{slot}</html>`)
	if _, err := Load(root); err == nil {
		t.Fatal("accepted nested layout")
	}
	put(t, root, "custom.layout.html", `<html>{slot}</html>`)
	put(t, root, "index.html", `{render custom}`)
	if _, err := Load(root); err == nil {
		t.Fatal("layout included as partial")
	}
}

func TestPageMetadataDefaults(t *testing.T) {
	root := fixture(t)
	put(t, root, "custom.layout.html", `<html><head><title>{page.title}</title><meta name="description" content="{page.description}"></head><body>Custom {slot}</body></html>`)
	const settings = "<!-- itsy\nlayout: /custom.layout.html\ntitle: Books: {app.site_name} & <friends>\ndescription: A shop: \"near you\" & <script>bad()</script>\n-->\n"
	put(t, root, "about/index.html", settings+`<h1 snip>About</h1>`)
	put(t, root, "itsy/snips.json", `{"pages":{"/about":{"title":"","description":""}}}`)
	before, err := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(mustLoad(t, root))
	got := request(s, "GET", "/about").Body.String()
	for _, want := range []string{`<title>Books: {app.site_name} &amp; &lt;friends&gt;</title>`, `content="A shop: &#34;near you&#34; &amp; &lt;script&gt;bad()&lt;/script&gt;"`, `<h1 snip>About</h1>`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "<!-- itsy") {
		t.Fatal("page settings leaked into output", got)
	}
	after, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	if string(before) != string(after) {
		t.Fatal("read-only load persisted metadata defaults")
	}
	// Each nonempty saved field overrides only its corresponding default.
	put(t, root, "itsy/snips.json", `{"pages":{"/about":{"title":"Saved title","description":""}}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	got = request(s, "GET", "/about").Body.String()
	if !strings.Contains(got, "<title>Saved title</title>") || !strings.Contains(got, `content="A shop:`) {
		t.Fatal(got)
	}
	put(t, root, "itsy/snips.json", `{"pages":{"/about":{"title":"","description":"Saved description"}}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	got = request(s, "GET", "/about").Body.String()
	if !strings.Contains(got, "<title>Books:") || !strings.Contains(got, `content="Saved description"`) {
		t.Fatal(got)
	}
	// Extraction preserves the comment, without copying defaults into saved metadata.
	if _, err := LoadEditing(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	page, _ := os.ReadFile(filepath.Join(root, "site/about/index.html"))
	if !strings.HasPrefix(string(page), settings) || !strings.Contains(string(page), "{snip ") {
		t.Fatal(string(page))
	}
	loaded := mustLoad(t, root)
	if p := loaded.Content.Pages["/about"]; p.Title != "" || p.Description != "Saved description" {
		t.Fatal("extraction changed saved metadata", p)
	}
	put(t, root, "about/index.html", "<!-- itsy\nlayout: /custom.layout.html\ntitle: Changed in HTML\n-->\nAbout")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if got := request(s, "GET", "/about").Body.String(); !strings.Contains(got, "Changed in HTML") {
		t.Fatal("metadata edit not reloaded", got)
	}
}

func TestInvalidPageMetadata(t *testing.T) {
	for _, tc := range []struct{ file, settings, message string }{
		{"index.html", "title: First\ntitle: Second", "duplicate title"},
		{"index.html", "description:\ndescription: Second", "duplicate description"},
		{"index.html", "title: Missing close", "unclosed itsy"},
		{"_partial.html", "title: Partial", "partials cannot have page settings"},
		{"main.layout.html", "description: Layout", "page titles and descriptions belong in pages"},
	} {
		t.Run(tc.file+tc.settings, func(t *testing.T) {
			root := fixture(t)
			source := "<!-- itsy\n" + tc.settings
			if tc.message != "unclosed itsy" {
				source += "\n-->\nBody"
			}
			put(t, root, tc.file, source)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("want %q, got %v", tc.message, err)
			}
		})
	}
}
