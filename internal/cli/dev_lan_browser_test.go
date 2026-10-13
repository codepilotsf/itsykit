package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"flea/internal/devaccess"
	"flea/internal/site"
)

// Real HTTP LAN origin, real login/session, real editor bootstrap and page save.
// The isolated test mailbox contains only synthetic test mail, never real codes.
func TestLANBrowserEditing(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER to a Chromium executable")
	}
	addresses, err := devLANAddresses()
	if err != nil {
		t.Skip(err)
	}
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	if _, _, err := site.AddAdmin(root, "owner@example.test"); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	mail := make(previewInbox, 1)
	admin := devAdmin(root, "LAN browser", public, nil, mail)
	listener, err := net.Listen("tcp4", net.JoinHostPort(addresses[0], "0"))
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	result := make(chan string, 1)
	runner := `<!doctype html><html><body><script>
(async()=>{
 const assert=(v,m)=>{if(!v)throw Error(m)};
 const wait=async fn=>{for(let n=0;n<200;n++){if(fn())return;await new Promise(r=>setTimeout(r,50));}throw Error('editor wait timed out')};
 const post=(path,data)=>fetch(path,{method:'POST',body:new URLSearchParams(data)});
 try {
  assert(!isSecureContext,'test must use ordinary LAN HTTP, not localhost');
  crypto.getRandomValues(new Uint8Array(16));
  const login=new DOMParser().parseFromString(await (await fetch('/admin')).text(),'text/html');
  const csrf=login.querySelector('[name=csrf]').value;
  assert((await post('/admin/code',{csrf,email:'owner@example.test'})).ok,'request code');
  const code=await (await fetch('/_test/mail')).text();
  assert(/^\d{6}$/.test(code),'mailbox code');
  assert((await post('/admin/verify',{csrf,code})).ok,'verify code');
  location.href='/';
 }catch(e){await fetch('/_test/result',{method:'POST',body:String(e.stack||e)})}
})();</script></body></html>`
	// A hash-authorized assertion script is injected only by this isolated fixture.
	// Keep the real frame-ancestors policy: owner pages must not run in an iframe.
	checks := `(async()=>{
 const assert=(v,m)=>{if(!v)throw Error(m)};
 const wait=async fn=>{for(let n=0;n<200;n++){if(fn())return;await new Promise(r=>setTimeout(r,50));}throw Error('editor wait timed out')};
 try {
  const doc=document,win=window,host=doc.getElementById('flea-owner-root'),ui=host.shadowRoot;
  await wait(()=>ui.querySelector('#edit')?.disabled===false);
  await wait(()=>doc.querySelector('[contenteditable=true]'));
  const field=doc.querySelector('[contenteditable=true]');
  field.innerHTML='Saved in a LAN browser';field.dispatchEvent(new win.Event('input',{bubbles:true}));
  const previousVersion=host.dataset.version;
  await wait(()=>!ui.querySelector('#save').disabled);ui.querySelector('#save').click();
  await wait(()=>host.dataset.version!==previousVersion&&!ui.querySelector('#edit').disabled&&ui.querySelector('#save').disabled&&ui.querySelector('#cancel').disabled);
  assert((await (await fetch('/',{credentials:'omit'})).text()).includes('Saved in a LAN browser'),'save did not publish');
  assert(field.isConnected,'save unexpectedly replaced the page');
  assert(!ui.querySelector('#theme')&&!ui.querySelector('#theme-panel'),'removed Theme UI is present');
  await fetch('/_test/result',{method:'POST',body:'PASS'});
 }catch(e){await fetch('/_test/result',{method:'POST',body:String(e.stack||e)})}
})();`
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_test/run":
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, runner)
		case "/_test/mail":
			select {
			case m := <-mail:
				_, _ = io.WriteString(w, regexp.MustCompile(`\b[0-9]{6}\b`).FindString(m.Text))
			case <-time.After(5 * time.Second):
				http.Error(w, "no test mail", 500)
			}
		case "/_test/result":
			body, _ := io.ReadAll(io.LimitReader(r.Body, 8192))
			select {
			case result <- string(body):
			default:
			}
			w.WriteHeader(204)
		default:
			if r.URL.Path != "/" {
				admin.ServeHTTP(w, r)
				return
			}
			recorder := httptest.NewRecorder()
			admin.ServeHTTP(recorder, r)
			body := recorder.Body.Bytes()
			if strings.Contains(string(body), `id="flea-owner-root"`) {
				digest := sha256.Sum256([]byte(checks))
				recorder.Header().Set("Content-Security-Policy", recorder.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(digest[:])+"'")
				body = append(body, []byte("<script>"+checks+"</script>")...)
				recorder.Header().Del("Content-Length")
			}
			for name, values := range recorder.Header() {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(body)
		}
	})
	server := &http.Server{Handler: devaccess.WrapLAN(handler, addresses, []string{"asahi", "asahi.local"}, port), ReadHeaderTimeout: 5 * time.Second}
	defer server.Close()
	go server.Serve(listener)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--no-proxy-server", fmt.Sprintf("--host-resolver-rules=MAP asahi %s, MAP unpkg.com ~NOTFOUND, MAP fonts.googleapis.com ~NOTFOUND, MAP fonts.gstatic.com ~NOTFOUND", addresses[0]), "--disable-background-mode", "--disable-background-networking", "--disable-component-update", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), fmt.Sprintf("http://asahi:%d/_test/run", port))
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
	case got := <-result:
		cancel()
		<-done
		if got != "PASS" {
			t.Fatal(got)
		}
	case err := <-done:
		t.Fatalf("browser ended without result: %v\n%s", err, output.String())
	}
}
