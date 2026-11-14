package site

import (
	"context"
	"html"
	"itsy/internal/snips"
	"itsy/internal/viewer"
	"itsy/templates"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestStarterEditableCopyAndNavigation(t *testing.T) {
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
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	// Contact has six editable headings/paragraphs; form labels stay in HTML.
	for route, minimum := range map[string]int{"/": 9, "/team": 2, "/contact": 6, "/contact/thank-you": 2} {
		if page := p.Content.Pages[route]; page == nil || len(page.Snips) < minimum {
			t.Errorf("%s lacks editable starter copy", route)
		}
	}
	if len(p.Content.Shared.Snips) == 0 {
		t.Error("missing shared footer snip")
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
	r := httptest.NewRequest("GET", "https://example.test/contact", nil)
	r = r.WithContext(viewer.With(r.Context(), viewer.Admin{Email: "owner@example.test", CSRF: "test"}))
	w := httptest.NewRecorder()
	s.ServeOwner(w, r, func(b []byte, _ *snips.Store, _ string) ([]byte, bool) { return b, true })
	if w.Code != 200 || strings.Contains(w.Body.String(), `href="/admin"`) {
		t.Fatal("signed-in navigation must not offer Admin", w.Code)
	}
	if _, err = Load(root); err != nil {
		t.Fatal(err)
	}
}
