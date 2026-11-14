package site

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"itsy/internal/snips"
)

func seoConfig(t *testing.T, root, settings string) {
	t.Helper()
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"seo":`+settings+`}`)
}

func sitemapLocations(t *testing.T, body []byte) []string {
	t.Helper()
	var doc struct {
		XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
		URLs    []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(body, &doc); err != nil {
		t.Fatal(err, string(body))
	}
	var locations []string
	for _, entry := range doc.URLs {
		locations = append(locations, entry.Loc)
	}
	return locations
}

func TestSEOAutomaticDiscovery(t *testing.T) {
	root := fixture(t)
	seoConfig(t, root, `{"url":"https://Example.test/","exclude":["/receipt"]}`)
	put(t, root, "main.layout.html", "<!doctype html><html><head><title>{page.title}</title></head><body>{slot}</body></html>")
	put(t, root, "receipt.html", "Receipt")
	put(t, root, "research & notes.html", "Research")
	put(t, root, "_private.html", "PRIVATE")
	put(t, root, "admin/index.html", "PRIVATE")
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	w := request(s, "GET", "/sitemap.xml?ignored=1")
	want := []string{"https://example.test/", "https://example.test/about", "https://example.test/research%20&%20notes"}
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/xml; charset=utf-8" || !reflect.DeepEqual(sitemapLocations(t, w.Body.Bytes()), want) {
		t.Fatalf("sitemap: %d %s %s", w.Code, w.Header(), w.Body.String())
	}
	if strings.Contains(w.Body.String(), "lastmod") || strings.Contains(w.Body.String(), "PRIVATE") || w.Header().Get("Set-Cookie") != "" {
		t.Fatal("sitemap contains invented dates or private state")
	}
	for _, route := range []string{"/sitemap.xml", "/robots.txt", "/about", "/receipt"} {
		get := request(s, "GET", route)
		head := request(s, "HEAD", route)
		if head.Code != get.Code || head.Body.Len() != 0 || !reflect.DeepEqual(head.Header(), get.Header()) {
			t.Fatalf("HEAD differs for %s: %v / %v", route, head, get)
		}
	}
	robots := request(s, "GET", "/robots.txt")
	if robots.Body.String() != "User-agent: *\nDisallow:\n\nSitemap: https://example.test/sitemap.xml\n" || robots.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatal(robots)
	}
	for _, route := range []string{"/sitemap.xml", "/robots.txt"} {
		if w := request(s, "POST", route); w.Code != 405 {
			t.Fatal("discovery endpoint accepted POST", route, w.Code)
		}
	}
	r := httptest.NewRequest("GET", "https://staging.test/about?campaign=one", nil)
	r.Header.Set("X-Forwarded-Host", "attacker.test")
	r.Header.Set("Forwarded", "host=attacker.test;proto=http")
	page := httptest.NewRecorder()
	s.ServeHTTP(page, r)
	if !strings.Contains(page.Body.String(), `<link rel="canonical" href="https://example.test/about">`) || strings.Contains(page.Body.String(), "attacker.test") || strings.Contains(page.Body.String(), "campaign=") {
		t.Fatal("canonical was derived from the request", page.Body.String())
	}
	excluded := request(s, "GET", "/receipt")
	if excluded.Code != 200 || excluded.Header().Get("X-Robots-Tag") != "noindex" || strings.Contains(excluded.Body.String(), `rel="canonical"`) {
		t.Fatal("page exclusion did not stay crawlable and noindex", excluded)
	}
	if w := request(s, "GET", "/missing"); w.Code != 404 || strings.Contains(w.Body.String(), `rel="canonical"`) || w.Header().Get("Link") != "" {
		t.Fatal("error acquired a canonical", w)
	}
	dev := request(NewServer(p), "GET", "/sitemap.xml")
	if !bytes.Equal(dev.Body.Bytes(), w.Body.Bytes()) {
		t.Fatal("dev injected a reload script into XML")
	}
}

func TestSEOPreservesCustomMetadata(t *testing.T) {
	root := fixture(t)
	seoConfig(t, root, `{"url":"https://example.test"}`)
	put(t, root, "index.html", `<html><head><link rel="alternate CANONICAL" href="https://elsewhere.test/"></head><body>Custom</body></html>`)
	put(t, root, "about/index.html", `<html><head><meta name="robots" content="noindex, follow"></head><body>Hidden from search</body></html>`)
	put(t, root, "self.html", `<html><head><link rel="canonical" href="https://example.test/self"><link rel="stylesheet" href="/style.css"></head></html>`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	if got := sitemapLocations(t, request(s, "GET", "/sitemap.xml").Body.Bytes()); !reflect.DeepEqual(got, []string{"https://example.test/self"}) {
		t.Fatal("sitemap conflicts with custom canonical or noindex", got)
	}
	for _, route := range []string{"/", "/about", "/self"} {
		before, err := p.render(p.Pages[route], route)
		if err != nil {
			t.Fatal(err)
		}
		after := request(s, "GET", route)
		if !bytes.Equal(before, after.Body.Bytes()) || after.Header().Get("Link") != "" {
			t.Fatal("custom metadata changed", route, after)
		}
	}
}

func TestSEOCustomDiscoveryAndOptIn(t *testing.T) {
	for _, large := range []bool{false, true} {
		t.Run(fmt.Sprint(large), func(t *testing.T) {
			root := fixture(t)
			seoConfig(t, root, `{"url":"https://example.test"}`)
			robots, sitemap := "User-agent: *\nDisallow: /special\n", "CUSTOM SITEMAP"
			if large {
				robots += strings.Repeat("# custom\n", 10000)
				sitemap += strings.Repeat(" ", memoryAssetLimit)
			}
			put(t, root, "robots.txt", robots)
			put(t, root, "sitemap.xml", sitemap)
			p := mustLoad(t, root)
			if p.seoRobots != nil || p.seoSitemap != nil {
				t.Fatal("custom document acquired a generated replacement")
			}
			s := NewProductionServer(p, nil)
			for route, want := range map[string]string{"/robots.txt": robots, "/sitemap.xml": sitemap} {
				get := request(s, "GET", route)
				head := request(s, "HEAD", route)
				if get.Code != 200 || get.Body.String() != want || head.Code != 200 || head.Body.Len() != 0 {
					t.Fatal("custom document changed", route)
				}
			}
			if large && (p.assetFiles["/robots.txt"].file == nil || p.assetFiles["/sitemap.xml"].file == nil) {
				t.Fatal("large custom discovery documents were buffered")
			}
		})
	}
	root := fixture(t)
	s := NewProductionServer(mustLoad(t, root), nil)
	for _, route := range []string{"/sitemap.xml", "/robots.txt"} {
		if got := request(s, "GET", route).Code; got != 404 {
			t.Fatal("SEO enabled without a configured origin", route, got)
		}
	}
	if w := request(s, "GET", "/"); w.Header().Get("Link") != "" {
		t.Fatal("unconfigured site received a canonical")
	}
}

func TestSEOBlogSnapshotsAndAppBoundaries(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{},"seo":{"url":"https://example.test"},"apps":{"blog":{"mount":"/journal"},"contact":{"mount":"/contact","email":{"to":"admin@example.test"}},"demo":{"mount":"/demo"}}}`)
	put(t, root, "journal/_index.html", "Blog")
	put(t, root, "journal/_post.html", "{blog.data.post.title}")
	put(t, root, "contact/index.html", `<html><head><title>Contact</title></head><body><form>{csrf}</form></body></html>`)
	put(t, root, "contact/thank-you.html", "Receipt")
	put(t, root, "demo/index.html", "Custom app")
	put(t, root, "demo/account.html", "Session page")
	put(t, root, "itsy/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	writeExecutable(t, root, "itsy/apps/demo/demo")
	content := snips.NewStore()
	content.Blog = []snips.BlogPost{
		{ID: "abcdefghij", Slug: "published", Title: "Public", Status: "published", PublishedAt: "2026-10-01T12:00:00Z"},
		{ID: "klmnopqrst", Slug: "private-draft", Title: "Private", Status: "draft"},
	}
	p, err := LoadWithContent(root, content)
	if err != nil {
		t.Fatal(err)
	}
	s := NewProductionServer(p, nil)
	original := append([]byte(nil), p.seoSitemap...)
	want := []string{"https://example.test/", "https://example.test/about", "https://example.test/contact", "https://example.test/journal", "https://example.test/journal/published"}
	if got := sitemapLocations(t, original); !reflect.DeepEqual(got, want) {
		t.Fatal("non-public app route entered sitemap", got)
	}
	contact := request(s, "GET", "https://example.test/contact")
	if contact.Code != 200 || !strings.Contains(contact.Body.String(), `href="https://example.test/contact"`) {
		t.Fatal("Contact did not receive its canonical", contact)
	}
	if w := request(s, "HEAD", "https://example.test/contact/thank-you"); w.Code != 200 || w.Header().Get("X-Robots-Tag") != "noindex" {
		t.Fatal("Contact receipt was indexable", w)
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	next, err := snips.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	next.Blog[0].Status = "draft"
	next.Blog[1].Status, next.Blog[1].PublishedAt = "published", "2026-10-02T12:00:00Z"
	updated, err := LoadWithContent(root, next)
	if err != nil {
		t.Fatal(err)
	}
	s.Publish(updated)
	current := request(s, "GET", "/sitemap.xml").Body.String()
	if strings.Contains(current, "/journal/published") || !strings.Contains(current, "/journal/private-draft") || !bytes.Equal(original, p.seoSitemap) {
		t.Fatal("publication did not replace sitemap atomically", current)
	}
	// A source reload replaces both origin and documents without mutating the old
	// snapshot or retaining an endpoint that is no longer configured.
	put(t, root, "itsy.config.json", `{"app":{}}`)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", "/sitemap.xml"); w.Code != 404 {
		t.Fatal("disabled SEO retained old sitemap", w)
	}
}

