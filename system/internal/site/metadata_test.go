package site

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPageMetadataHTTPAndReloadWithoutWrites(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<html><head><title>{page.title}</title><meta name="description" content="{page.description}"></head><body>{slot}</body></html>`)
	put(t, root, "private/custom.layout.html", `<html><head><title>{page.title}</title><meta name="description" content="{page.description}"></head><body>Custom {slot}</body></html>`)
	put(t, root, "itsy/snips.json", "{}\n")
	before, _ := os.Stat(filepath.Join(root, "itsy/snips.json"))
	s := NewServer(mustLoad(t, root))
	server := httptest.NewServer(s)
	defer server.Close()
	get := func(route string) string {
		t.Helper()
		resp, err := server.Client().Get(server.URL + route)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode, string(b))
		}
		return string(b)
	}
	for _, tc := range []struct{ route, file, comment, title, desc string }{
		{"/", "index.html", "", "Home", ""},
		{"/visit-us", "visit-us.html", "<!-- itsy\ntitle:\ndescription:\n-->\n", "Visit us", ""},
		{"/nested/open-hours", "nested/open-hours/index.html", "<!-- itsy\ndescription: Open daily: 9 to 5 {app.site_name}\n-->\n", "Open hours", "Open daily: 9 to 5 {app.site_name}"},
		{"/details", "details.html", "<!-- itsy\ntitle: Visit <b> & \"bookshop\" {app.site_name}\ndescription: Quote \" onmouseover=\"bad <script>: {render missing}\nlayout: /private/custom.layout.html\n-->\n", "Visit &lt;b&gt; &amp; &#34;bookshop&#34; {app.site_name}", "Quote &#34; onmouseover=&#34;bad &lt;script&gt;: {render missing}"},
		{"/legacy", "legacy.html", "\ufeff \r\n<!-- flea\r\ntitle: Legacy: {braces}\r\ndescription: Existing spelling\r\n-->\n", "Legacy: {braces}", "Existing spelling"},
	} {
		put(t, root, tc.file, tc.comment+"<p>Ordinary body</p>")
		if err := s.Reload(root); err != nil {
			t.Fatal(err)
		}
		body := get(tc.route)
		if !strings.Contains(body, "<title>"+tc.title+"</title>") || !strings.Contains(body, `content="`+tc.desc+`"`) {
			t.Fatalf("%s: %s", tc.route, body)
		}
		if tc.route == "/details" && !strings.Contains(body, "Custom") {
			t.Fatal("custom layout not used")
		}
	}
	put(t, root, "details.html", "<!-- itsy\ntitle: Reloaded\n-->\n<p>New</p>")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if body := get("/details"); !strings.Contains(body, "<title>Reloaded</title>") || !strings.Contains(body, `content=""`) {
		t.Fatal(body)
	}
	// Dev loading must also leave HTML and content bytes untouched.
	htmlBefore, _ := os.ReadFile(filepath.Join(root, "site/details.html"))
	if _, err := LoadDev(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "itsy/snips.json"))
	info, _ := os.Stat(filepath.Join(root, "itsy/snips.json"))
	htmlAfter, _ := os.ReadFile(filepath.Join(root, "site/details.html"))
	if string(after) != "{}\n" || !before.ModTime().Equal(info.ModTime()) || string(htmlBefore) != string(htmlAfter) {
		t.Fatal("load rewrote source or content")
	}
	resp, err := server.Client().Get(server.URL + "/private/custom.layout.html")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("private layout exposed")
	}
}

func TestRemovedPageCMSIsRejectedWithoutWrites(t *testing.T) {
	for _, source := range []string{`<p snip>Old</p>`, `<img snip src="/a.png">`, `<video snip src="/a.mp4"></video>`, `<div multisnip="team">Old</div>`, `{snip Abcdef1234}`, `{snip notice}`, `{snip.team.title}`, `{#multisnip team}{/multisnip}`} {
		t.Run(source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "index.html", source)
			if _, err := LoadDev(context.Background(), root); err == nil {
				t.Fatal("removed CMS accepted")
			}
			b, _ := os.ReadFile(filepath.Join(root, "site/index.html"))
			if string(b) != source {
				t.Fatal("rewritten")
			}
			if _, err := os.Stat(filepath.Join(root, "itsy/snips.json")); !os.IsNotExist(err) {
				t.Fatal("content written", err)
			}
		})
	}
}
