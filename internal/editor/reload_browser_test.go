package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"flea/internal/snips"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestOwnerReloadBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	checks := `setTimeout(async()=>{
 const host=document.getElementById('flea-owner-root'),ui=host.shadowRoot;
 const wait=ms=>new Promise(r=>setTimeout(r,ms));
 const report=value=>fetchResult('/test-result?value='+encodeURIComponent(value));const fetchResult=window.fetch.bind(window);
 const assert=(ok,msg)=>{if(!ok)throw new Error(msg)};
 const result=document.createElement('output');result.id='reload-result';
 try {
  if(location.pathname==='/off-check'){assert(!ui.querySelector('#edit').checked&&!document.getElementById('region').isContentEditable,'off preference lost across pages');ui.querySelector('#edit').click();location.href='/on-check';return;}
  if(location.pathname==='/on-check'){assert(ui.querySelector('#edit').checked&&document.getElementById('region').isContentEditable,'on preference lost across pages');result.textContent='PASS';document.body.append(result);await report('PASS');return;}
  if(sessionStorage.getItem('reload-ready')){
    sessionStorage.removeItem('reload-ready');
    assert(ui.querySelector('#edit').checked,'edit mode not restored');
    assert(document.getElementById('new-region').isContentEditable,'new snip not editable');
    ui.querySelector('#edit').click();location.href='/off-check';return;
  }
  window.fetch=async url=>url==='/_flea/reload'?{ok:true,text:async()=>'2'}:{ok:true,json:async()=>({version:'v',fields:[],scales:{}})};
  window.confirm=()=>true;
  assert(ui.querySelector('#edit').checked,'editing defaults on');
  const region=document.getElementById('region');region.textContent='Draft';region.dispatchEvent(new Event('input'));
  await wait(1000);
  assert(region.textContent==='Draft' && ui.querySelector('#status').textContent.includes('Source changed'),'draft not protected');
  ui.querySelector('#cancel').click();ui.querySelector('#theme').click();
  await wait(1000);assert(!ui.querySelector('#theme-panel').hidden,'theme preview not protected');
  ui.querySelector('#theme-cancel').click();
  sessionStorage.setItem('reload-ready','yes');
 }catch(e){result.textContent='FAIL: '+e.message;document.body.append(result);await report(result.textContent);}
},100);`
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		body := []byte(`<!doctype html><html><head><title>Reload test</title></head><body><h2 id="region"><!--snip-start:a7Km2pQ9xB-->Original<!--snip-end:a7Km2pQ9xB--></h2><h2 id="new-region"><!--snip-start:b7Km2pQ9xB-->New snip<!--snip-end:b7Km2pQ9xB--></h2></body></html>`)
		body, ok := browserDecorate(w, r, decorate)(body, snips.NewStore(), "v")
		if !ok {
			return
		}
		body = bytes.Replace(body, []byte(`data-revision=""`), []byte(`data-revision="1"`), 1)
		digest := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(digest[:])+"'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(append(body, []byte("<script>"+checks+"</script>")...))
	}
	results := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test-result" {
			select {
			case results <- r.URL.Query().Get("value"):
			default:
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path == "/_flea/reload" {
			w.Write([]byte("1"))
			return
		}
		authenticated := httptest.NewRequest("GET", "https://example.test/", nil)
		authenticated.AddCookie(cookie)
		h.ServePublic(w, authenticated, http.NotFoundHandler())
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Use real time and an explicit result endpoint: dump-dom virtual time can
	// stop advancing after repeated cross-document navigation.
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--disable-background-networking", "--disable-component-update", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), server.URL)
	cmd.WaitDelay = 2 * time.Second
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
		t.Fatalf("reload browser ended without result: %v\n%s", err, output.Bytes())
	}
}