func TestSEOConfigurationRejectsInvalidSettings(t *testing.T) {
	for _, cfg := range []string{
		`null`, `[]`, `{}`, `{"url":"http://example.test"}`, `{"url":"https://user:password@example.test"}`,
		`{"url":"https://example.test/subdir"}`, `{"url":"https://example.test?"}`, `{"url":"https://example.test/#fragment"}`,
		`{"url":"https://bad_host.test"}`, `{"url":"https://example.test:65536"}`, `{"url":"https://example.test:"}`,
		`{"url":"https://example.test","unknown":true}`, `{"url":"https://example.test","exclude":["/about/"]}`,
		`{"url":"https://example.test","exclude":["/private/*"]}`, `{"url":"https://example.test","exclude":["/about?preview=1"]}`,
		`{"url":"https://example.test","exclude":["/about","/about"]}`,
	} {
		t.Run(cfg, func(t *testing.T) {
			root := fixture(t)
			seoConfig(t, root, cfg)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "seo") || strings.Contains(err.Error(), "password") {
				t.Fatal("invalid configuration accepted or exposed", err)
			}
		})
	}
}

func TestSEOCanonicalInsertionPreservesSource(t *testing.T) {
	p := mustLoad(t, fixture(t))
	p.seo.URL = "https://example.test"
	for _, body := range []string{
		`<!doctype html><HTML><HEAD><script>const x = "</head><link rel='canonical'>";</script><!-- <link rel="canonical"> --></HEAD><body>Body</body></HTML>`,
		`<html><head><template><link rel="canonical" href="https://inert.test"></template></head><body>Body</body></html>`,
	} {
		w := httptest.NewRecorder()
		got := string(p.canonicalResponse(w, "/", []byte(body)))
		tag := `<link rel="canonical" href="https://example.test/">` + "\n"
		if !strings.Contains(got, tag) || strings.Replace(got, tag, "", 1) != body {
			t.Fatal("canonical insertion changed source or treated inert markup as active", got)
		}
	}
	w := httptest.NewRecorder()
	body := []byte("<main>A fragment</main>")
	if got := p.canonicalResponse(w, "/", body); !bytes.Equal(got, body) || w.Header().Get("Link") != `<https://example.test/>; rel="canonical"` {
		t.Fatal("fragment did not get an HTTP canonical", w.Header())
	}
}

