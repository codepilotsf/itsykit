package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"flea/internal/snips"
	"flea/templates"
)

// Exercise native cross-document transitions, not only CSS parsing.
func TestOwnerNavigationBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	css, err := templates.Files.ReadFile("site/css/navigation.css")
	if err != nil {
		t.Fatal(err)
	}
	layout, err := templates.Files.ReadFile("site/main.layout.html")
	if err != nil {
		t.Fatal(err)
	}
	start := bytes.Index(layout, []byte("<style>@view-transition"))
	if start < 0 {
		t.Fatal("missing early transition opt-in")
	}
	end := bytes.Index(layout[start:], []byte("</style>"))
	if end < 0 {
		t.Fatal("unclosed transition opt-in")
	}
	optIn := string(layout[start : start+end+len("</style>")])
	checks := `
addEventListener('pagereveal',async event=>{
 if(location.pathname!=='/next')return;
 let result='PASS';
 try {
  if(!event.viewTransition)throw Error('No native navigation transition');
  await event.viewTransition.ready;
  if(getComputedStyle(document.querySelector('main')).viewTransitionName!=='flea-page')throw Error('Missing page transition name');
  if(getComputedStyle(document.documentElement,'::view-transition-new(flea-page)').animationName!=='flea-page-in')throw Error('Missing content animation');
  await event.viewTransition.finished;
 }catch(error){result='FAIL '+error.message;}
 fetch('/test-result?value='+encodeURIComponent(result));
});
addEventListener('load',()=>{if(location.pathname==='/')setTimeout(()=>document.querySelector('a').click(),500);});`
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		body := []byte(`<!doctype html><html><head><link rel="stylesheet" href="/css/navigation.css">` + optIn + `<link rel="stylesheet" href="/css/global.css"><script>` + checks + `</script></head><body><header><a href="/next">Next page</a></header><main><h1><!--snip-start:a7Km2pQ9xB-->Page heading<!--snip-end:a7Km2pQ9xB--></h1></main></body></html>`)
		body, ok := browserDecorate(w, r, decorate)(body, snips.NewStore(), "v")
		if !ok {
			return
		}
		body = bytes.Replace(body, append(append([]byte("<style>"), editorAssets["editor.css"]...), []byte("</style>")...), []byte(`<link rel="stylesheet" href="/admin/assets/editor.css">`), 1)
		digest := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(digest[:])+"'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(body)
	}
	results := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serveEditorAsset(w, r) {
			return
		}
		if r.URL.Path == "/css/global.css" {
			w.Header().Set("Content-Type", "text/css")
			w.Write([]byte(`body { margin: 0; }`))
			return
		}
		if r.URL.Path == "/css/navigation.css" {
			w.Header().Set("Content-Type", "text/css")
			w.Write(css)
			return
		}
		if r.URL.Path == "/test-result" {
			select {
			case results <- r.URL.Query().Get("value"):
			default:
			}
			w.WriteHeader(204)
			return
		}
		if r.URL.Path != "/" && r.URL.Path != "/next" {
			http.NotFound(w, r)
			return
		}
		auth := httptest.NewRequest("GET", "https://example.test"+r.URL.Path, nil)
		auth.AddCookie(cookie)
		h.ServePublic(w, auth, http.NotFoundHandler())
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--disable-background-networking", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), server.URL)
	cmd.WaitDelay = time.Second
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case result := <-results:
		cancel()
		<-done
		if result != "PASS" {
			t.Fatal(result)
		}
	case err := <-done:
		t.Fatalf("navigation result missing: %v; %s", err, strings.TrimSpace(output.String()))
	}
}
