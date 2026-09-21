package site

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" stroke="currentColor"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>`

type iconTransport func(*http.Request) (*http.Response, error)

func (f iconTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestIconsResolveOnlyNeededAndWorkOffline(t *testing.T) {
	root := fixture(t)
	put(t, root, "html/index.html", `{icon arrow-right}{icon arrow-right size="18"}{render card}`)
	put(t, root, "html/_card.html", `{icon search aria-label="Search"}`)
	if p, err := Load(root); err != nil || len(p.Warnings) != 2 {
		t.Fatalf("read-only load must warn for two missing icons: %v, %+v", err, p)
	}
	if _, err := os.Stat(filepath.Join(root, "_/icons")); !os.IsNotExist(err) {
		t.Fatal("Load wrote cache")
	}
	requested := map[string]int{}
	client := &http.Client{Transport: iconTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" || r.URL.Host != "unpkg.com" || !strings.HasPrefix(r.URL.Path, "/lucide-static@"+LucideVersion+"/") {
			t.Fatal(r.URL)
		}
		requested[r.URL.Path]++
		body := testIconSVG
		if strings.HasSuffix(r.URL.Path, "LICENSE") {
			body = "License fixture"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	p, err := loadWithIcons(context.Background(), root, func(ctx context.Context, name string) error { return downloadIcon(ctx, root, name, client) })
	if err != nil {
		t.Fatal(err)
	}
	if len(requested) != 3 {
		t.Fatalf("expected two SVGs and license, got %v", requested)
	}
	for name, count := range requested {
		if count != 1 {
			t.Errorf("%s fetched %d times", name, count)
		}
	}
	_, err = loadWithIcons(context.Background(), root, func(context.Context, string) error { t.Fatal("cached load accessed network"); return nil })
	if err != nil {
		t.Fatal(err)
	}
	w := request(NewServer(p), "GET", "/")
	if w.Code != 200 || strings.Count(w.Body.String(), "<svg ") != 3 || !strings.Contains(w.Body.String(), `width="18"`) {
		t.Fatal(w.Body.String())
	}
	if w := request(NewServer(p), "GET", "/"+iconCacheDir+"/arrow-right.svg"); w.Code != 404 {
		t.Fatal("cache publicly accessible")
	}
}

func TestIconAttributesScopeAndEscaping(t *testing.T) {
	root := fixture(t)
	put(t, root, iconCacheDir+"/search.svg", testIconSVG)
	put(t, root, "html/_card.html", `{icon search class="inline {site.name}" aria-label="{member.name}" size="1em" stroke-width='3' data-note="a}b &amp; c"}`)
	put(t, root, "html/index.html", `{#each snip.team.members as member}{render card}{/each}{icon search aria-label="{site.unsafe}"}{icon search}`)
	p := mustLoad(t, root)
	b, err := p.render(p.Pages["/"], "/")
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{`aria-label="Ada"`, `aria-label="Cy"`, `class="lucide lucide-search inline Test"`, `width="1em"`, `height="1em"`, `stroke-width="3"`, `role="img"`, `aria-hidden="true"`, `data-note="a}b &amp; c"`, `&lt;script&gt;`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "<script>alert") {
		t.Fatal("attribute injection")
	}
}

func TestBadIconDirectives(t *testing.T) {
	for _, source := range []string{
		`{icon}`, `{icon ../secret}`, `{icon Search}`, `{icon "search"}`, `{icon search.svg}`,
		`{icon search class=x}`, `{icon search onclick="alert(1)"}`, `{icon search href="https://example.com"}`, `{icon search style="color:red"}`,
		`{icon search class="a" class="b"}`, `{icon search size="12"width="8"}`, `{icon search aria-label="{site.name + 1}"}`, `{icon search class="unfinished}`,
	} {
		t.Run(source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, iconCacheDir+"/search.svg", testIconSVG)
			put(t, root, "html/index.html", source)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "index.html:") {
				t.Fatalf("expected source error, got %v", err)
			}
		})
	}
}

func TestIconsDoNotParseContentOrRawElements(t *testing.T) {
	root := fixture(t)
	put(t, root, "_/snips/content.json", `{"text":"{icon search}<svg onload=alert(1)>"}`)
	put(t, root, "html/index.html", `<script>const icon="{icon search}";</script><style>/* {icon search} */</style>{snip.content.text}`)
	p := mustLoad(t, root)
	b, err := p.render(p.Pages["/"], "/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `{icon search}&lt;svg`) {
		t.Fatal(string(b))
	}
	if _, err := os.Stat(filepath.Join(root, "_/icons")); !os.IsNotExist(err) {
		t.Fatal("data or raw element triggered download")
	}
}

