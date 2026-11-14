package site

import (
	"context"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func put(t *testing.T, root, name, body string) {
	t.Helper()
	if !strings.HasPrefix(name, "itsy/") && !strings.HasPrefix(name, "mailers/") && !strings.HasPrefix(name, "site/") && !strings.HasPrefix(name, ".") && name != "AGENTS.md" && name != "README.md" {
		name = "site/" + name
	}
	dest := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test","url":"javascript:alert(1)","unsafe":"\"/><script>alert(1)</script>"}}`)
	put(t, root, "itsy/snips/team.json", `{"text":"<b>{app.site_name}</b>","members":[{"name":"Ada","children":[{"name":"Bea"}]},{"name":"Cy","children":[]}],"empty":[],"false":false,"zero":0,"blank":""}`)
	put(t, root, "index.html", "<h1>{app.site_name}</h1>")
	put(t, root, "about/index.html", "About")
	put(t, root, "css/main.css", "body{color:red}")
	put(t, root, "_404.html", "Not here")
	put(t, root, "_500.html", "Try later")
	return root
}
func mustLoad(t *testing.T, root string) *Project {
	t.Helper()
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestConfigurationRequiresJSONFilename(t *testing.T) {
	root := fixture(t)
	current := filepath.Join(root, ".itsy.config.json")
	old := filepath.Join(root, "itsy/flea.config")
	if err := os.Rename(current, old); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "itsy.config.json") {
		t.Fatal("old filename unexpectedly accepted", err)
	}
	if err := os.Rename(old, current); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	for _, route := range []string{"/.itsy.config.json", "/flea/flea.config"} {
		if got := request(NewProductionServer(p, nil), "GET", route).Code; got != 404 {
			t.Fatalf("private configuration route %s returned %d", route, got)
		}
	}
}
func request(s *Server, method, url string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(method, url, nil))
	return w
}

func TestRoutingAndSecurity(t *testing.T) {
	root := fixture(t)
	put(t, root, "_secret.txt", "SECRET")
	put(t, root, ".env", "SECRET")
	put(t, root, ".hidden/file", "SECRET")
	put(t, root, "_private/file", "SECRET")
	put(t, root, "contact.html", "Contact")
	s := NewServer(mustLoad(t, root))
	for _, url := range []string{"/", "/about", "/contact", "/css/main.css"} {
		if w := request(s, "GET", url); w.Code != 200 {
			t.Errorf("%s: %d", url, w.Code)
		}
	}
	for _, url := range []string{"/index.html", "/about.html", "/about/index.html", "/about/", "/missing", "/_secret.txt", "/.env", "/.hidden/file", "/_private/file", "/.itsy.config.json", "/../.itsy.config.json", "/%2e%2e/.itsy.config.json", "/%5fsecret.txt", "/%252e%252e/secret", "//about", "/css/../_secret.txt", "/css%5c..%5c_secret.txt", "/%00"} {
		w := request(s, "GET", url)
		if w.Code != 404 || strings.Contains(w.Body.String(), "SECRET") {
			t.Errorf("unsafe route %s: %d %s", url, w.Code, w.Body.String())
		}
	}
	if w := request(s, "POST", "/"); w.Code != 405 || w.Header().Get("Allow") != "GET, HEAD" {
		t.Fatal(w)
	}
	for _, url := range []string{"/", "/css/main.css", "/missing"} {
		w := request(s, "HEAD", url)
		if w.Body.Len() != 0 {
			t.Fatalf("HEAD %s has body", url)
		}
	}
	put(t, root, "css/main.css", "changed")
	if w := request(s, "GET", "/css/main.css"); w.Body.String() != "body{color:red}" {
		t.Fatal("asset was not a snapshot")
	}
}

func TestCollisionsAndSymlinks(t *testing.T) {
	for _, name := range []string{"about.html", "asset"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			if name == "asset" {
				put(t, root, "asset.html", "page")
			}
			put(t, root, name, "collision")
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), "collision") {
				t.Fatalf("got %v", err)
			}
		})
	}
	for _, name := range []string{"link.txt", "nested", "itsy/snips/link.json", ".itsy.config.json"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			target := filepath.Join(t.TempDir(), "secret")
			if err := os.WriteFile(target, []byte("SECRET"), 0600); err != nil {
				t.Fatal(err)
			}
			dest := filepath.Join(root, "site", name)
			if strings.HasPrefix(name, "itsy/") || name == ".itsy.config.json" {
				dest = filepath.Join(root, name)
			}
			if name == ".itsy.config.json" {
				if err := os.Remove(dest); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(target, dest); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestRenderingScopeAndEscaping(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<html><title>{app.site_name}</title><body>{slot}</body></html>`)
	put(t, root, "_member.html", `[{member.name}:{app.site_name}:{#each member.children as member}{member.name}/{app.site_name}{/each}:{member.name}]`)
	put(t, root, "index.html", `{#each snip.team.members as member}{render member}{/each}<p>{app.unsafe}</p><div title="{app.unsafe}"></div><a href="{app.url}">link</a>{snip team}{#if snip.team.false}BAD{/if}{#if snip.team.zero}BAD{/if}{#if snip.team.blank}BAD{/if}{#if snip.team.empty}BAD{/if}{#each snip.team.empty as x}BAD{/each}{#each snip.missing as x}BAD{/each}`)
	p := mustLoad(t, root)
	b, err := p.render(p.Pages["/"], "/")
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{"[Ada:Test:Bea/Test:Ada][Cy:Test::Cy]", "#ZgotmplZ", "&lt;script&gt;", "&lt;b&gt;{app.site_name}&lt;/b&gt;", "<body>"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "BAD") || strings.Contains(got, "<script>alert") {
		t.Fatal(got)
	}
}

