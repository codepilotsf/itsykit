package site

import (
	"context"
	"itsy/templates"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Headless checks CSS parsing and computed styles. Native transition lifecycle
// is checked with real link clicks in a visible browser, not --dump-dom.
func TestNavigationStylesBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER to a Chromium executable")
	}
	root := fixture(t)
	css, err := templates.Files.ReadFile("site/css/navigation.css")
	if err != nil {
		t.Fatal(err)
	}
	put(t, root, "css/navigation.css", string(css))
	put(t, root, "main.layout.html", `<!doctype html><html><head><meta charset="utf-8"><link rel="stylesheet" href="/css/navigation.css"><script>
 addEventListener('DOMContentLoaded',()=>{
   const result=document.createElement('output');result.id='navigation-result';
   try {
     const rules=[...document.querySelector('link[rel=stylesheet]').sheet.cssRules];
     if(!rules.some(rule=>rule.cssText.startsWith('@view-transition') && /navigation:\s*auto/.test(rule.cssText))) throw Error('native navigation rule missing');
     if(getComputedStyle(document.querySelector('main')).viewTransitionName!=='itsy-page') throw Error('main not isolated');
     result.textContent='PASS';
   }catch(error){result.textContent='FAIL: '+error.message;}
   document.body.append(result);
 });
 </script></head><body><header>Stable header</header><main>{slot}</main></body></html>`)
	put(t, root, "index.html", `<h1>First page</h1>`)
	s := NewProductionServer(mustLoad(t, root), nil)
	server := httptest.NewServer(s)
	defer server.Close()
	browserCheck(t, browser, server.URL, "navigation-result")
}

func TestDevErrorBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER to a Chromium executable")
	}
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	put(t, root, "index.html", "<main>\n  <h1>Meet our team</h1>\n  {render team-members}\n</main>")
	if err := s.Reload(root); err == nil {
		t.Fatal("expected missing partial")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.ServeHTTP(w, r) }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	viewport := os.Getenv("FLEA_TEST_VIEWPORT")
	if viewport == "" {
		viewport = "1280,1000"
	}
	args := []string{"--headless", "--disable-gpu", "--disable-background-networking", "--disable-component-update", "--no-first-run", "--no-default-browser-check", "--user-data-dir=" + t.TempDir(), "--window-size=" + viewport, "--dump-dom", "--virtual-time-budget=1200"}
	if screenshot := os.Getenv("FLEA_TEST_SCREENSHOT"); screenshot != "" {
		args = append(args, "--screenshot="+screenshot)
	}
	cmd := exec.CommandContext(ctx, browser, append(args, server.URL)...)
	cmd.WaitDelay = time.Second
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), "How to fix it") || !strings.Contains(string(output), "index.html:3") {
		t.Fatalf("dev error browser: %v", err)
	}
}

func browserCheck(t *testing.T, browser, url, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--disable-background-networking", "--disable-component-update", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", "--virtual-time-budget=3000", url)
	cmd.WaitDelay = time.Second
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), `<output id="`+id+`">PASS</output>`) {
		result := regexp.MustCompile(`<output id="` + id + `">[^<]*</output>`).Find(output)
		page := regexp.MustCompile(`<h1>[^<]*</h1>`).Find(output)
		t.Fatalf("browser check: %v; result: %s; page: %s", err, result, page)
	}
}
