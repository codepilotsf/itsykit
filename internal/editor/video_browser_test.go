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

	"itsy/internal/media"
	"itsy/internal/snips"
	"itsy/internal/video"
)

func TestVideosBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	videoBytes, e := os.ReadFile("../video/testdata/valid.mp4")
	if e != nil {
		t.Fatal(e)
	}
	checks := `setTimeout(async()=>{const result=document.createElement('output');result.id='video-result';const assert=(v,m)=>{if(!v)throw Error(m)},wait=()=>new Promise(r=>setTimeout(r,350));try{
 const host=document.getElementById('flea-owner-root'),ui=host.shadowRoot,$=s=>ui.querySelector(s),movie=document.getElementById('movie');let sent,body;
 window.fetch=async(url,options)=>{if(url!=='/admin/page')return {ok:true};body=options.body;sent=JSON.parse(body.get('edit'));assert(options.headers['X-CSRF-Token']===host.dataset.csrf,'CSRF header');return {ok:false,status:409};};window.confirm=()=>true;
 assert(!$('#videos')&&!$('#photos'),'library buttons remain');assert($('#edit').checked,'editing defaults on');movie.click();assert($('#media-dialog').open&&!$('#media-poster-field').hidden,'video editor missing');
 $('#media-name').value='New movie';$('#media-clear-poster').checked=true;const transfer=new DataTransfer(),raw=Uint8Array.from(atob('FIXTURE'),c=>c.charCodeAt(0));transfer.items.add(new File([raw],'movie.mp4',{type:'video/mp4'}));$('#media-file').files=transfer.files;$('#media-form').requestSubmit();await wait();assert(movie.getAttribute('src').startsWith('blob:')&&!movie.hasAttribute('poster'),'preview missing');assert(movie.hasAttribute('controls')&&movie.getAttribute('preload')==='metadata','developer playback attributes altered');
 $('#save').click();await wait();assert(sent.media[0].id==='Abcdef1234'&&sent.media[0].name==='New movie'&&sent.media[0].clear_poster&&body.get('media-Abcdef1234').size===raw.length&&body.get('csrf')===host.dataset.csrf,'upload payload');assert(!$('#save').disabled&&movie.getAttribute('src').startsWith('blob:'),'failed save lost draft');
 movie.click();assert($('#media-name').value==='New movie','reopening lost metadata');$('#media-cancel').click();await wait();$('#cancel').click();assert(movie.getAttribute('src')==='/video.mp4'&&movie.getAttribute('poster')==='/poster.png','cancel did not restore source/poster');
 result.textContent='PASS';}catch(e){result.textContent='FAIL '+e.stack}document.body.append(result)},50);`
	checks = strings.ReplaceAll(checks, "FIXTURE", base64.StdEncoding.EncodeToString(videoBytes))
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		store := snips.NewStore()
		store.Page("/").Snips["Abcdef1234"] = snips.Entry{Type: "video", Video: &video.Video{Metadata: video.Metadata{File: strings.Repeat("a", 64) + ".mp4", Size: 1024, Width: 32, Height: 24, Duration: 1}, Name: "Intro"}, Poster: &media.Photo{File: strings.Repeat("b", 64) + ".png", Width: 1, Height: 1}}
		body, ok := browserDecorate(w, r, decorate)([]byte(`<!doctype html><html><head><meta charset="utf-8"></head><body><main><video id="movie" data-snip-media="Abcdef1234" src="/video.mp4" poster="/poster.png" controls preload="metadata"></video></main></body></html>`), store, "v1")
		if !ok {
			return
		}
		sum := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(append(body, []byte("<script>"+checks+"</script>")...))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.Header().Set("Content-Type", "video/mp4")
			http.ServeContent(w, r, "video.mp4", time.Time{}, bytes.NewReader(videoBytes))
			return
		}
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
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", "--virtual-time-budget=2500", server.URL)
	cmd.WaitDelay = 2 * time.Second
	output, err := cmd.CombinedOutput()
	text := string(output)
	if !strings.Contains(text, `<output id="video-result">PASS</output>`) {
		if at := strings.Index(text, `<output id="video-result">`); at >= 0 {
			tail := text[at:]
			if end := strings.Index(tail, "</output>"); end >= 0 {
				t.Fatal(tail[:end+9])
			}
		}
		t.Fatalf("browser assertions missing: %v", err)
	}
}
