package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	stdhtml "html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	xhtml "golang.org/x/net/html"

	"flea/internal/editor"
	"flea/internal/editstore"
	"flea/internal/release"
	"flea/internal/site"
	"flea/internal/snips"
)

// hostNativeExecutable returns a minimal valid native header for this platform
// so read-only load tests exercise the format check without building anything.
func hostNativeExecutable() []byte {
	if runtime.GOOS == "darwin" {
		cpu := uint32(0x01000007)
		if runtime.GOARCH == "arm64" {
			cpu = 0x0100000c
		}
		h := make([]byte, 32)
		binary.LittleEndian.PutUint32(h[0:], 0xfeedfacf)
		binary.LittleEndian.PutUint32(h[4:], cpu)
		return h
	}
	class, machine := byte(2), uint16(62)
	if runtime.GOARCH == "arm64" {
		class, machine = 2, 183
	} else if runtime.GOARCH == "386" {
		class, machine = 1, 3
	} else if runtime.GOARCH == "arm" {
		class, machine = 1, 40
	}
	h := make([]byte, 64)
	copy(h, "\x7fELF")
	h[4], h[5] = class, 1
	binary.LittleEndian.PutUint16(h[18:], machine)
	return h
}

func appRelease(t *testing.T, base, revision string, withApp bool) (string, string) {
	t.Helper()
	root := filepath.Join(base, "releases", revision)
	config := `{"app":{"site_name":"Example"}}`
	if withApp {
		config = `{"app":{"site_name":"Example"},"apps":{"demo":{"mount":"/demo"}}}`
	}
	files := map[string]string{
		".flea.config.json":    config,
		"site/index.html":      "<h1>Home</h1>",
		"site/demo/index.html": "<p>Demo</p>",
	}
	if withApp {
		files["flea/apps/demo/app.json"] = `{"name":"demo","version":"1.0.0","protocol":2,"executable":{"` + runtime.GOOS + "-" + runtime.GOARCH + `":"demo"}}`
	}
	for name, body := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if withApp {
		if err := os.WriteFile(filepath.Join(root, "flea", "apps", "demo", "demo"), hostNativeExecutable(), 0555); err != nil {
			t.Fatal(err)
		}
	}
	store := snips.NewStore()
	data, err := store.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	content := filepath.Join(base, "content", revision+".json")
	if err := os.MkdirAll(filepath.Dir(content), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(content, data, 0644); err != nil {
		t.Fatal(err)
	}
	return root, content
}

func appConfig(t *testing.T, dir string, root, content string) string {
	t.Helper()
	config := Config{Listen: "127.0.0.1:8080", Sites: map[string]string{"demo.test": root, "alias.demo.test": root}, Content: map[string]string{"demo.test": content, "alias.demo.test": content}}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "server.json")
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	return file
}

// TestCustomAppRequiresManagedContent proves production refuses a custom app
// without a managed content snapshot, and that loading a valid managed site is
// side-effect-free and never writes storage into the immutable release.
func TestCustomAppRequiresManagedContent(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "sites", "example")
	root, content := appRelease(t, base, "first", true)

	// No managed content: refused before any app could run.
	config := Config{Listen: "127.0.0.1:8080", Sites: map[string]string{"demo.test": root}}
	data, _ := json.Marshal(config)
	file := filepath.Join(dir, "server.json")
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(file, nil); err == nil {
		t.Fatal("custom app accepted without managed content")
	}

	// Managed content: accepted, and Load creates no runtime storage.
	file = appConfig(t, dir, root, content)
	if _, _, err := Load(file, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "flea", "runtime")); !os.IsNotExist(err) {
		t.Fatal("read-only load created runtime state inside the release")
	}
}