func TestRenderPaths(t *testing.T) {
	root := fixture(t)
	put(t, root, "_footer.html", "ROOT")
	put(t, root, "about/_footer.html", "LOCAL")
	put(t, root, "foo/bar/_footer.html", "NESTED:{render note}:{render ../../footer}")
	put(t, root, "foo/bar/_note.html", "NOTE")
	put(t, root, "main.layout.html", "{slot}|{render footer}")
	put(t, root, "about/index.html", "{render footer}|{render ./footer}|{render /foo/bar/footer}")
	s := NewServer(mustLoad(t, root))
	w := request(s, "GET", "/about")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "LOCAL|LOCAL|NESTED:NOTE:ROOT|ROOT") {
		t.Fatalf("unexpected rendering: %d %s", w.Code, w.Body.String())
	}
	for _, url := range []string{"/_footer", "/_footer.html", "/about/_footer", "/foo/bar/_footer.html", "/foo/bar/footer"} {
		if w := request(s, "GET", url); w.Code != 404 {
			t.Errorf("partial exposed at %s: %d", url, w.Code)
		}
	}
}

func TestInvalidRenderPaths(t *testing.T) {
	for _, name := range []string{"", "../footer", "/../footer", "foo/../../footer", "footer.html", "_footer", `"footer"`, "foo\\footer", "foo/", "foo bar", "/layout", "/404", "/500"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "index.html", "{render "+name+"}")
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "index.html:1:") {
				t.Fatalf("expected error with source location, got %v", err)
			}
		})
	}
	for _, source := range []string{"{render missing}", "{render /foo/../foo/bar/card}"} {
		t.Run(source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "foo/bar/_card.html", source)
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "foo/bar/_card.html:1:") {
				t.Fatalf("expected unused partial validation error, got %v", err)
			}
		})
	}
}

