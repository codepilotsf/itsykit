package site

import (
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/appapi"
)

// A site may enable the built-in Blog and a custom managed app at the same
// time while ordinary developer templates carry literal brace groups and raw
// JSON. The three features share the loader, compiler and app registry, so this
// guards their coexistence after the custom-app reconciliation.
func TestBlogCustomAppAndLiteralBracesCoexist(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"blog":{"mount":"/blog"},"demo":{"mount":"/demo"}}}`)
	put(t, root, "site/blog/_index.html", `<main><h1>Blog</h1>{#each blog.data.posts as post}<a href="{post.path}">{post.title}</a>{/each}</main>`)
	put(t, root, "site/blog/_post.html", `<main><h1>{blog.data.post.title}</h1>{blog.data.post.body}</main>`)
	put(t, root, "site/demo/index.html", `<form method="post" action="/demo">{csrf}</form><p data-app='{"name":"demo","value":"{app.site_name}"}'></p>`)
	put(t, root, "site/index.html", `<p>Use { and } to group items.</p><p>{1, 2, {3, 4}}</p>`)
	put(t, root, "itsy/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	writeExecutable(t, root, "itsy/apps/demo/demo")

	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.Apps["blog"]; !ok {
		t.Fatal("blog not registered")
	}
	if _, ok := p.appPackages["demo"]; !ok {
		t.Fatal("custom app package not registered")
	}
	w := request(NewServer(p), "GET", "/")
	body := w.Body.String()
	if w.Code != 200 {
		t.Fatalf("index status %d: %s", w.Code, body)
	}
	for _, want := range []string{"Use { and } to group items.", "{1, 2, {3, 4}}"} {
		if !strings.Contains(body, want) {
			t.Fatalf("literal %q not preserved: %s", want, body)
		}
	}
}

// The built-in template globals are reserved, so a custom package cannot shadow
// a core template value. "blog" stays a valid app name because it selects the
// bundled Blog app, whose own global is therefore never shadowed.
func TestAppNamesReserveCurrentNamespaces(t *testing.T) {
	for _, name := range []string{"app", "page", "snip", "data", "media", "videos", "admin", "csrf"} {
		if appapi.ValidName(name) {
			t.Fatalf("template namespace %q must be reserved", name)
		}
	}
	if !appapi.ValidName("demo") || !appapi.ValidName("blog") {
		t.Fatal("ordinary app name rejected")
	}

	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"media":{"mount":"/m"}}}`)
	put(t, root, "itsy/apps/media/app.json", manifest("media", "1.0.0", 2, "media"))
	writeExecutable(t, root, "itsy/apps/media/media")
	put(t, root, "site/m/index.html", "Form")
	if _, err := Load(root); err == nil {
		t.Fatal("app shadowing a reserved namespace was accepted")
	}
}

// A custom app mount cannot overlap the built-in Blog subtree.
func TestCustomAppMountCannotOverlapBlog(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"blog":{"mount":"/blog"},"demo":{"mount":"/blog/extra"}}}`)
	put(t, root, "site/blog/_index.html", `<main>Blog</main>`)
	put(t, root, "site/blog/_post.html", `<main>{blog.data.post.title}</main>`)
	put(t, root, "site/blog/extra/index.html", "Form")
	put(t, root, "itsy/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	writeExecutable(t, root, "itsy/apps/demo/demo")
	if _, err := Load(root); err == nil {
		t.Fatal("mount overlapping the Blog subtree was accepted")
	}
}

// The runtime root defaults inside the development site and the resolved app
// storage stays under it, outside any release path.
func TestCustomAppRuntimeRootDefaultsUnderSite(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`)
	put(t, root, "itsy/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	writeExecutable(t, root, "itsy/apps/demo/demo")
	put(t, root, "site/demo/index.html", "Form")

	p := mustLoad(t, root)
	want := filepath.Join(root, "itsy", "runtime")
	if p.RuntimeRoot != want {
		t.Fatalf("runtime root %q, want %q", p.RuntimeRoot, want)
	}
	if got := p.appPackages["demo"].Storage; got != filepath.Join(want, "apps", "demo") {
		t.Fatalf("storage %q, want %q", got, filepath.Join(want, "apps", "demo"))
	}
}

// Reloading after an unrelated template edit keeps the validated custom package
// registered and stable, exercising the reload path touched by the port.
func TestCustomAppRegistrationSurvivesReload(t *testing.T) {
	root := fixture(t)
	put(t, root, "itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"demo":{"mount":"/demo"}}}`)
	put(t, root, "itsy/apps/demo/app.json", manifest("demo", "1.0.0", 2, "demo"))
	writeExecutable(t, root, "itsy/apps/demo/demo")
	put(t, root, "site/demo/index.html", "Form")
	put(t, root, "site/index.html", "<p>first</p>")

	s := NewServer(mustLoad(t, root))
	put(t, root, "site/index.html", "<p>second</p>")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	s.mu.RLock()
	p := s.project
	s.mu.RUnlock()
	if _, ok := p.appPackages["demo"]; !ok {
		t.Fatal("custom package lost on reload")
	}
}
