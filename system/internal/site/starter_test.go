package site

import (
	"context"
	"html"
	"itsy/templates"
	"os"
	"strings"
	"testing"
)

func TestStarterOrdinaryCopyAndNavigation(t *testing.T) {
	theme, err := templates.Files.ReadFile("site/css/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"Inter", system-ui, sans-serif`, `--bg-color: var(--lighter)`, `--brand-1: #286047`} {
		if !strings.Contains(string(theme), want) {
			t.Errorf("starter theme missing %s", want)
		}
	}
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	p, err := LoadDev(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	s := NewProductionServer(p, nil)
	thankYou := html.UnescapeString(request(s, "GET", "https://example.test/contact/thank-you").Body.String())
	if !strings.Contains(thankYou, "We'll be in touch.") || strings.Contains(thankYou, "Your message has been sent.") {
		t.Fatal("thank-you copy must not duplicate the success flash", thankYou)
	}
	body := html.UnescapeString(request(s, "GET", "https://example.test/contact").Body.String())
	for _, want := range []string{`href="/team"`, `>Team</a>`, `href="/contact"`, `action="/contact"`, `name="_csrf"`, `name="website"`, `name="message"`, "Let's talk."} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s", want)
		}
	}
	if strings.Contains(body, `href="/test"`) {
		t.Error("test page remains in navigation")
	}
	if strings.Count(body, `href="/admin"`) != 2 {
		t.Fatal("visitors need desktop and mobile Admin links")
	}
	if _, err = Load(root); err != nil {
		t.Fatal(err)
	}
}
