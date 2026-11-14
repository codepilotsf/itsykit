package editor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"itsy/internal/media"
	"itsy/internal/snips"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPhotosBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	checks := `setTimeout(async()=>{const result=document.createElement('output');result.id='photo-result';const assert=(v,m)=>{if(!v)throw Error(m)},wait=()=>new Promise(r=>setTimeout(r,350));try{
 const host=document.getElementById('flea-owner-root'),ui=host.shadowRoot,$=s=>ui.querySelector(s),image=document.getElementById('portrait');let sent,body;
 window.fetch=async(url,options)=>{if(url!=='/admin/page')return {ok:true};body=options.body;sent=JSON.parse(body.get('edit'));assert(options.headers['X-CSRF-Token']===host.dataset.csrf,'missing CSRF header');return {ok:false,status:409};};window.confirm=()=>true;
 assert(!$('#photos')&&!$('#videos'),'library buttons remain');
 assert($('#edit').checked,'editing defaults on');assert($('#save').disabled,'dirty on open');document.getElementById('ordinary').click();assert(!$('#media-dialog').open,'undesignated image editable');image.click();assert($('#media-dialog').open,'image click did not open editor');
 $('#media-alt').value='New description';$('#media-form').requestSubmit();await wait();assert(image.alt==='New description'&&!$('#save').disabled,'metadata preview or dirty missing');assert(image.className==='rounded-lg','developer class altered');
 $('#save').click();await wait();assert(sent.media[0].id==='Abcdef1234'&&sent.media[0].alt==='New description'&&sent.version==='v1'&&body.get('csrf')===host.dataset.csrf,'save payload');assert(!$('#save').disabled&&image.alt==='New description','failed save discarded draft');
 $('#cancel').click();$('#edit').click();assert(image.alt==='Joe'&&image.getAttribute('role')===null&&image.getAttribute('tabindex')===null,'cancel did not restore image');
 $('#edit').click();image.dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true}));assert($('#media-dialog').open,'keyboard inaccessible');$('#media-alt').value='Discard me';$('#media-cancel').click();await wait();assert(image.alt==='Joe'&&$('#save').disabled,'dialog cancel saved draft');
 window.fetch=async(url,options)=>{const edit=JSON.parse(options.body.get('edit'));return {status:200,json:async()=>({page:{...edit,version:'v2',changes:[]},media:{Abcdef1234:{type:'image',image:{file:'saved.png',name:'Saved',alt:'Saved description'}}},photos:{}})};};
 image.click();assert($('#media-dialog').open&&!$('#media-dialog').classList.contains('closing'),'media dialog did not reopen');$('#media-alt').value='Saved description';$('#media-form').requestSubmit();await wait();$('#save').click();await wait();
 assert(image.alt==='Saved description'&&image.src.endsWith('/media/saved.png')&&$('#save').disabled,'media save did not commit: '+image.alt+' '+image.src+' '+$('#status').textContent);
 image.click();$('#media-alt').value='Discard';$('#media-form').requestSubmit();await wait();$('#cancel').click();
 assert(image.alt==='Saved description'&&image.src.endsWith('/media/saved.png'),'media Cancel lost saved baseline');
 result.textContent='PASS';}catch(e){result.textContent='FAIL '+e.stack}document.body.append(result)},50);`
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		store := snips.NewStore()
		store.Page("/").Snips["Abcdef1234"] = snips.Entry{Type: "image", Image: &media.Photo{File: strings.Repeat("a", 64) + ".png", Name: "Portrait", Alt: "Joe", Width: 1, Height: 1}}
		body, ok := browserDecorate(w, r, decorate)([]byte(`<!doctype html><html><head><meta charset="utf-8"></head><body><main><img id="portrait" data-snip-media="Abcdef1234" src="/portrait.png" alt="Joe" class="rounded-lg"><img id="ordinary" src="/ordinary.png" alt="Not editable"></main></body></html>`), store, "v1")
		if !ok {
			return
		}
		sum := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(append(body, []byte("<script>"+checks+"</script>")...))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		request := httptest.NewRequest("GET", "https://example.test/", nil)
		request.AddCookie(cookie)
		h.ServePublic(w, request, http.NotFoundHandler())
	}))
	defer server.Close()
	ctx, stop := context.WithTimeout(context.Background(), 20*time.Second)
	defer stop()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--force-prefers-reduced-motion", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", "--virtual-time-budget=4000", server.URL)
	cmd.WaitDelay = 2 * time.Second
	output, err := cmd.CombinedOutput()
	text := string(output)
	if !strings.Contains(text, `<output id="photo-result">PASS</output>`) {
		if at := strings.Index(text, `<output id="photo-result">`); at >= 0 {
			tail := text[at:]
			if end := strings.Index(tail, "</output>"); end >= 0 {
				t.Fatal(tail[:end+9])
			}
		}
		t.Fatalf("browser assertions missing: %v", err)
	}
}
