package site

import (
	"encoding/json"
	"strings"
	"testing"

	"flea/internal/snips"
)

func TestConfiguredRedirects(t *testing.T) {
	root := fixture(t)
	put(t, root, ".flea.config.json", `{"app":{},"redirects":{
		"/old.html":"/about", "/old/":"/about#team",
		"/old%20name.php":"/about?source=migration#team", "/style.css":"/css/main.css",
		"/clear":"/about?", "/encoded":"/caf%C3%A9"
	}}`)
	put(t, root, "café.html", "Café")
	p := mustLoad(t, root)
	for _, s := range []*Server{NewServer(p), NewProductionServer(p, nil)} {
		for _, tc := range []struct{ from, to string }{
			{"/old.html", "/about"}, {"/old.html?utm_source=old", "/about?utm_source=old"},
			{"/old/", "/about#team"}, {"/old%20name.php?discard=yes", "/about?source=migration#team"},
			{"/style.css", "/css/main.css"}, {"/clear?discard=yes", "/about?"}, {"/encoded", "/caf%C3%A9"},
		} {
			for _, method := range []string{"GET", "HEAD"} {
				w := request(s, method, tc.from)
				if w.Code != 301 || w.Header().Get("Location") != tc.to {
					t.Errorf("%s %s: %d %q, want 301 %q", method, tc.from, w.Code, w.Header().Get("Location"), tc.to)
				}
				if method == "HEAD" && w.Body.Len() != 0 {
					t.Errorf("HEAD %s returned a body", tc.from)
				}
			}
		}
		// The earlier request's query must never mutate the shared rule.
		if got := request(s, "GET", "/old.html").Header().Get("Location"); got != "/about" {
			t.Fatalf("query leaked between requests: %q", got)
		}
		for _, method := range []string{"POST", "PUT", "DELETE"} {
			if w := request(s, method, "/old.html"); w.Code != 405 || w.Header().Get("Location") != "" {
				t.Errorf("%s redirected a mutation: %d", method, w.Code)
			}
		}
		for _, path := range []string{"/about.html", "/old.html/", "/OLD.html", "/_old.html", "/%252e%252e/old.html", "/../old.html", "//old.html"} {
			if w := request(s, "GET", path); w.Code != 404 {
				t.Errorf("unlisted/unsafe URL %s: %d", path, w.Code)
			}
		}
		if w := request(s, "GET", "/about"); w.Code != 200 {
			t.Errorf("destination: %d", w.Code)
		}
	}
}

func TestRedirectValidation(t *testing.T) {
	for _, raw := range []string{
		`null`, `[]`, `"/about"`, `{"/old":42}`, `{"/old":null}`,
		`{"/old":"/about","/old":"/"}`, `{"/old":"/about","/%6fld":"/"}`,
		`{"/old":"/old"}`, `{"/old":"/next","/next":"/about"}`, `{"/old":"/next","/next":"/old"}`,
		`{"/about":"/"}`, `{"/css/main.css":"/about"}`, `{"/old":"/missing"}`,
		`{"/old":"https://evil.test/about"}`, `{"/old":"//evil.test/about"}`, `{"/old":"/%2fevil.test"}`,
		`{"/old":"/about.html"}`, `{"/old":"/about/"}`, `{"/old":"/admin"}`, `{"/old":"/.env"}`,
		`{"/old":"/about\r\nLocation: https://evil.test"}`, `{"/old":"/%0d%0aevil"}`,
		`{"old":"/about"}`, `{"https://old.test/page":"/about"}`, `{"//old":"/about"}`,
		`{"/old?x=1":"/about"}`, `{"/old#section":"/about"}`, `{"/old?":"/about"}`,
		`{"/a/../old":"/about"}`, `{"/old//":"/about"}`, `{"/%252e%252e/old":"/about"}`,
		`{"/old\\page":"/about"}`, `{"/old%00page":"/about"}`, `{"/old%":"/about"}`,
		`{"/.env":"/about"}`, `{"/%5fprivate/file.html":"/about"}`, `{"/main.layout.html":"/about"}`,
		`{"/admin":"/about"}`, `{"/admin/old.html":"/about"}`, `{"/_flea/health":"/about"}`, `{"/releases/old":"/about"}`,
	} {
		t.Run(raw, func(t *testing.T) {
			root := fixture(t)
			put(t, root, ".flea.config.json", `{"app":{},"redirects":`+raw+`}`)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "redirect") {
				t.Fatalf("expected redirect validation error, got %v", err)
			}
		})
	}
}

func TestRedirectAppBoundariesAndReload(t *testing.T) {
	root := fixture(t)
	put(t, root, "contact/index.html", "Contact")
	put(t, root, "contact/thank-you.html", "Thanks")
	config := `{"app":{},"apps":{"contact":{"mount":"/contact","email":{"to":"admin@example.test"}}},"redirects":`
	put(t, root, ".flea.config.json", config+`{"/old-contact.html":"/contact"}}`)
	s := NewServer(mustLoad(t, root))
	if w := request(s, "GET", "/old-contact.html"); w.Code != 301 || w.Header().Get("Location") != "/contact" {
		t.Fatal("could not redirect to a Contact page", w)
	}
	put(t, root, ".flea.config.json", config+`{"/contact/old.html":"/about"}}`)
	if err := s.Reload(root); err == nil {
		t.Fatal("allowed redirect inside an app's subtree")
	}
	if w := request(s, "GET", "/old-contact.html"); w.Code == 301 {
		t.Fatal("served a stale redirect after failed reload")
	}
	put(t, root, ".flea.config.json", config+`{"/old-contact.html":"/about"}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if got := request(s, "GET", "/old-contact.html").Header().Get("Location"); got != "/about" {
		t.Fatalf("reload kept the old redirect: %q", got)
	}
	put(t, root, ".flea.config.json", `{"app":{},"redirects":{}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", "/old-contact.html"); w.Code != 404 {
		t.Fatal("removed redirect is still served", w)
	}
}

func TestRedirectFormerBlogSlug(t *testing.T) {
	root := blogFixture(t)
	content := snips.NewStore()
	content.Blog = []snips.BlogPost{
		{ID: "Abcdef1234", Title: "Moved post", Slug: "new", Status: "published", PublishedAt: "2026-10-30T10:00:00Z"},
		{ID: "Bbcdef1234", Title: "Private draft", Slug: "draft", Status: "draft"},
	}
	body, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	put(t, root, "flea/snips.json", string(body))
	put(t, root, ".flea.config.json", `{"app":{},"apps":{"blog":{"mount":"/journal"}},"redirects":{"/journal/old":"/journal/new"}}`)
	s := NewProductionServer(mustLoad(t, root), nil)
	if w := request(s, "GET", "/journal/old"); w.Code != 301 || w.Header().Get("Location") != "/journal/new" {
		t.Fatal("former slug did not redirect", w)
	}
	if w := request(s, "GET", "/journal/new"); w.Code != 200 || !strings.Contains(w.Body.String(), "Moved post") {
		t.Fatal("published destination failed", w)
	}
	put(t, root, ".flea.config.json", `{"app":{},"apps":{"blog":{"mount":"/journal"}},"redirects":{"/journal/old":"/journal/draft"}}`)
	if _, err := Load(root); err == nil {
		t.Fatal("accepted an unpublished redirect destination")
	}
}
