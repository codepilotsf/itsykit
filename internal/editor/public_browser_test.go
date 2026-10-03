package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flea/internal/snips"
	"flea/internal/theme"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Optional real-engine smoke check, using a disposable browser profile and only
// synthetic authenticated HTML. No external mail, live credentials or login bypass.
func TestPublicPreviewBrowser(t *testing.T) {
	browser := os.Getenv("FLEA_TEST_BROWSER")
	if browser == "" {
		t.Skip("set FLEA_TEST_BROWSER to a Chromium executable")
	}
	checks := `setTimeout(async () => {
 const result = document.createElement('output'); result.id = 'browser-result';
 const assert = (ok, message) => { if (!ok) throw new Error(message); };
 try {
  const root=document.getElementById('flea-owner-root'), ui=root.shadowRoot;
  const article=document.querySelector('article'), leaf=document.querySelector('h2');
  const originalMatchMedia=window.matchMedia;window.matchMedia=query=>query==='(prefers-reduced-motion: reduce)'?{matches:false}:originalMatchMedia(query);
  assert(ui.querySelector('#edit').checked,'default editing mode');ui.querySelector('#edit').click();
  if(!matchMedia('(prefers-reduced-motion: reduce)').matches){const exiting=ui.querySelector('.collection-outlines-exit');assert(exiting?.inert&&!exiting.querySelector('button'),'outline exit must be inert and control-free');assert(exiting.getAnimations().some(a=>a.effect.getTiming().duration===160),'missing outline fade out');}
  assert(ui.querySelector('.editing-switch').textContent.trim()==='Edit'&&!ui.querySelector('nav strong,.divider'),'toolbar labels');
  assert(getComputedStyle(ui.querySelector('#edit')).backgroundColor==='rgb(71, 85, 105)','off switch must be dark');
  assert(getComputedStyle(ui.querySelector('#theme')).height==='30px'&&getComputedStyle(ui.querySelector('#logout')).borderRadius==='999px','compact pill buttons');
  const before=article.innerHTML, leafBefore=leaf.innerHTML;
  article.style.outline='3px solid purple';
  const originalOutline=article.style.outline;
  await new Promise(resolve=>setTimeout(resolve,300));
  assert(ui.querySelector('#format').inert && ui.querySelector('#details').disabled,'browse state');
  assert(ui.querySelector('.edit-row').getBoundingClientRect().height===0,'browse row still occupies space');
  const browseHeight=root.getBoundingClientRect().height;
  assert(root.getBoundingClientRect().bottom <= document.querySelector('header').getBoundingClientRect().top,'header overlap');
  ui.querySelector('#edit').click();
  assert(!ui.querySelector('.collection-outlines-exit'),'rapid toggle retained stale outlines');
  if(!matchMedia('(prefers-reduced-motion: reduce)').matches)assert(ui.querySelector('#collection-overlays').getAnimations().some(a=>a.effect.getTiming().duration===160),'missing outline fade in');
  window.matchMedia=originalMatchMedia;
  assert(!ui.querySelector('#format').inert&&root.getBoundingClientRect().height>browseHeight,'Edit did not restore second row');
  assert(getComputedStyle(ui.querySelector('.edit-row')).transitionDuration==='0s','reduced motion still animates row');
  assert(getComputedStyle(ui.querySelector('#edit')).backgroundColor==='rgb(22, 163, 74)','on switch must be green');
  assert(!ui.querySelector('#status').textContent&&!ui.querySelector('#notice').textContent,'idle hints remain');
  assert(!ui.querySelector('#theme').hidden && !ui.querySelector('#cancel').hidden && !ui.querySelector('#logout').hidden,'edit mode topbar');
  await new Promise(resolve=>setTimeout(resolve,50));
  const outlineStyle=getComputedStyle(ui.querySelector('.editable-outline'));assert(outlineStyle.borderTopColor==='rgb(22, 163, 74)'&&outlineStyle.borderTopWidth==='2px'&&outlineStyle.borderRadius==='6px','editable outline styling');
  assert(!ui.querySelector('#format').hidden && ui.querySelector('[data-mark]').disabled,'inactive formatting');
  assert(article.isContentEditable && leaf.isContentEditable,'region activation');
  assert(article.querySelector('span.text-brand-one'),'styled span lost entering edit');
  assert(document.querySelector('nav span').getAttribute('contenteditable')===null,'navigation editable');
  assert(document.querySelector('#plain').getAttribute('contenteditable')===null,'unmarked text editable');
  assert(document.querySelector('#mixed').getAttribute('contenteditable')===null,'mixed template region editable');
  assert(ui.querySelector('#status').textContent!=='Unsaved preview','dirty on entry');
  article.focus();
  assert(getComputedStyle(article).outlineStyle==='none','rich region editing outline');
  assert(getComputedStyle(leaf).outlineStyle==='none','leaf region editing outline');
  const selection=getSelection(), range=document.createRange();range.selectNodeContents(article.querySelector('p'));selection.removeAllRanges();selection.addRange(range);
  assert(!ui.querySelector('[data-mark]').disabled,'formatting inactive on selection');
  ui.querySelector('[data-mark="B"]').click();
  assert(article.querySelector('b,strong'),'bold not applied');
  assert(article.querySelector('span.text-brand-one'),'styled span lost applying formatting');
  assert(!ui.querySelector('#save').disabled,'changed content cannot save');
  window.confirm=()=>false;ui.querySelector('#edit').click();assert(ui.querySelector('#edit').checked&&article.isContentEditable&&!ui.querySelector('#save').disabled,'refused mode switch lost draft');
  assert(ui.querySelectorAll('[data-list]').length===2 && ui.querySelectorAll('[data-align]').length===3,'missing formatting controls');
  assert(!ui.querySelector('[data-command="undo"],[data-command="redo"]'),'history buttons present');
  assert(ui.querySelector('#details').tagName==='BUTTON' && !ui.textContent.includes('Preview only'),'preview UI retained');
  const selectText = node => { article.focus(); const r=document.createRange();r.selectNodeContents(node);selection.removeAllRanges();selection.addRange(r); };
  selectText(article.querySelector('p'));
  ui.querySelector('#heading').value='H1';ui.querySelector('#heading').dispatchEvent(new Event('change'));
  assert(article.querySelector('h1'),'H1 not applied');
  ui.querySelector('[data-align="center"]').click();
  assert(article.querySelector('h1').style.textAlign==='center','alignment not applied');
  selectText(article.querySelector('h1'));ui.querySelector('#link').click();
  assert(ui.querySelector('#link-dialog').open && ui.querySelector('#link-url').autocomplete==='off','link dialog');
  ui.querySelector('#link-url').value='example.com';ui.querySelector('#link-new-tab').checked=true;ui.querySelector('#link-form').dispatchEvent(new Event('submit',{cancelable:true}));
  await new Promise(resolve=>setTimeout(resolve,350));
  assert(article.querySelector('a')?.getAttribute('href')==='https://example.com' && article.querySelector('a').target==='_blank','new-tab link: '+article.innerHTML);
  selectText(article.querySelector('a'));ui.querySelector('#link').click();
  assert(ui.querySelector('#link-url').value==='https://example.com' && !ui.querySelector('#link-remove').hidden,'current link not editable');
  ui.querySelector('#link-new-tab').checked=false;ui.querySelector('#link-form').dispatchEvent(new Event('submit',{cancelable:true}));
  await new Promise(resolve=>setTimeout(resolve,350));
  assert(!article.querySelector('a').hasAttribute('target'),'new-tab toggle not cleared');
  ui.querySelector('#details').click();ui.querySelector('#page-title').value='Changed title';ui.querySelector('#page-details form').dispatchEvent(new Event('submit',{cancelable:true}));
  await new Promise(resolve=>setTimeout(resolve,350));
  let sent;
  window.fetch=async(url,options)=>{sent=JSON.parse(options.body.get('edit'));return {status:409};};
  ui.querySelector('#save').click();await new Promise(resolve=>setTimeout(resolve,30));
  assert(sent.title==='Changed title' && sent.changes.length===1 && sent.changes[0].content.includes('<h1'),'save not atomic');
  assert(sent.changes[0].content.includes('class="text-brand-one"'),'styled span lost in save payload');
  assert(ui.querySelector('#status').textContent.includes('draft is still here') && article.isContentEditable,'failed save lost draft: '+ui.querySelector('#status').textContent+' editable='+article.getAttribute('contenteditable'));
  leaf.focus();const leafRange=document.createRange();leafRange.selectNodeContents(leaf);selection.removeAllRanges();selection.addRange(leafRange);
  assert(!ui.querySelector('[data-mark]').disabled && ui.querySelector('#heading').disabled,'inline formatting controls');
  ui.querySelector('[data-mark="I"]').click();assert(leaf.querySelector('i,em'),'heading italic not applied');
  assert(!leaf.querySelector('p,div,h1,h2,ul,ol'),'nested block inside heading');
  ui.querySelector('#link').click();ui.querySelector('#link-url').value='example.org';ui.querySelector('#link-form').dispatchEvent(new Event('submit',{cancelable:true}));
  await new Promise(resolve=>setTimeout(resolve,350));assert(leaf.querySelector('a')?.getAttribute('href')==='https://example.org','inline link failed');
  ui.querySelector('#save').click();await new Promise(resolve=>setTimeout(resolve,30));
  const leafEdit=sent.changes.find(c=>c.id==='b7Km2pQ9xB');assert(leafEdit?.content.includes('<a') && !/<(p|div|h[1-6])\b/.test(leafEdit.content),'inline save contained blocks');
  window.confirm=()=>true;ui.querySelector('#cancel').click();assert(ui.querySelector('#edit').checked,'Cancel leaves editing on');ui.querySelector('#edit').click();
  assert(!ui.querySelector('.collection-outlines-exit'),'reduced motion must hide outlines immediately');
  assert(article.innerHTML===before && leaf.innerHTML===leafBefore,'cancel did not restore source');
  assert(article.style.outline===originalOutline && leaf.style.outline==='','cancel did not restore outline');
  assert(!article.hasAttribute('contenteditable') && !leaf.hasAttribute('contenteditable'),'editing attributes leaked');
  assert(!ui.querySelector('#format').hidden && ui.querySelector('#details').disabled && ui.querySelector('[data-mark]').disabled && !ui.querySelector('#logout').hidden,'cancel state');
  await new Promise(resolve=>setTimeout(resolve,300));
  assert(root.getBoundingClientRect().height===browseHeight,'toolbar did not collapse after leaving Edit');
  ui.querySelector('#edit').click();
  const fontSize=getComputedStyle(leaf).fontSize,originalLeaf=leaf;
  let saves=0,previousVersion=JSON.parse(root.dataset.page).version;
  window.fetch=async(url,options)=>{
    const edit=JSON.parse(options.body.get('edit'));assert(edit.version===previousVersion,'save used stale version');
    previousVersion='saved-'+(++saves);
    return {status:200,json:async()=>({page:{...edit,version:previousVersion,changes:edit.changes.map(c=>({...c,content:c.content.replace('draft','canonical')}))},media:{},photos:{}})};
  };
  leaf.textContent='First draft';leaf.dispatchEvent(new Event('input'));
  ui.querySelector('#save').click();await new Promise(resolve=>setTimeout(resolve,30));
  assert(leaf===originalLeaf&&leaf.textContent==='First canonical'&&getComputedStyle(leaf).fontSize===fontSize,'save replaced wrapper or changed typography');
  assert(ui.querySelector('#save').disabled&&ui.querySelector('#cancel').disabled&&leaf.isContentEditable,'saved state not clean and editable');
  leaf.textContent='Second draft';leaf.dispatchEvent(new Event('input'));
  ui.querySelector('#save').click();await new Promise(resolve=>setTimeout(resolve,30));
  assert(saves===2&&leaf.textContent==='Second canonical','second save failed');
  leaf.textContent='Discard me';leaf.dispatchEvent(new Event('input'));ui.querySelector('#cancel').click();
  assert(leaf.textContent==='Second canonical','Cancel restored pre-save content');
  ui.querySelector('#edit').click();
  assert(leaf.textContent==='Second canonical','leaving edit restored pre-save content');
  const oldBrand=document.documentElement.style.getPropertyValue('--brand-1');
  window.fetch=async(url,options)=>options?.method==='POST'?{status:409}:{ok:true,json:async()=>({version:'theme-version',fields:THEME_FIELDS,scales:THEME_SCALES})};
  ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));
  article.style.removeProperty('outline');
  assert(parseFloat(document.documentElement.style.getPropertyValue('--flea-toolbar-height'))>0,'public toolbar height token missing');
  assert(!ui.querySelector('#theme-panel').hidden && ui.querySelector('#theme-save').disabled,'theme initial state');
  await new Promise(resolve=>setTimeout(resolve,300));
  assert(ui.querySelector('#format').inert&&ui.querySelector('.edit-row').getBoundingClientRect().height===0,'Theme retained edit row');
  assert(Math.abs(ui.querySelector('#theme-panel').getBoundingClientRect().top-root.getBoundingClientRect().bottom)<1,'Theme did not track collapsed toolbar');
  const closeButton=ui.querySelector('#theme-cancel').getBoundingClientRect(),closeIcon=ui.querySelector('#theme-cancel svg').getBoundingClientRect();assert(Math.abs(closeButton.left+closeButton.width/2-closeIcon.left-closeIcon.width/2)<1&&Math.abs(closeButton.top+closeButton.height/2-closeIcon.top-closeIcon.height/2)<1,'theme close icon not centered');
  assert(ui.querySelectorAll('#theme-fields>details[open]').length===0,'theme sections should start closed');
  assert(ui.querySelector('#edit').disabled && !ui.querySelector('#theme').hidden && !ui.querySelector('#theme-cancel').hidden && !ui.querySelector('#logout').hidden,'theme mode topbar');
  assert(ui.querySelector('#theme-save').closest('#theme-panel footer') && ui.querySelector('#theme-cancel').closest('#theme-panel header'),'theme actions not in panel');
  assert(ui.querySelector('#theme-title').textContent==='Theme' && !ui.textContent.includes('A little personality'),'theme labels');
  const typography=[...ui.querySelectorAll('#theme-fields details')].find(d=>d.querySelector('summary').textContent==='Typography');
  typography.querySelector('summary').click();assert(typography.open,'accordion opening');
  const pageColors=[...ui.querySelectorAll('#theme-fields details')].find(d=>d.querySelector('summary').textContent==='Page colors');
  const fontColor=pageColors.querySelector('[aria-label="Font Color"]');
  const background=pageColors.querySelector('[aria-label="Background Color"]');assert(background.value==='var(--lighter)','background default');
  background.value='var(--brand-1-light)';background.dispatchEvent(new Event('input'));assert(document.documentElement.style.getPropertyValue('--bg-color')==='var(--brand-1-light)','background preview');
  assert(fontColor.tagName==='SELECT' && fontColor.value==='var(--black)','font color default');
  assert([...fontColor.querySelectorAll('optgroup')].map(g=>g.label).join(',')==='Neutral colors,Brand colors','font color ordering');
  assert(![...fontColor.options].some(o=>o.value==='var(--error)') && !typography.querySelector('[aria-label="Font Color color picker"]'),'semantic/font color picker present');
  fontColor.value='var(--brand-1)';fontColor.dispatchEvent(new Event('input'));
  assert(document.documentElement.style.getPropertyValue('--font-color')==='var(--brand-1)','font color preview');
  assert(typography.querySelector('[aria-label="Body Font"]').tagName==='INPUT','font not a text field');
  assert(ui.querySelectorAll('#theme-fields>details[open]').length===1,'multiple sections open');
  const help=typography.querySelector('[aria-label="About Body Font"]');help.click();assert(help.getAttribute('aria-expanded')==='true','help bubble missing');
  const font=typography.querySelector('[aria-label="Body Font"]');font.focus();assert(getComputedStyle(font).outlineOffset==='-3px','focus ring can be clipped');
  const originalAppend=document.head.append;let fontPreview;
  document.head.append=function(node){if(node.tagName==='LINK'&&node.href.startsWith('https://fonts.googleapis.com/css2?')){fontPreview=node;return;}return originalAppend.call(this,node);};
  font.value='"Lora", serif';font.dispatchEvent(new Event('input'));await new Promise(resolve=>setTimeout(resolve,450));
  assert(fontPreview?.href.includes('family=Lora'),'Google font preview missing');
  document.head.append=originalAppend;
  typography.querySelector('summary').click();assert(!typography.open,'accordion closing');
  const base=ui.querySelector('[aria-label="Text large screen base size (px)"]');
  assert(base && !ui.querySelector('[aria-label="Text Base"]'),'scale controls missing');
  const oldText=document.documentElement.style.getPropertyValue('--text-base');
  const small=ui.querySelector('[aria-label="Text small screen base size (px)"]');small.value=10;small.dispatchEvent(new Event('input'));
  base.value=100;base.dispatchEvent(new Event('input'));
  const expected=document.createElement('div');expected.style.setProperty('--text-base',EXPECTED_CLAMP);
  assert(document.documentElement.style.getPropertyValue('--text-base')===expected.style.getPropertyValue('--text-base'),'Go/browser negative-intercept mismatch');
  assert(!ui.querySelector('#theme-save').disabled,'scale change cannot save');
  for(const fixture of SCALE_FIXTURES){
    for(const [key,value] of Object.entries(fixture.settings)){
      const suffix=key.endsWith('Base')?(fixture.kind==='text'?'base size (px)':'small spacing (px)'):key.endsWith('Ratio')?'scale ratio':'width (px)';
      const input=ui.querySelector('[aria-label="'+(fixture.kind==='text'?'Text':'Spacing')+' '+(key.endsWith('Viewport')?'viewport '+(key.startsWith('min')?'min':'max')+' (px)':(key.startsWith('min')?'small':'large')+' screen '+suffix)+'"]');
      input.value=value;input.dispatchEvent(new Event('input'));
    }
    for(const [name,value] of Object.entries(fixture.values))assert(document.documentElement.style.getPropertyValue(name)===value,'scale parity '+name+': '+document.documentElement.style.getPropertyValue(name)+' expected '+value);
  }
  const color=ui.querySelector('#theme-fields input[aria-label="Brand 1"]');color.value='#123456';color.dispatchEvent(new Event('input'));
  assert(document.documentElement.style.getPropertyValue('--brand-1')==='#123456' && !ui.querySelector('#theme-save').disabled,'private theme preview');
  ui.querySelector('#theme-save').click();await new Promise(resolve=>setTimeout(resolve,30));
  assert(ui.querySelector('#theme-message').textContent.includes('preview is still here'),'theme failure lost preview');
  assert(ui.querySelector('#edit').disabled,'failed theme save enabled conflicting edit mode');
  window.confirm=()=>false;ui.querySelector('#edit').click();assert(!ui.querySelector('#theme-panel').hidden,'mode switch lost draft');
  window.confirm=()=>false;document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,cancelable:true}));assert(!ui.querySelector('#theme-panel').hidden,'Escape discarded dirty preview');
  window.confirm=()=>true;document.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,cancelable:true}));assert(document.documentElement.style.getPropertyValue('--brand-1')===oldBrand,'theme cancel not restored');
  assert(document.documentElement.style.getPropertyValue('--text-base')===oldText,'scale cancel not restored');
  assert(document.documentElement.style.getPropertyValue('--font-color')==='','font color cancel not restored');
  assert(!ui.querySelector('#edit').checked&&!article.isContentEditable,'Theme did not restore browsing preference');
  ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));
  window.matchMedia=query=>query==='(prefers-reduced-motion: reduce)'?{matches:false}:originalMatchMedia(query);
  ui.querySelector('#theme-cancel').click();const closingPanel=ui.querySelector('#theme-panel');assert(!closingPanel.hidden&&closingPanel.inert,'theme hidden before exit animation');assert(closingPanel.getAnimations().some(a=>a.effect.getTiming().duration===220),'theme exit animation missing');
  ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));assert(!closingPanel.hidden&&!closingPanel.inert,'reopening retained stale closing state');
  ui.querySelector('#theme-cancel').click();closingPanel.getAnimations().forEach(a=>a.finish());await new Promise(resolve=>setTimeout(resolve,30));assert(closingPanel.hidden,'theme exit did not finish');window.matchMedia=originalMatchMedia;
  ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));
  ui.querySelector('#theme-cancel').click();
  ui.querySelector('#edit').click();ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));
  assert(ui.querySelector('#format').inert,'Theme opened from Edit retained interactive row');
  ui.querySelector('#theme-cancel').click();
  assert(ui.querySelector('#edit').checked&&!ui.querySelector('#format').inert,'Theme failed to restore Edit row');
  ui.querySelector('#theme').click();await new Promise(resolve=>setTimeout(resolve,30));
  const textGroup=[...ui.querySelectorAll('#theme-fields details')].find(d=>d.querySelector('summary').textContent==='Text sizes');
  [...ui.querySelectorAll('#theme-fields>details[open]')].forEach(d=>d.querySelector('summary').click());
  textGroup.querySelector('summary').click();textGroup.scrollIntoView({block:'start'});
  result.textContent='PASS';
 } catch(error) { result.textContent='FAIL: '+error.message; }
 document.body.append(result);
},100);`
	themeCSS, err := os.ReadFile("../../templates/site/css/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	fields, err := theme.Fields(string(themeCSS))
	if err != nil {
		t.Fatal(err)
	}
	fieldJSON, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	checks = strings.ReplaceAll(checks, "THEME_FIELDS", string(fieldJSON))
	scaleJSON, _ := json.Marshal(theme.Scales(string(themeCSS)))
	checks = strings.ReplaceAll(checks, "THEME_SCALES", string(scaleJSON))
	clampJSON, _ := json.Marshal(theme.Clamp(theme.Scale{MinViewport: 320, MaxViewport: 1600, MinBase: 10, MaxBase: 100, MinRatio: 1.25, MaxRatio: 1.414}, 0))
	checks = strings.ReplaceAll(checks, "EXPECTED_CLAMP", string(clampJSON))
	var fixtures []any
	for _, kind := range []string{"text", "space"} {
		for _, s := range []theme.Scale{
			{MinViewport: 320, MaxViewport: 1600, MinBase: 10, MaxBase: 100, MinRatio: 1.25, MaxRatio: 1.414},
			{MinViewport: 320, MaxViewport: 1600, MinBase: 100, MaxBase: 10, MinRatio: 1.5, MaxRatio: 1.667},
			{MinViewport: 360, MaxViewport: 1440, MinBase: 16, MaxBase: 20, MinRatio: 1.2, MaxRatio: 1.333},
		} {
			values, _ := theme.Generate(kind, s)
			fixtures = append(fixtures, map[string]any{"kind": kind, "settings": s, "values": values})
		}
	}
	fixtureJSON, _ := json.Marshal(fixtures)
	checks = strings.ReplaceAll(checks, "SCALE_FIXTURES", string(fixtureJSON))
	// Reuse exactly the production injection and CSP through an authenticated session.
	h, mail := fixture(t)
	cookie := login(t, h, mail)
	h.options.Public = func(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) {
		body := []byte(`<!doctype html><html><body><header>Original header</header><article><!--snip-start:a7Km2pQ9xB--><p><span class="text-brand-one">Editable text</span></p><!--snip-end:a7Km2pQ9xB--></article><h2><!--snip-start:b7Km2pQ9xB-->Leaf heading<!--snip-end:b7Km2pQ9xB--></h2><nav><span><!--snip-start:c7Km2pQ9xB-->Navigation<!--snip-end:c7Km2pQ9xB--></span></nav><section id="mixed">Fixed <!--snip-start:d7Km2pQ9xB-->Mixed<!--snip-end:d7Km2pQ9xB--></section><p id="plain">Not editable</p></body></html>`)
		css := regexp.MustCompile(`@import[^\n]+`).ReplaceAll(themeCSS, nil)
		body = bytes.Replace(body, []byte("<body>"), []byte("<style>"+string(css)+"body{margin:0;background:#fafafa;font-family:system-ui;color:#1e293b}body>header{padding:24px 40px;background:white;border-bottom:1px solid #e2e8f0}body>article,body>h2,body>nav,body>section,body>p{margin:32px 40px;max-width:620px}body>article{font-size:32px;font-weight:650}body>output{display:none}</style><body>"), 1)
		body, ok := browserDecorate(w, r, decorate)(body, snips.NewStore(), "test-version")
		if !ok {
			return
		}
		digest := sha256.Sum256([]byte(checks))
		w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+" 'sha256-"+base64.StdEncoding.EncodeToString(digest[:])+"'")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(append(body, []byte("<script>"+checks+"</script>")...))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated := httptest.NewRequest("GET", "https://example.test/", nil)
		authenticated.AddCookie(cookie)
		h.ServePublic(w, authenticated, http.NotFoundHandler())
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, browser, "--headless", "--disable-gpu", "--force-prefers-reduced-motion", "--disable-background-networking", "--disable-component-update", "--disable-background-mode", "--no-first-run", "--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", "--virtual-time-budget=4000", server.URL)
	cmd.WaitDelay = 2 * time.Second
	if screenshot := os.Getenv("FLEA_TEST_SCREENSHOT"); screenshot != "" {
		size := os.Getenv("FLEA_TEST_WINDOW_SIZE")
		if size == "" {
			size = "1280,800"
		}
		cmd.Args = append(cmd.Args, "--window-size="+size, "--screenshot="+screenshot)
	}
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), `<output id="browser-result">PASS</output>`) {
		result := regexp.MustCompile(`<output id="browser-result">[^<]*</output>`).Find(output)
		t.Fatalf("browser check: %v; result: %s", err, result)
	}
	if err != nil {
		// Some Chrome builds keep background processes alive after --dump-dom.
		// The in-page assertion result, not process shutdown, is the test verdict.
		t.Logf("Browser assertions passed; bounded browser shutdown: %v", err)
	}
}