func TestRawElements(t *testing.T) {
	root := fixture(t)
	source := `<script data-value=">">const x={a: "{app.site_name}"}; const t="{{.App}}"; const close="</scripture>"; const untouched="{app.site_name}";</script><STYLE>.x { color:red; }</STYLE><p>{app.site_name}</p>`
	put(t, root, "index.html", source)
	p := mustLoad(t, root)
	compiled, err := p.compile(source, "index.html", nil, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(compiled, `const close="</scripture>"; const untouched="{app.site_name}";`) {
		t.Fatalf("parser changed script body: %s", compiled)
	}
	b, err := p.render(p.Pages["/"], "/")
	if err != nil {
		t.Fatal(err)
	}
	// html/template normalizes this literal to an equivalent JS escape; the
	// FleaCMS parser must leave both it and subsequent braces untouched.
	want := strings.Replace(source, "</scripture>", `\x3C/scripture>`, 1)
	if string(b) != strings.Replace(want, "<p>{app.site_name}</p>", "<p>Test</p>", 1) {
		t.Fatalf("raw content changed: %s", b)
	}
}

func TestInvalidTemplatesAndData(t *testing.T) {
	cases := map[string]string{
		"missing partial": `{render missing}`, "unclosed each": `{#each snip.team.members as m}`,
		"unmatched": `{/if}`, "empty if": `{#if }x{/if}`, "empty each": `{#each }`,
		"non-array": `{#each app.site_name as item}x{/each}`, "reserved alias": `{#each snip.team.members as app}{/each}`,
		"bad alias": `{#each snip.team.members as x-y}{/each}`, "slot": `{slot}`, "unclosed directive": `{app.site_name`,
		"missing snip": `{snip missing}`, "unclosed script": `<script>{anything}`,
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "index.html", source)
			if _, err := Load(root); err == nil {
				t.Fatal("expected useful validation error")
			} else if !strings.Contains(err.Error(), "index.html") {
				t.Fatal(err)
			}
		})
	}
	t.Run("recursive", func(t *testing.T) {
		root := fixture(t)
		put(t, root, "_a.html", `{render b}`)
		put(t, root, "_b.html", `{render a}`)
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "recursive") {
			t.Fatal(err)
		}
	})
	for _, name := range []string{".itsy.config.json", "itsy/snips/team.json"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			put(t, root, name, "{bad}")
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
				t.Fatal(err)
			}
		})
	}
}

func TestReloadAndConcurrentRequests(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				w := request(s, "GET", "/")
				if w.Code != 200 && w.Code != 500 {
					t.Errorf("status %d", w.Code)
				}
			}
		}()
	}
	put(t, root, "index.html", "UPDATED")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(request(s, "GET", "/").Body.String(), "UPDATED") {
		t.Fatal("edit not loaded")
	}
	put(t, root, "index.html", `{render private-error-detail}`)
	if err := s.Reload(root); err == nil {
		t.Fatal("expected reload error")
	}
	w := request(s, "GET", "/")
	if w.Code != 500 || !strings.Contains(w.Body.String(), "Try later") || strings.Contains(w.Body.String(), "private-error-detail") {
		t.Fatal(w)
	}
	put(t, root, "index.html", "RECOVERED")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", "/"); w.Code != 200 || !strings.Contains(w.Body.String(), "RECOVERED") {
		t.Fatal(w)
	}
	wg.Wait()
}

func TestErrorPageLayouts(t *testing.T) {
	root := fixture(t)
	put(t, root, "main.layout.html", `<!doctype html><html><body><header>{app.site_name}</header>{slot}<footer>Shared footer</footer></body></html>`)
	s := NewServer(mustLoad(t, root))
	for _, url := range []string{"/missing", "/_private"} {
		w := request(s, "GET", url)
		if w.Code != 404 || !strings.Contains(w.Body.String(), "<header>Test</header>Not here<footer>Shared footer</footer>") {
			t.Fatalf("404 layout: %d %s", w.Code, w.Body.String())
		}
		if w := request(s, "HEAD", url); w.Code != 404 || w.Body.Len() != 0 {
			t.Fatal("HEAD error response must have no body")
		}
	}
	put(t, root, "index.html", "{#if broken syntax}")
	if err := s.Reload(root); err == nil {
		t.Fatal("expected invalid edit")
	}
	w := request(s, "GET", "/")
	if w.Code != 500 || !strings.HasPrefix(w.Body.String(), "Try later") || strings.Contains(w.Body.String(), "Shared footer") {
		t.Fatal("500 must stay standalone")
	}
	if err := os.Remove(filepath.Join(root, "site", "main.layout.html")); err != nil {
		t.Fatal(err)
	}
	put(t, root, "index.html", "fixed")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	w = request(s, "GET", "/missing")
	if w.Code != 404 || !strings.HasPrefix(w.Body.String(), "Not here") {
		t.Fatal("404 without layout must still work")
	}
}