// TestCustomAppStorageSurvivesReleaseChange proves two releases of the same site
// resolve to one stable runtime storage location outside the release trees.
func TestCustomAppStorageSurvivesReleaseChange(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "sites", "example")
	runtimeRoot := filepath.Join(base, "flea", "runtime")
	for _, revision := range []string{"first", "second"} {
		root, content := appRelease(t, base, revision, true)
		if got := site.ManagedRuntimeRoot(content); got != runtimeRoot {
			t.Fatalf("runtime root %q, want %q", got, runtimeRoot)
		}
		file := appConfig(t, dir, root, content)
		if _, _, err := Load(file, nil); err != nil {
			t.Fatal(revision, err)
		}
		if _, err := os.Stat(filepath.Join(root, "flea", "runtime")); !os.IsNotExist(err) {
			t.Fatal("runtime state entered a release")
		}
	}
	if _, err := os.Stat(runtimeRoot); !os.IsNotExist(err) {
		t.Fatal("read-only load created the managed runtime root")
	}
}

// buildSample compiles the independent stdlib sample app for this platform.
func buildSample(t *testing.T) string {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain unavailable")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate repository root")
	}
	repo := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	out := filepath.Join(t.TempDir(), "session")
	cmd := exec.Command(goTool, "build", "-o", out, "./samples/session-app")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if combined, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building sample app: %v\n%s", err, combined)
	}
	return out
}

// sampleReleaseFiles builds a real app package from an explicit site file map,
// ships it through Pack and Extract (the production path), adds the session app
// manifest and executable, and returns the extracted release plus managed
// content. It is the shared production-path harness behind sampleRelease.
func sampleReleaseFiles(t *testing.T, base, revision, binary string, files map[string]string) (string, string) {
	t.Helper()
	src := filepath.Join(base, "src", revision)
	write := func(name string, body []byte, mode os.FileMode) {
		t.Helper()
		dest := filepath.Join(src, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, body, mode); err != nil {
			t.Fatal(err)
		}
	}
	for name, body := range files {
		write(name, []byte(body), 0644)
	}
	write("flea/apps/session/app.json", []byte(`{"name":"session","version":"0.1.0","protocol":2,"executable":{"`+runtime.GOOS+"-"+runtime.GOARCH+`":"session"}}`), 0644)
	body, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	write("flea/apps/session/session", body, 0755)

	var packed bytes.Buffer
	if err := release.Pack(src, &packed); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "releases", revision)
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := release.Extract(bytes.NewReader(packed.Bytes()), root); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, "flea", "apps", "session", "session"))
	if err != nil || info.Mode().Perm() != 0555 {
		t.Fatalf("packaged executable mode %v %v", info, err)
	}
	store := snips.NewStore()
	contentData, err := store.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	content := filepath.Join(base, "content", revision+".json")
	if err := os.MkdirAll(filepath.Dir(content), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(content, contentData, 0644); err != nil {
		t.Fatal(err)
	}
	return root, content
}

// sampleRelease builds the minimal single-app release used by the lifecycle
// test: one session app mounted at /session.
func sampleRelease(t *testing.T, base, revision, binary, page string) (string, string) {
	t.Helper()
	return sampleReleaseFiles(t, base, revision, binary, map[string]string{
		".flea.config.json":       `{"app":{"site_name":"Example"},"apps":{"session":{"mount":"/session"}}}`,
		"site/session/index.html": page,
	})
}

func hostConfigFile(t *testing.T, dir string, sites, content map[string]string) string {
	t.Helper()
	config := Config{Listen: "127.0.0.1:8080", Sites: sites, Content: content}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "server.json")
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	return file
}

func hostFormToken(body string) string {
	_, rest, _ := strings.Cut(body, `name="_csrf" value="`)
	token, _, _ := strings.Cut(rest, `"`)
	return token
}

func hostCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// startSampleHost loads a host config for demo.test (plus any aliases) and
// launches its apps. The returned halt stops the apps; it also runs on cleanup.
func startSampleHost(t *testing.T, dir, root, content string, aliases map[string]string) (http.Handler, func()) {
	t.Helper()
	sites := map[string]string{"demo.test": root}
	contentMap := map[string]string{"demo.test": content}
	for host, aliasRoot := range aliases {
		sites[host] = aliasRoot
		contentMap[host] = content
	}
	file := hostConfigFile(t, dir, sites, contentMap)
	_, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	rh, ok := handler.(*runtimeHandler)
	if !ok {
		t.Fatal("unexpected host handler")
	}
	ctx, cancel := context.WithCancel(context.Background())
	stop := rh.start(ctx)
	halt := func() { cancel(); stop() }
	t.Cleanup(halt)
	return handler, halt
}

// sampleHostSite returns the single public site server a test host loaded. It is
// the same server the host wires into its editor Backend for Publish.
func sampleHostSite(t *testing.T, handler http.Handler) *site.Server {
	t.Helper()
	rh, ok := handler.(*runtimeHandler)
	if !ok || len(rh.sites) == 0 {
		t.Fatal("host has no site server")
	}
	return rh.sites[0]
}

func hostGet(handler http.Handler, host, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "https://"+host+path, nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func hostPost(handler http.Handler, host, path, action, token string, formCookie, session *http.Cookie) *httptest.ResponseRecorder {
	body := url.Values{"action": {action}, "_csrf": {token}}
	r := httptest.NewRequest("POST", "https://"+host+path, strings.NewReader(body.Encode()))
	r.Header.Set("Origin", "https://"+host)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if formCookie != nil {
		r.AddCookie(formCookie)
	}
	if session != nil {
		r.AddCookie(session)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func waitAppReady(t *testing.T, handler http.Handler, host string) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		if w := hostGet(handler, host, "/session"); w.Code == 200 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("app %s did not become ready", host)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

// TestCustomAppHostLifecycleAcrossReleases drives the real sample app through the
// production host: Pack, Extract, Load, launch and an app-owned session, then a
// release/content change under the same stable site. The session survives, an
// alias route works, a separate site does not accept the token, and the archived
// executable is service-readable.
func TestCustomAppHostLifecycleAcrossReleases(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and launches the sample app")
	}
	binary := buildSample(t)
	dir := t.TempDir()
	base := filepath.Join(dir, "sites", "example")
	pageA := `<p id="state">{session.data.authenticated}</p><p id="rev">A</p><form method="post" action="/session">{csrf}<input name="action" value="login"></form>`
	pageB := strings.Replace(pageA, `<p id="rev">A</p>`, `<p id="rev">B</p>`, 1)
	rootA, contentA := sampleRelease(t, base, "first", binary, pageA)
	rootB, contentB := sampleRelease(t, base, "second", binary, pageB)

	handlerA, haltA := startSampleHost(t, dir, rootA, contentA, nil)
	waitAppReady(t, handlerA, "demo.test")
	w := hostGet(handlerA, "demo.test", "/session")
	if !strings.Contains(w.Body.String(), "false") || !strings.Contains(w.Body.String(), "A") {
		t.Fatalf("release A read: %d %s", w.Code, w.Body.String())
	}
	token := hostFormToken(w.Body.String())
	formCookie := w.Result().Cookies()[0]
	w = hostPost(handlerA, "demo.test", "/session", "login", token, formCookie, nil)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	session := hostCookie(w, "__Host-flea-session-session")
	if session == nil || session.Value == "" {
		t.Fatalf("login set no session cookie: %v", w.Result().Cookies())
	}
	if !strings.Contains(hostGet(handlerA, "demo.test", "/session", session).Body.String(), "true") {
		t.Fatal("live token did not authenticate")
	}
	haltA()

	// Release B under the same stable site: the persisted token still works.
	handlerB, haltB := startSampleHost(t, dir, rootB, contentB, map[string]string{"alias.demo.test": rootB})
	defer haltB()
	waitAppReady(t, handlerB, "demo.test")
	if body := hostGet(handlerB, "demo.test", "/session", session).Body.String(); !strings.Contains(body, "true") {
		t.Fatalf("session did not survive the release change: %s", body)
	}
	if body := hostGet(handlerB, "demo.test", "/session").Body.String(); !strings.Contains(body, "B") {
		t.Fatalf("release B was not served: %s", body)
	}
	if body := hostGet(handlerB, "alias.demo.test", "/session", session).Body.String(); !strings.Contains(body, "true") {
		t.Fatalf("alias route rejected the session: %s", body)
	}

	// A separate site with its own runtime must not accept another site's token.
	base2 := filepath.Join(dir, "sites", "other")
	root2, content2 := sampleRelease(t, base2, "first", binary, pageA)
	handler2, halt2 := startSampleHost(t, dir, root2, content2, nil)
	defer halt2()
	waitAppReady(t, handler2, "demo.test")
	if body := hostGet(handler2, "demo.test", "/session", session).Body.String(); !strings.Contains(body, "false") {
		t.Fatalf("cross-site token accepted: %s", body)
	}
}

// htmlAttrValue returns the value of an attribute on the element carrying the
// given id, parsed with an HTML tokenizer so the result matches browser decoding.
func htmlAttrValue(t *testing.T, body, id, key string) string {
	t.Helper()
	z := xhtml.NewTokenizer(strings.NewReader(body))
	for z.Next() != xhtml.ErrorToken {
		tok := z.Token()
		if tok.Type != xhtml.StartTagToken && tok.Type != xhtml.SelfClosingTagToken {
			continue
		}
		matched := false
		for _, a := range tok.Attr {
			if a.Key == "id" && a.Val == id {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		for _, a := range tok.Attr {
			if a.Key == key {
				return a.Val
			}
		}
	}
	t.Fatalf("attribute %s of #%s not found", key, id)
	return ""
}

// TestBlogSaveKeepsCustomAppSessionAndManagedRuntime proves the combined
// save/runtime boundary: a site that enables the built-in Blog and a real custom
// app together, with a literal-JSON attribute on the app page, keeps the app's
// session and the managed stable runtime root across an owner Blog save and a
// full host reload. The raw JSON attribute is decoded the way a browser would.
func TestBlogSaveKeepsCustomAppSessionAndManagedRuntime(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and launches the sample app")
	}
	binary := buildSample(t)
	dir := t.TempDir()
	base := filepath.Join(dir, "sites", "example")
	page := `<div id="app-config" data-config='[{"id":"session","kind":"toggle","default":true}]'></div>` +
		`<p id="state">{session.data.authenticated}</p>` +
		`<form method="post" action="/session">{csrf}<input name="action" value="login"></form>`
	root, content := sampleReleaseFiles(t, base, "first", binary, map[string]string{
		".flea.config.json":       `{"app":{"site_name":"Example"},"apps":{"blog":{"mount":"/blog"},"session":{"mount":"/session"}}}`,
		"site/session/index.html": page,
		"site/blog/_index.html":   `<main id="posts">{#each blog.data.posts as post}<a href="{post.url}">{post.title}</a>{/each}</main>`,
		"site/blog/_post.html":    `<main id="post"><h1>{blog.data.post.title}</h1>{blog.data.post.body}</main>`,
	})

	// Custom-app storage is stable per-site runtime state derived from managed
	// content, never the immutable release, so it survives a release or save.
	runtimeRoot := site.ManagedRuntimeRoot(content)
	if want := filepath.Join(base, "flea", "runtime"); runtimeRoot != want {
		t.Fatalf("managed runtime root %q, want %q", runtimeRoot, want)
	}
	if strings.HasPrefix(runtimeRoot, root) {
		t.Fatalf("runtime root inside the immutable release: %q", runtimeRoot)
	}

	handler, halt := startSampleHost(t, dir, root, content, nil)
	waitAppReady(t, handler, "demo.test")

	// The app page renders its JSON data and a literal-JSON attribute whose
	// braces must survive unmodified and still decode as JSON in a browser.
	w := hostGet(handler, "demo.test", "/session")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `<p id="state">false</p>`) {
		t.Fatalf("app read: %d %s", w.Code, w.Body.String())
	}
	attribute := htmlAttrValue(t, w.Body.String(), "app-config", "data-config")
	var decoded []map[string]any
	if err := json.Unmarshal([]byte(stdhtml.UnescapeString(attribute)), &decoded); err != nil {
		t.Fatalf("literal-JSON attribute did not decode: %q: %v", attribute, err)
	}
	if len(decoded) != 1 || decoded[0]["id"] != "session" || decoded[0]["kind"] != "toggle" || decoded[0]["default"] != true {
		t.Fatalf("literal-JSON attribute changed: %#v", decoded)
	}

	// Establish an app-owned session the way the browser does.
	token := hostFormToken(w.Body.String())
	formCookie := w.Result().Cookies()[0]
	w = hostPost(handler, "demo.test", "/session", "login", token, formCookie, nil)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	session := hostCookie(w, "__Host-flea-session-session")
	if session == nil || session.Value == "" {
		t.Fatalf("login set no session cookie: %v", w.Result().Cookies())
	}
	if body := hostGet(handler, "demo.test", "/session", session).Body.String(); !strings.Contains(body, `<p id="state">true</p>`) {
		t.Fatalf("live token did not authenticate: %s", body)
	}
	sessionFile := filepath.Join(runtimeRoot, "apps", "session", "sessions.json")
	if _, err := os.Stat(sessionFile); err != nil {
		t.Fatalf("app session store is not under the managed runtime root: %v", err)
	}

	// An owner Blog save republishes the site through the same editstore path the
	// host wires for admin edits.
	store := &editstore.Store{Root: root, ContentPath: content, ConfigPath: filepath.Join(dir, "server.json"), Publish: sampleHostSite(t, handler).Publish}
	doc, err := store.Read()
	if err != nil {
		t.Fatal(err)
	}
	edit := editor.BlogEdit{Version: doc.Version, Post: snips.BlogPost{ID: "Abcdef1234", Title: "Reconciled", Slug: "reconciled", Body: "<p>Saved after login</p>", Status: "published"}, Action: "publish"}
	var result editor.BlogResult
	if err := store.Save(editor.Update{Version: doc.Version, Blog: &edit, BlogResult: &result}); err != nil {
		t.Fatal(err)
	}

	// The save published the post without disturbing the running app, its
	// session or the literal-JSON attribute.
	if body := hostGet(handler, "demo.test", "/blog/reconciled").Body.String(); !strings.Contains(body, "Reconciled") || !strings.Contains(body, "Saved after login") {
		t.Fatalf("blog save was not published: %s", body)
	}
	if body := hostGet(handler, "demo.test", "/session", session).Body.String(); !strings.Contains(body, `<p id="state">true</p>`) {
		t.Fatalf("session lost after the blog save: %s", body)
	}
	if body := hostGet(handler, "demo.test", "/session").Body.String(); !strings.Contains(body, "data-config") {
		t.Fatalf("app page lost its literal-JSON attribute: %s", body)
	}

	// A full host reload keeps the same managed runtime root, and the app session
	// persists through the app restart and the updated managed content.
	halt()
	reloaded, haltReloaded := startSampleHost(t, dir, root, content, nil)
	defer haltReloaded()
	waitAppReady(t, reloaded, "demo.test")
	if site.ManagedRuntimeRoot(content) != runtimeRoot {
		t.Fatal("managed runtime root changed across reload")
	}
	if _, err := os.Stat(filepath.Join(root, "flea", "runtime")); !os.IsNotExist(err) {
		t.Fatal("runtime state entered the immutable release")
	}
	if body := hostGet(reloaded, "demo.test", "/session", session).Body.String(); !strings.Contains(body, `<p id="state">true</p>`) {
		t.Fatalf("session did not survive the reload: %s", body)
	}
	if body := hostGet(reloaded, "demo.test", "/blog/reconciled").Body.String(); !strings.Contains(body, "Reconciled") {
		t.Fatalf("blog post lost after reload: %s", body)
	}
}