func TestSEOGeneratedDocumentsNeverWriteSiteFiles(t *testing.T) {
	root := fixture(t)
	seoConfig(t, root, `{"url":"https://example.test"}`)
	before, err := os.ReadFile(filepath.Join(root, "itsy.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	_ = mustLoad(t, root)
	for _, name := range []string{"robots.txt", "sitemap.xml"} {
		if _, err := os.Stat(filepath.Join(root, "site", name)); !os.IsNotExist(err) {
			t.Fatal("generation wrote into the project", name, err)
		}
	}
	after, err := os.ReadFile(filepath.Join(root, "itsy.config.json"))
	if err != nil || !bytes.Equal(before, after) || !json.Valid(after) {
		t.Fatal("generation rewrote configuration", err)
	}
}

func TestSEORedirectBoundaries(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{},"seo":{"url":"https://example.test"},"redirects":{"/old-about.html":"/about"}}`)
	s := NewProductionServer(mustLoad(t, root), nil)
	w := request(s, "GET", "/old-about.html")
	if w.Code != 301 || w.Header().Get("Location") != "/about" || w.Header().Get("Link") != "" || strings.Contains(w.Body.String(), `rel="canonical"`) {
		t.Fatal("SEO changed a migration redirect", w)
	}
	locations := sitemapLocations(t, request(s, "GET", "/sitemap.xml").Body.Bytes())
	if !reflect.DeepEqual(locations, []string{"https://example.test/", "https://example.test/about"}) {
		t.Fatal("redirect source entered sitemap", locations)
	}
	for _, source := range []string{"/sitemap.xml", "/robots.txt"} {
		put(t, root, "itsy.config.json", `{"app":{},"seo":{"url":"https://example.test"},"redirects":{"`+source+`":"/about"}}`)
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "automatic SEO endpoint") {
			t.Fatal("redirect shadowed generated discovery", source, err)
		}
	}
}