func TestRenderingFailureFallback(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{#each snip.team.members as m}{m.name}{/each}`)
	p := mustLoad(t, root)
	p.Snips["team"]["members"] = "wrong type"
	w := request(NewServer(p), "GET", "/")
	if w.Code != 500 || !strings.HasPrefix(w.Body.String(), "Try later") {
		t.Fatal(w)
	}
	p.Errors[500] = p.Pages["/"]
	w = request(NewServer(p), "GET", "/")
	if w.Code != 500 || !strings.Contains(w.Body.String(), "Internal Server Error") || strings.Contains(w.Body.String(), "expects a JSON array") {
		t.Fatal(w)
	}
}

func TestWatchChangesAndRecovery(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); s.Watch(ctx, root, log.New(io.Discard, "", 0)) }()
	time.Sleep(50 * time.Millisecond)
	wait := func(check func() bool) {
		t.Helper()
		deadline := time.Now().Add(4 * time.Second)
		for time.Now().Before(deadline) {
			if check() {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatal("watch did not publish change")
	}
	for _, edit := range []struct{ name, body, url, want string }{
		{".itsy.config.json", `{"app":{"site_name":"Renamed"}}`, "/", "Renamed"},
		{"index.html", `{snip.team.text}`, "/", "&lt;b&gt;"},
		{"itsy/snips/team.json", `{"text":"new content"}`, "/", "new content"},
		{"css/main.css", "new css", "/css/main.css", "new css"},
	} {
		put(t, root, edit.name, edit.body)
		wait(func() bool { return strings.Contains(request(s, "GET", edit.url).Body.String(), edit.want) })
	}
	put(t, root, "index.html", "{#if bad syntax}")
	wait(func() bool { return request(s, "GET", "/").Code == 500 })
	put(t, root, "index.html", "fixed")
	wait(func() bool { return request(s, "GET", "/").Code == 200 })
	put(t, root, "_live.html", "partial one")
	put(t, root, "new.html", `{render live}`)
	wait(func() bool { return strings.Contains(request(s, "GET", "/new").Body.String(), "partial one") })
	put(t, root, "_live.html", "partial two")
	wait(func() bool { return strings.Contains(request(s, "GET", "/new").Body.String(), "partial two") })
	if err := os.Remove(filepath.Join(root, "site", "new.html")); err != nil {
		t.Fatal(err)
	}
	wait(func() bool { return request(s, "GET", "/new").Code == 404 })
	if !strings.Contains(request(s, "GET", "/").Body.String(), reloadPath) {
		t.Fatal("missing browser reload script")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watch did not stop")
	}
}

func TestConfinedReads(t *testing.T) {
	root := fixture(t)
	boundary, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer boundary.Close()
	if err := os.Symlink(filepath.Join(root, "_"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := readFile(boundary, "link/flea.config.json"); err == nil {
		t.Fatal("followed a private directory symlink")
	}
	if err := os.Symlink(filepath.Join(root, ".itsy.config.json"), filepath.Join(root, "secret")); err != nil {
		t.Fatal(err)
	}
	if _, err := readFile(boundary, "secret"); err == nil {
		t.Fatal("followed a private file symlink")
	}
}
