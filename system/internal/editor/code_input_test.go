package editor

import (
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
)

func TestCodeInputPolicyAndEmail(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Security-Policy", "default-src 'none'")
	body := string(appendCodeInput(w, []byte("<html></html>")))
	hash := sha256.Sum256([]byte(codeInputJS))
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "'sha256-"+base64.StdEncoding.EncodeToString(hash[:])+"'") || !strings.Contains(body, "<script>"+codeInputJS+"</script>") {
		t.Fatal("code enhancement must use exact hash-pinned bytes")
	}
	m := emailMessage("owner@example.test", "123456", "<unsafe>")
	if !strings.Contains(m.Text, "123456") || !strings.Contains(m.HTML, "font-size:36px;font-weight:700") || !strings.Contains(m.HTML, "&lt;unsafe&gt;") || strings.Contains(m.HTML, "<unsafe>") {
		t.Fatal("email needs prominent escaped HTML and plain text")
	}
}

func TestAdminReloadCSP(t *testing.T) {
	for _, mode := range []string{"login", "verify", "setup", "error"} {
		h, _ := fixture(t)
		h.options.Local = true
		h.options.ReloadScript = func() string { return "/* trusted reload */" }
		w := httptest.NewRecorder()
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		h.page(w, view{Mode: mode})
		policy := w.Header().Get("Content-Security-Policy")
		hash := sha256.Sum256([]byte(h.options.ReloadScript()))
		if !strings.Contains(w.Body.String(), h.options.ReloadScript()) || !strings.Contains(policy, base64.StdEncoding.EncodeToString(hash[:])) || !strings.Contains(policy, "connect-src 'self'") || strings.Count(policy, "script-src") != 1 {
			t.Fatalf("%s reload blocked by policy: %s", mode, policy)
		}
		if mode == "verify" {
			hash = sha256.Sum256([]byte(codeInputJS))
			if !strings.Contains(policy, base64.StdEncoding.EncodeToString(hash[:])) {
				t.Fatal("code input hash lost")
			}
		}
		h.options.Local = false
		w = httptest.NewRecorder()
		h.page(w, view{Mode: mode})
		if strings.Contains(w.Body.String(), h.options.ReloadScript()) {
			t.Fatal("production reload")
		}
	}
}

func TestCodeInputBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	checks := `const out=document.createElement('output');out.id='result';try {
 const boxes=[...document.querySelectorAll('.code-input input')], original=document.querySelector('[name=code]');
 const assert=(ok,msg)=>{if(!ok)throw Error(msg)};
 assert(boxes.length===6 && original.type==='hidden','six fields');
 assert(!boxes[0].closest('label'),'nested labels');
 let submissions=0;original.form.addEventListener('submit',event=>{event.preventDefault();submissions++;});
 const clip=new DataTransfer();clip.setData('text','123456');boxes[3].dispatchEvent(new ClipboardEvent('paste',{clipboardData:clip,bubbles:true,cancelable:true}));
 assert(original.value==='123456' && boxes.map(b=>b.value).join('')==='123456','paste');
 assert(submissions===1,'full code paste submits once');
 clip.setData('text','12');boxes[0].dispatchEvent(new ClipboardEvent('paste',{clipboardData:clip,bubbles:true,cancelable:true}));
 assert(submissions===1,'partial paste must not submit');
 boxes[0].value='654321';boxes[0].dispatchEvent(new Event('input'));assert(original.value==='654321','autofill');
 boxes[5].value='';boxes[5].dispatchEvent(new Event('input'));boxes[5].dispatchEvent(new KeyboardEvent('keydown',{key:'Backspace',cancelable:true}));assert(boxes[4].value==='' && document.activeElement===boxes[4],'backspace');
 boxes[4].value='9';boxes[4].dispatchEvent(new Event('input'));assert(document.activeElement===boxes[5],'advance');
 boxes[5].value='8';boxes[5].dispatchEvent(new Event('input'));assert(new FormData(original.form).get('code')==='654398' && original.form.checkValidity(),'submission');
 out.textContent='PASS';}catch(e){out.textContent='FAIL '+e.message}document.body.append(out);`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		body := appendCodeInput(w, []byte(`<!doctype html><html><body><form action="/admin/verify"><label>Code<input name="code" required></label></form></body></html>`))
		hash := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(hash[:])+"'")
		w.Write(append(body, []byte("<script>"+checks+"</script>")...))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", server.URL)
	cmd.WaitDelay = 2 * time.Second
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), `<output id="result">PASS</output>`) {
		t.Fatalf("%v\n%s", err, output)
	}
}
