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
	put(t, root, "_layout.html", `<html><body>Default {slot}</body></html>`)
	put(t, root, "shell/_brand.html", `<header>Custom header</header>`)
	put(t, root, "shell/_admin.html", `<html><body>{render brand}{slot}</body></html>`)
	put(t, root, "custom.html", "<!-- flea\nlayout: /shell/_admin.html\n-->\n<h1>{site.name}</h1>")
	s := NewServer(mustLoad(t, root))
	got := request(s, "GET", "/custom").Body.String()
	if !strings.Contains(got, "Custom header") || strings.Contains(got, "Default") || strings.Contains(got, "<!-- flea") {
		t.Fatal(got)
	}
	if request(s, "GET", "/shell/_admin.html").Code != 404 {
		t.Fatal("layout publicly accessible")
	}
	put(t, root, "shell/_admin.html", `<html><body>Changed {slot}</body></html>`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(request(s, "GET", "/custom").Body.String(), "Changed") {
		t.Fatal("layout edit not loaded")
	}
}

func TestCustomLayoutExtractionPreservesSettings(t *testing.T) {
	root := fixture(t)
	put(t, root, "_custom.html", `<html><body>{slot}</body></html>`)
	const settings = "<!-- flea\nlayout: /_custom.html\n-->\n"
	put(t, root, "custom.html", settings+`<h1 flea="Heading">Editable</h1>`)
	if _, err := LoadEditing(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "custom.html"))
	if err != nil || !strings.HasPrefix(string(data), settings) || !strings.Contains(string(data), "{flea ") {
		t.Fatal(string(data), err)
	}
}

func TestInvalidPageSettings(t *testing.T) {
	for _, settings := range []string{
		"layout: /missing.html", "layout: ../../outside.html", "layout: /../_outside.html",
		"layout: /_custom.html\nlayout: /_custom.html", "unknown: value", "layout: /public.html",
		"layout: /_custom.html\n-->\n<!-- flea\nlayout: /_custom.html",
	} {
		t.Run(settings, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "_custom.html", `<html>{slot}</html>`)
			put(t, root, "custom.html", "<!-- flea\n"+settings+"\n-->\nBody")
			if _, err := Load(root); err == nil {
				t.Fatal("accepted", settings)
			}
		})
	}
}

func TestLayoutsCannotNestOrBeIncluded(t *testing.T) {
	root := fixture(t)
	put(t, root, "custom.html", "<!-- flea\nlayout: /_custom.html\n-->Body")
	put(t, root, "_custom.html", "<!-- flea\nlayout: /_layout.html\n-->{slot}")
	put(t, root, "_layout.html", `<html>{slot}</html>`)
	if _, err := Load(root); err == nil {
		t.Fatal("accepted nested layout")
	}
	put(t, root, "_custom.html", `<html>{slot}</html>`)
	put(t, root, "index.html", `{render custom}`)
	if _, err := Load(root); err == nil {
		t.Fatal("layout included as partial")
	}
}
