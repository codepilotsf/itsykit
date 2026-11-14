package site

import (
	"encoding/json"
	"itsy/internal/snips"
	"strings"
	"testing"
)

func blogFixture(t *testing.T) string {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Blog"},"apps":{"blog":{"mount":"/journal"}}}`)
	put(t, root, "journal/_index.html", `{#each blog.data.posts as post}<a href="{post.url}">{post.title}</a>{/each}`)
	put(t, root, "journal/_post.html", `<h1>{blog.data.post.title}</h1><p>{page.description}</p><p>{page.title}</p>{blog.data.post.body}`)
	return root
}
func TestBlogPublishedViewsAndDraftPrivacy(t *testing.T) {
	root := blogFixture(t)
	s := snips.NewStore()
	s.Page("/journal/older").Title = "Custom SEO"
	s.Blog = []snips.BlogPost{
		{ID: "Abcdef1234", Title: "Older", Slug: "older", Status: "published", PublishedAt: "2026-10-12T10:00:00Z", Body: `<p>Hello {app.site_name}</p><script>bad()</script><img src=x onerror=bad()>`, Excerpt: "An excerpt"},
		{ID: "Bbcdef1234", Title: "Newer <b>", Slug: "newer", Status: "published", PublishedAt: "2026-10-13T10:00:00Z"},
		{ID: "Cbcdef1234", Title: "Secret draft", Slug: "secret", Status: "draft", Body: "Private body"},
	}
	bytes, _ := json.Marshal(s)
	put(t, root, "itsy/snips.json", string(bytes))
	server := NewProductionServer(mustLoad(t, root), nil)
	list := request(server, "GET", "/journal")
	if list.Code != 200 || strings.Contains(list.Body.String(), "Secret") || strings.Index(list.Body.String(), "Newer") > strings.Index(list.Body.String(), "Older") {
		t.Fatal(list.Code, list.Body.String())
	}
	page := request(server, "GET", "/journal/older")
	if page.Code != 200 || !strings.Contains(page.Body.String(), "<p>Hello {app.site_name}</p>") || strings.Contains(page.Body.String(), "bad()") || strings.Contains(page.Body.String(), "<img") || !strings.Contains(page.Body.String(), "An excerpt") || !strings.Contains(page.Body.String(), "Custom SEO") {
		t.Fatal(page.Code, page.Body.String())
	}
	for _, path := range []string{"/journal/secret", "/journal/_post.html", "/flea/snips.json", "/journal/no-post", "/journal/secret.html"} {
		if got := request(server, "GET", path).Code; got != 404 {
			t.Fatal(path, got)
		}
	}
}
func TestBlogRichBodyContextAndMountValidation(t *testing.T) {
	for _, source := range []string{`<p title="{blog.data.post.body}">x</p>`, `<script>{blog.data.post.body}</script>`, `<textarea>{blog.data.post.body}</textarea>`, `<svg>{blog.data.post.body}</svg>`} {
		root := blogFixture(t)
		put(t, root, "journal/_post.html", source)
		if _, err := Load(root); err == nil {
			t.Fatal("unsafe rich context", source)
		}
	}
	for _, path := range []string{"journal/other.html", "journal/page.css"} {
		root := blogFixture(t)
		put(t, root, path, "conflict")
		if _, err := Load(root); err == nil {
			t.Fatal("blog route collision", path)
		}
	}
	root := blogFixture(t)
	put(t, root, "itsy.config.json", `{"app":{},"apps":{"blog":{"mount":"/admin"}}}`)
	if _, err := Load(root); err == nil {
		t.Fatal("reserved mount")
	}
}
