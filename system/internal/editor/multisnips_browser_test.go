package editor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"image"
	"image/png"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/templates"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMultisnipsBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER")
	}
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	photo, err := os.Create(filepath.Join(root, "site", "images", "qa-photo.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(photo, image.NewRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	if err = photo.Close(); err != nil {
		t.Fatal(err)
	}
	source := `<main><div id="team"><article multisnip="team"><img snip="photo" src="/images/qa-photo.png" alt="QA portrait" width="80" height="40"><h2 snip="name">Joe</h2><p snip="title">Founder</p><div snip="bio"><p>Hello</p></div></article></div><div id="services"><article multisnip="services"><h2 snip="name">Consulting</h2><p snip="description">Helpful advice.</p></article></div></main>`
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site", "main.layout.html"), []byte(`<!doctype html><html><head><title>Test</title><style>body{margin:0;background:#f8fafc;color:#1e293b;font:16px/1.5 system-ui}main{max-width:1040px;margin:auto;padding:48px 24px}#team{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:24px}article{box-sizing:border-box;background:white;border:1px solid #e2e8f0;border-radius:12px;padding:24px}h2{font-size:24px;margin:24px 0 12px}p{margin:12px 0}#services{margin-top:48px}output{display:none}@media(max-width:640px){#team{grid-template-columns:1fr}}</style></head><body>{slot}</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := site.LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	public := site.NewServer(p)
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	checks := `setTimeout(async()=>{const out=document.createElement('output');out.id='multi-result';const assert=(v,m)=>{if(!v)throw Error(m)};try{
 const host=document.getElementById('itsy-owner-root'),ui=host.shadowRoot,$=s=>ui.querySelector(s),team=document.getElementById('team'),services=document.getElementById('services'),original=(assert($('#edit').checked,'editing defaults on'),$('#edit').click(),team.innerHTML),originalServices=services.innerHTML;
 // Headless dump-dom does not reliably tick compositor frames under virtual time.
 // Drive layout/drag callbacks with virtual timers; reduced motion closes dialogs.
 window.requestAnimationFrame=fn=>setTimeout(()=>fn(performance.now()),16);window.cancelAnimationFrame=clearTimeout;
 const settle=()=>new Promise(r=>setTimeout(r,40)),key=(node,value)=>node.dispatchEvent(new KeyboardEvent('keydown',{key:value,bubbles:true,composed:true,cancelable:true}));
 const outlines=name=>[...ui.querySelectorAll('.collection-outline[data-collection="'+name+'"]')],action=(index,name)=>outlines('team')[index].querySelector('[data-action="'+name+'"]'),add=name=>$('.collection-add[data-collection="'+name+'"]');
 $('#edit').click();assert(!$('#multisnips'),'Collections toggle should be removed');assert($('#save').disabled,'dirty at entry');
 await settle();assert(getComputedStyle(outlines('team')[0]).zIndex==='1','collection controls must paint above field outlines');for(const name of ['drag','remove'])assert(getComputedStyle(action(0,name)).backgroundColor==='rgb(255, 255, 255)','collection control background must be opaque white');
 await settle();assert(!$('#multi-panel')&&!$('#multi-group'),'old sidebar present');assert(outlines('team').length===1&&outlines('services').length===1,'collections not outlined independently');assert(add('team')&&add('services'),'missing per-collection add tiles');assert(![...ui.querySelectorAll('button')].some(b=>/Move up|Move down/.test(b.textContent)),'old reorder buttons');assert(!team.querySelector('button'),'editor controls leaked into authored cards');assert($('#save').disabled,'arrangement toggle made draft dirty');
 const originalPhoto=team.querySelector('img').dataset.photoId;assert(originalPhoto,'seed photo ID missing');team.querySelector('img').click();const photoSelect=$('#collection-photo');assert($('#collection-photo-dialog').open&&$('#collection-photo-grid').children.length===2,'direct photo picker missing');assert($('#collection-photo-grid [aria-pressed=true]').dataset.photoId===originalPhoto,'selected thumbnail');$('#collection-photo-grid [data-photo-id=""]').click();$('#collection-photo-dialog form').requestSubmit();await new Promise(r=>setTimeout(r,350));assert(!$('#collection-photo-dialog').open,'photo dialog did not close');assert(team.querySelector('img').dataset.photoId===''&&team.querySelector('img').getAttribute('src'),'clear must show placeholder');assert([...ui.querySelectorAll('.collection-photo-hint')].some(b=>!b.hidden&&b.textContent.includes('Add photo')),'missing Add photo hint');assert(!$('#save').disabled,'photo clear not dirty');
 add('team').click();await settle();assert(team.querySelectorAll('article:not([data-flea-collection-add])').length===2,'new card absent');const name=team.querySelectorAll('h2')[1];name.textContent='Sam';name.dispatchEvent(new Event('input',{bubbles:true}));assert(name.contentEditable==='true','new card not inline editable');
 action(1,'drag').focus();key(action(1,'drag'),'ArrowUp');await settle();assert(team.querySelector('h2').textContent==='Sam','keyboard reorder failed');assert(services.querySelectorAll('article:not([data-flea-collection-add])').length===1&&services.textContent.includes('Consulting'),'other collection changed');key(action(0,'drag'),'End');await settle();assert(team.querySelectorAll('h2')[1].textContent==='Sam','End reorder failed');
 const dragHandle=action(1,'drag');dragHandle.hasPointerCapture=()=>false;const pointer=(type,x,y)=>dragHandle.dispatchEvent(new PointerEvent(type,{pointerId:42,pointerType:'touch',button:0,clientX:x,clientY:y,bubbles:true,cancelable:true}));let target=team.querySelector('article').getBoundingClientRect(),from=team.querySelectorAll('article')[1].getBoundingClientRect();pointer('pointerdown',from.left+20,from.top+20);pointer('pointermove',target.left+20,target.top+20);await settle();assert(team.querySelector('h2').textContent==='Sam','touch pointer reorder failed');assert(document.querySelector('[data-flea-inline-tool="drag-preview"]'),'missing floating drag preview');key(dragHandle,'Escape');await settle();assert(team.querySelectorAll('h2')[1].textContent==='Sam','Escape did not restore drag order');
 assert(document.querySelector('[data-flea-inline-tool="drag-preview"]')===null,'drag preview leaked after Escape');
 const originalMatchMedia=window.matchMedia;window.matchMedia=query=>query==='(prefers-reduced-motion: reduce)'?{matches:false}:originalMatchMedia(query);
 target=team.querySelector('article').getBoundingClientRect();from=team.querySelectorAll('article')[1].getBoundingClientRect();pointer('pointerdown',from.left+20,from.top+20);pointer('pointermove',target.left+20,target.top+20);await settle();const ghost=document.querySelector('[data-flea-inline-tool="drag-preview"]');assert(ghost?.inert&&getComputedStyle(ghost).boxShadow!=='none','drag preview has no lift');assert([...team.querySelectorAll('article')].some(node=>node.getAnimations().length),'neighbors did not animate');pointer('pointerup',target.left+20,target.top+20);assert(ghost.getAnimations().length,'drop did not animate');ghost.getAnimations().forEach(animation=>animation.finish());await settle();assert(!ghost.isConnected&&getComputedStyle(team.querySelector('article')).visibility!=='hidden','drop cleanup failed');window.matchMedia=originalMatchMedia;key(action(0,'drag'),'End');await settle();
 action(0,'remove').click();await settle();assert(team.querySelectorAll('article:not([data-flea-collection-add])').length===1&&team.textContent.includes('Sam'),'remove failed');
 window.confirm=()=>true;$('#cancel').click();assert($('#edit').checked,'Cancel must keep editing on');$('#edit').click();await settle();assert(team.innerHTML===original&&services.innerHTML===originalServices,'cancel not exact');assert(!outlines('team').length||$('#collection-overlays').hidden,'overlays persisted after cancel');
 $('#edit').click();await settle();action(0,'remove').click();await settle();assert(team.querySelectorAll('article:not([data-flea-collection-add])').length===0,'last item removal');assert(add('team')&&!add('team').hidden,'empty collection lost add tile');add('team').click();await settle();const nextName=team.querySelector('h2');nextName.textContent='New';nextName.dispatchEvent(new Event('input',{bubbles:true}));
 team.querySelector('img').click();const nextPhoto=$('#collection-photo');assert(nextPhoto.value==='','new card should have no photo');$('#collection-photo-grid [data-photo-id="'+originalPhoto+'"]').click();$('#collection-photo-dialog form').requestSubmit();await new Promise(r=>setTimeout(r,350));assert(team.querySelector('img').dataset.photoId===originalPhoto&&team.querySelector('img').alt==='QA portrait','selected photo preview');
 let payload;const fetchBefore=window.fetch;window.fetch=async(url,options)=>{if(url==='/admin/page'){payload=JSON.parse(options.body.get('edit'));return {status:409};}return fetchBefore(url,options)};
 $('#save').click();await settle();assert(payload.multisnips.team.length===1&&payload.multisnips.team[0].fields.name==='New','save payload');assert(payload.multisnips.team[0].fields.photo===originalPhoto,'save lost photo selection');assert(payload.multisnips.services.length===1&&payload.multisnips.services[0].fields.name==='Consulting','save changed unrelated collection');assert(!$('#cancel').hidden&&team.textContent.includes('New'),'stale save lost draft');
 team.querySelector('img').click();const upload=new DataTransfer();upload.items.add(new File(['photo bytes'], 'new.png',{type:'image/png'}));$('#collection-photo-file').files=upload.files;$('#collection-photo-file').dispatchEvent(new Event('change'));assert(!$('#collection-photo-selection').hidden&&$('#collection-photo-preview').src.startsWith('blob:'),'upload dialog preview');$('#collection-photo-alt').value='New portrait';$('#collection-photo-dialog form').requestSubmit();await settle();const uploadID=team.querySelector('img').dataset.photoId;assert(uploadID!==originalPhoto&&team.querySelector('img').src.startsWith('blob:')&&team.querySelector('img').alt==='New portrait','upload preview missing');
 team.querySelector('img').click();assert([...$('#collection-photo-grid').children].some(tile=>tile.dataset.photoId===uploadID),'draft absent from library');$('#collection-photo-cancel').click();await settle();
 let uploaded;window.fetch=async(url,options)=>{if(url==='/admin/page'){payload=JSON.parse(options.body.get('edit'));uploaded=options.body.get('media-'+uploadID);return {status:409};}return fetchBefore(url,options)};$('#save').click();await settle();assert(uploaded?.name==='new.png'&&payload.media.some(entry=>entry.id===uploadID&&entry.library&&entry.alt==='New portrait'),'upload missing from multipart save');assert(team.querySelector('img').src.startsWith('blob:'),'failed save discarded upload');$('#cancel').click();assert($('#edit').checked,'Cancel must keep editing on');$('#edit').click();await settle();assert(team.innerHTML===original,'upload cancel failed');$('#edit').click();team.querySelector('img').click();assert(![...$('#collection-photo-grid').children].some(tile=>tile.dataset.photoId===uploadID),'cancel retained draft library photo');await settle();assert($('#collection-photo-preview').getAttribute('src')&&$('#collection-photo-preview').naturalWidth>0,'reopened preview cleared or broken');
 $('#collection-photo-cancel').click();await settle();add('team').click();await settle();
 const savedName=team.querySelectorAll('h2')[1];savedName.textContent='Saved member';savedName.dispatchEvent(new Event('input',{bubbles:true}));
 window.fetch=async(url,options)=>{const edit=JSON.parse(options.body.get('edit'));return {status:200,json:async()=>({page:{...edit,version:'saved-version'},photos:JSON.parse(host.dataset.photos),media:{}})};};
 $('#save').click();await settle();assert($('#save').disabled&&team.querySelectorAll('h2').length===2,'collection save failed');
 action(0,'remove').click();savedName.textContent='Discard';savedName.dispatchEvent(new Event('input',{bubbles:true}));$('#cancel').click();await settle();
 assert(team.querySelectorAll('h2').length===2&&team.querySelectorAll('h2')[1].textContent==='Saved member','collection Cancel lost saved membership or text');
 out.textContent='PASS';}catch(e){out.textContent='FAIL '+e.message}document.body.append(out);},80);`
	if os.Getenv("FLEA_TEST_DRAG_SCREENSHOT") != "" {
		checks = strings.Replace(checks, "key(dragHandle,'Escape');", "out.textContent='PASS';document.body.append(out);return;key(dragHandle,'Escape');", 1)
	}
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		public.ServeOwner(w, r, func(body []byte, store *snips.Store, version string) ([]byte, bool) {
			body, ok := browserDecorate(w, r, decorate)(body, store, version)
			if !ok {
				return nil, false
			}
			hash := sha256.Sum256([]byte(checks))
			w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(hash[:])+"'")
			return append(body, []byte("<script>"+checks+"</script>")...), true
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/media/") {
			public.ServeHTTP(w, r)
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"--headless", "--disable-gpu", "--force-prefers-reduced-motion", "--no-first-run", "--no-default-browser-check", "--user-data-dir=" + t.TempDir(), "--dump-dom", "--virtual-time-budget=3000"}
	if screenshot := os.Getenv("FLEA_TEST_SCREENSHOT"); screenshot != "" {
		size := os.Getenv("FLEA_TEST_WINDOW_SIZE")
		if size == "" {
			size = "1280,1200"
		}
		args = append(args, "--window-size="+size, "--screenshot="+screenshot)
	}
	cmd := exec.CommandContext(ctx, browser, append(args, server.URL)...)
	cmd.WaitDelay = 2 * time.Second
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), `<output id="multi-result">PASS</output>`) {
		if at := strings.Index(string(output), `<output id="multi-result">`); at >= 0 {
			result := string(output)[at:]
			if end := strings.Index(result, "</output>"); end >= 0 {
				t.Fatal(result[:end+9])
			}
		}
		t.Fatalf("%v\n%s", err, output)
	}
}