func TestUnsafeIconSVGs(t *testing.T) {
	for _, source := range []string{
		`<svg viewBox="0 0 24 24"><script>alert(1)</script></svg>`,
		`<svg viewBox="0 0 24 24"><use href="https://example.com/x"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="M0 0" onload="alert(1)"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="M0 0" fill="url(https://example.com)"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="[[NW . NW]]"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="{site.name}"/></svg>`,
		`<svg viewBox="0 0 24 24"><foreignObject><div>bad</div></foreignObject></svg>`,
		`<!DOCTYPE svg><svg viewBox="0 0 24 24"><path d="M0 0"/></svg>`,
		testIconSVG + testIconSVG, `<svg/>`, strings.Repeat(" ", maxIconBytes+1),
	} {
		if _, err := sanitizeIcon([]byte(source)); err == nil {
			t.Errorf("accepted unsafe SVG: %.100s", source)
		}
	}
	// Root attributes and comments are discarded, never copied into templates.
	clean, err := sanitizeIcon([]byte(`<svg viewBox="0 0 24 24" onload="alert(1)"><!-- [[NW . NW]] --><circle cx="12" cy="12" r="4"/></svg>`))
	if err != nil || strings.Contains(clean, "onload") || strings.Contains(clean, "[[NW") {
		t.Fatal(clean, err)
	}
}

func TestIconDownloadFailuresAndRecovery(t *testing.T) {
	root := fixture(t)
	put(t, root, "html/index.html", `{icon search}`)
	for _, tc := range []struct {
		code int
		body string
	}{{404, "not found"}, {200, "not SVG"}, {200, strings.Repeat("x", maxIconBytes+1)}} {
		client := &http.Client{Transport: iconTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: tc.code, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})}
		if err := downloadIcon(context.Background(), root, "search", client); err == nil {
			t.Fatal("accepted bad download")
		}
		if _, err := os.Stat(filepath.Join(root, iconCacheDir, "search.svg")); !os.IsNotExist(err) {
			t.Fatal("failed download poisoned cache")
		}
	}
	p, err := loadWithIcons(context.Background(), root, func(context.Context, string) error { return errors.New("offline") })
	if err != nil || len(p.Warnings) != 1 || !strings.Contains(p.Warnings[0], "index.html:1") || !strings.Contains(p.Warnings[0], "offline") {
		t.Fatal(p, err)
	}
	if w := request(NewServer(p), "GET", "/"); w.Code != 200 || strings.Contains(w.Body.String(), "<svg") {
		t.Fatal("missing icon must not break page")
	}
	put(t, root, iconCacheDir+"/search.svg", testIconSVG)
	mustLoad(t, root)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := LoadDev(ctx, root); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestMissingIconsContinueAndRecover(t *testing.T) {
	root := fixture(t)
	put(t, root, "html/index.html", `Before{icon absent}{icon search}{icon absent}After`)
	requested := map[string]int{}
	p, err := loadWithIcons(context.Background(), root, func(_ context.Context, name string) error {
		requested[name]++
		if name == "absent" {
			return errors.New("HTTP 404")
		}
		put(t, root, iconCacheDir+"/"+name+".svg", testIconSVG)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if requested["absent"] != 1 || requested["search"] != 1 || len(p.Warnings) != 1 {
		t.Fatal(requested, p.Warnings)
	}
	s := NewServer(p)
	w := request(s, "GET", "/")
	if w.Code != 200 || strings.Count(w.Body.String(), "<svg ") != 1 || !strings.Contains(w.Body.String(), "After") {
		t.Fatal(w.Body.String())
	}
	put(t, root, iconCacheDir+"/absent.svg", testIconSVG)
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/")
	if w.Code != 200 || strings.Count(w.Body.String(), "<svg ") != 3 || len(s.project.Warnings) != 0 {
		t.Fatal(w.Body.String())
	}
	put(t, root, iconCacheDir+"/absent.svg", `<svg viewBox="0 0 24 24"><script/></svg>`)
	if err := s.Reload(root); err == nil {
		t.Fatal("unsafe cached SVG must remain an error")
	}
	if w := request(s, "GET", "/"); w.Code != 500 {
		t.Fatal(w.Code)
	}
}

func TestUnsafeDownloadedIconRemainsError(t *testing.T) {
	root := fixture(t)
	put(t, root, "html/index.html", `{icon search}`)
	client := &http.Client{Transport: iconTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`<svg viewBox="0 0 24 24"><script/></svg>`))}, nil
	})}
	_, err := loadWithIcons(context.Background(), root, func(ctx context.Context, name string) error { return downloadIcon(ctx, root, name, client) })
	if err == nil || !strings.Contains(err.Error(), "unsupported SVG element") {
		t.Fatal(err)
	}
}

func TestIconCacheRejectsSymlinks(t *testing.T) {
	root := fixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "_/icons")); err != nil {
		t.Fatal(err)
	}
	if cache, err := openIconCache(root); err == nil {
		cache.Close()
		t.Fatal("followed symlink cache")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatal("wrote outside project", err)
	}
}
