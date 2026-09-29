package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"flea/internal/site"
	"flea/templates"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type previewBackend struct {
	memoryStore
	project *site.Project
}

func (b *previewBackend) Read() (Document, error) {
	return Document{Store: b.project.Content, Routes: b.project.Routes(), Version: b.project.EditingVersion()}, nil
}

func TestEditorContextAndAssets(t *testing.T) {
	if bytes.Contains(editorAssets["toolbar.html"], []byte("<style>")) || bytes.Contains(editorAssets["toolbar.html"], []byte("<script>")) {
		t.Fatal("toolbar asset contains inline code")
	}
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	csrf := h.sessions[key(cookie.Value)].csrf
	for _, tc := range []struct {
		cookie *http.Cookie
		csrf   string
		code   int
	}{{nil, csrf, 404}, {cookie, "", 403}, {cookie, "wrong", 403}, {cookie, csrf, 200}} {
		r := httptest.NewRequest("GET", "https://example.test/admin/editor-context?route=/", nil)
		if tc.cookie != nil {
			r.AddCookie(tc.cookie)
		}
		r.Header.Set("X-CSRF-Token", tc.csrf)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.code {
			t.Fatalf("context status %d want %d: %s", w.Code, tc.code, w.Body.String())
		}
		if w.Code == 200 && w.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatal("private context cached")
		}
	}
	for name := range editorAssets {
		r := httptest.NewRequest("GET", "https://example.test/admin/assets/"+name, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("ETag") == "" || strings.Contains(w.Body.String(), csrf) {
			t.Fatal("invalid public asset", name, w.Code)
		}
		r.Header.Set("If-None-Match", w.Header().Get("ETag"))
		w = httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 304 {
			t.Fatal("asset not cacheable", name, w.Code)
		}
	}
}

func TestExternalEditorBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	// Keep browser QA offline; third-party font/style availability is unrelated.
	for _, name := range []string{"site/css/global.css", "site/css/theme.css"} {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		data = regexp.MustCompile(`@import[^;]+https[^;]+;`).ReplaceAll(data, nil)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	var photo bytes.Buffer
	if err := png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site/images/editor-test.png"), photo.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(root, "site/index.html")
	pageSource, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	pageSource = append(pageSource, []byte(`<img snip src="/images/editor-test.png" alt="Real extracted image" class="rounded-lg">`)...)
	if err = os.WriteFile(pagePath, pageSource, 0644); err != nil {
		t.Fatal(err)
	}
	p, err := site.LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewProductionServer(p, nil)
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	h.options.Public = public.ServeOwner
	h.options.Backend = &previewBackend{project: p}
	// Hold editor data until the browser verifies the server-rendered shell.
	ready := make(chan struct{}, 1)
	checks := `addEventListener('load',()=>{let n=0;const timer=setInterval(()=>{const host=document.getElementById('flea-owner-root');if((!host?.shadowRoot||!host.mediaSnips) && ++n<50)return;clearInterval(timer);const out=document.createElement('output');out.id='external-result';try{if(!host?.shadowRoot)throw Error(host?.textContent||'no toolbar');const ui=host.shadowRoot;if(!ui.querySelector('#edit'))throw Error('no edit button');if(!document.querySelector('meta[name="csrf-token"]')?.content)throw Error('no CSRF');if(!ui.querySelector('link[rel=stylesheet]'))throw Error('no external CSS');if(getComputedStyle(ui.querySelector('#edit')).appearance!=='none')throw Error('switch CSS not loaded');if(!ui.querySelector('#edit').checked)throw Error('editing not on by default');if(!document.querySelector('[contenteditable=true]'))throw Error('not editable');if(ui.querySelector('#photos,#videos'))throw Error('obsolete library buttons');const image=document.querySelector('img[data-snip-media]');if(!image)throw Error('no rendered media binding');image.click();if(!ui.querySelector('#media-dialog').open||ui.querySelector('#media-alt').value!=='Real extracted image')throw Error('media context not loaded');out.textContent='PASS'}catch(e){out.textContent='FAIL '+e.message}document.body.append(out)},50)});`
	checks = `addEventListener('load',()=>{const host=document.getElementById('flea-owner-root'),out=document.createElement('output');out.id='shell-result';try{if(!host?.shadowRoot?.querySelector('section')?.inert)throw Error('shell missing before hydration');if(host.shadowRoot.querySelectorAll('button[data-icon] svg').length!==6)throw Error('icons missing before hydration');window.initialToolbarIcons=[...host.shadowRoot.querySelectorAll('button[data-icon] svg')];const height=host.getBoundingClientRect().height;if(height<90||[...host.shadowRoot.querySelectorAll('nav')].some(nav=>nav.getBoundingClientRect().height<40))throw Error('toolbar collapsed before context');out.textContent='PASS';}catch(e){out.textContent='FAIL '+e.message}document.body.append(out);fetch('/shell-ready');});` + checks
	if os.Getenv("FLEA_TEST_SCREENSHOT") != "" {
		checks = strings.Replace(checks, "out.textContent='PASS'}catch", "ui.querySelector('#media-dialog').close();out.textContent='PASS'}catch", 1)
	}
	checks = strings.Replace(checks, "window.initialToolbarIcons=", "if(host.shadowRoot.querySelector('#edit').checked)throw Error('off preference flashed before hydration');sessionStorage.removeItem('flea-editing:'+host.dataset.email);window.initialToolbarIcons=", 1)
	if os.Getenv("FLEA_TEST_EDIT_ON") == "1" {
		checks = strings.Replace(checks, "if(host.shadowRoot.querySelector('#edit').checked)throw Error('off preference flashed before hydration');", "if(!host.shadowRoot.querySelector('#edit').checked||host.shadowRoot.querySelector('#details').disabled)throw Error('Edit or Page Details flashed off before hydration');", 1)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/shell-ready" {
			select {
			case ready <- struct{}{}:
			default:
			}
			w.WriteHeader(204)
			return
		}
		if r.URL.Path == "/admin/editor-context" {
			select {
			case <-ready:
			case <-time.After(3 * time.Second):
				http.Error(w, "browser did not verify shell", 503)
				return
			case <-r.Context().Done():
				return
			}
		}
		auth := httptest.NewRequest(r.Method, "https://example.test"+r.URL.RequestURI(), r.Body)
		auth.Header = r.Header.Clone()
		auth.AddCookie(cookie)
		if IsPath(r.URL.Path) {
			h.ServeHTTP(w, auth)
			return
		}
		recorded := httptest.NewRecorder()
		h.ServePublic(recorded, auth, public)
		for k, v := range recorded.Header() {
			w.Header()[k] = v
		}
		// The fixture authenticates example.test but Chrome visits its loopback server.
		w.Header().Set("Content-Security-Policy", strings.ReplaceAll(w.Header().Get("Content-Security-Policy"), "https://example.test/admin/assets/editor.js", "http://"+r.Host+"/admin/assets/editor.js"))
		body := recorded.Body.Bytes()
		if r.URL.Path == "/" {
			seed := `sessionStorage.setItem('flea-editing:owner@example.test','off');`
			if os.Getenv("FLEA_TEST_EDIT_ON") == "1" {
				seed = strings.Replace(seed, "'off'", "'on'", 1)
			}
			seedHash := sha256.Sum256([]byte(seed))
			w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(seedHash[:])+"'")
			body = bytes.Replace(body, []byte("<head>"), []byte("<head><script>"+seed+"</script>"), 1)
			w.Header().Del("Content-Length")
			hash := sha256.Sum256([]byte(checks))
			w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(hash[:])+"'")
			body = append(bytes.Clone(body), []byte("<script>"+checks+"</script>")...)
		}
		w.WriteHeader(recorded.Code)
		w.Write(body)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	args := []string{"--headless", "--disable-gpu", "--disable-background-networking", "--no-first-run", "--no-default-browser-check", "--user-data-dir=" + t.TempDir(), "--dump-dom", "--virtual-time-budget=5000"}
	if screenshot := os.Getenv("FLEA_TEST_SCREENSHOT"); screenshot != "" {
		size := os.Getenv("FLEA_TEST_WINDOW_SIZE")
		if size == "" {
			size = "1280,1000"
		}
		args = append(args, "--window-size="+size, "--screenshot="+screenshot)
	}
	cmd := exec.CommandContext(ctx, browser, append(args, server.URL)...)
	cmd.WaitDelay = time.Second
	output, err := cmd.CombinedOutput()
	if !bytes.Contains(output, []byte(`<output id="external-result">PASS</output>`)) {
		t.Fatalf("external editor: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte(`<output id="shell-result">PASS</output>`)) {
		t.Fatalf("toolbar unavailable before bootstrap: %s", output)
	}
}
