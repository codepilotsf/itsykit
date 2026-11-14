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
