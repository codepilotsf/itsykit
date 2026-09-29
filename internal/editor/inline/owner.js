/* Authenticated inline editing. Server validation is the security boundary. */
(() => {
  'use strict';
  const host = document.getElementById('flea-owner-root');
  if (!host) return;
  const shared = new Set(JSON.parse(host.dataset.shared));
  const page = JSON.parse(host.dataset.page);
  const shadow = host.shadowRoot;
  if(!shadow)throw Error('Missing declarative toolbar shell.');
  const $ = selector => shadow.querySelector(selector);
  const all = selector => shadow.querySelectorAll(selector);
  $('#identity').textContent = host.dataset.email;
  for (const button of all('[data-icon]')) button.title = button.getAttribute('aria-label');
  const blocks = 'p,h1,h2,h3,h4,h5,h6,li,blockquote,pre';
  const safeLink = value => {
    if (!value || /[\\\u0000-\u0020\u007f]/.test(value) || value.startsWith('//')) return false;
    try { return ['http:','https:','mailto:','tel:'].includes(new URL(value,location.href).protocol); } catch { return false; }
  };
  function normalizeLink(value) {
    const text=value.trim();
    if (/^[^\s/@:]+@[^\s/@:]+\.[^\s/@:]+$/.test(text)) return 'mailto:'+text;
    if (/^\+?[\d ()-]{7,}$/.test(text)) return 'tel:'+text.replace(/[ ()-]/g,'');
    if (/^[^\s/@:]+(?:\.[^\s/@:]+)+(?:[/:?#]|$)/.test(text)) return 'https://'+text;
    return text;
  }
  DOMPurify.addHook('uponSanitizeAttribute',(node,attr)=>{
    if (attr.attrName==='href') attr.keepAttr=node.nodeName==='A' && safeLink(attr.attrValue);
    if (attr.attrName==='title') attr.keepAttr=node.nodeName==='A';
    if (attr.attrName==='target') attr.keepAttr=node.nodeName==='A' && attr.attrValue==='_blank';
    if (attr.attrName==='rel') attr.keepAttr=false;
    if (attr.attrName==='style') attr.keepAttr=node.matches(blocks) && /^\s*text-align\s*:\s*(left|center|right)\s*;?\s*$/i.test(attr.attrValue);
  });
  function fragment(html) {
    const result=DOMPurify.sanitize(html,{RETURN_DOM_FRAGMENT:true,ALLOWED_TAGS:['span','p','div','br','hr','b','strong','i','em','u','s','del','mark','small','sub','sup','a','h1','h2','h3','h4','h5','h6','ul','ol','li','blockquote','pre','code'],ALLOWED_ATTR:['class','href','title','style','target'],ALLOW_DATA_ATTR:false,ALLOW_ARIA_ATTR:false});
    for (const a of result.querySelectorAll('a[target="_blank"]')) a.setAttribute('rel','noopener noreferrer');
    return result;
  }
  function cleanHTML(html) {
    const root=document.createElement('div'); root.append(fragment(html));
    for (const div of [...root.querySelectorAll('div')].reverse()) {
      if(div.querySelector('p,div,h1,h2,h3,h4,ul,ol')) div.replaceWith(...div.childNodes);
      else { const p=document.createElement('p');p.append(...div.childNodes);div.replaceWith(p); }
    }
    return root.innerHTML;
  }
  let editing=false,dirty=false,active=null,saving=false,expired=false,saveError='';
  let draft={title:page.title,description:page.description};
  const regions=[];
  function findRegions(root) {
  const walker=document.createTreeWalker(root,NodeFilter.SHOW_COMMENT);
  while(walker.nextNode()) {
    const start=walker.currentNode,match=/^snip-start:([A-Za-z0-9]{10}|[a-z][a-z0-9_]*\/[A-Za-z0-9]{10}\/[a-z][a-z0-9_]*)$/.exec(start.data);
    if(!match)continue;
    const parent=start.parentElement;
    if(!parent || parent.closest('nav,button,a,form,script,style,head') || parent.matches('body,html') || regions.some(r=>r.node===parent))continue;
    const nodes=[...parent.childNodes],first=nodes.indexOf(start),last=nodes.findIndex(n=>n.nodeType===8 && n.data==='snip-end:'+match[1]);
    const blank=n=>n.nodeType===3 && !n.textContent.trim();
    if(last<=first || !nodes.slice(0,first).every(blank) || !nodes.slice(last+1).every(blank) || parent.querySelector('img,video,form,input,button,iframe,table'))continue;
    regions.push({node:parent,id:match[1],html:parent.innerHTML,editable:parent.getAttribute('contenteditable'),tabindex:parent.getAttribute('tabindex'),editor:null});
  }
  }
  findRegions(document.body);
  const originalRegions=regions.slice();
  const rich=r=>regions.filter(other=>other.id===r.id).every(other=>['DIV','SECTION','ARTICLE','MAIN','ASIDE','HEADER','FOOTER'].includes(other.node.tagName));
  function inlineHTML(html){const root=document.createElement('div');root.append(fragment(html));for(const node of [...root.querySelectorAll('p,div,h1,h2,h3,h4,h5,h6,ul,ol,li,blockquote,pre')].reverse())node.replaceWith(...node.childNodes);return root.innerHTML;}
  // Use Squire's formatting engine in a temporary prose container, never inside
  // a heading/paragraph where its block normalization would create invalid HTML.
  function inlineEditor(root){
    let range=document.createRange();range.selectNodeContents(root);range.collapse(false);
    const selection=()=>{const s=getSelection();if(s.rangeCount && root.contains(s.anchorNode) && root.contains(s.focusNode))range=s.getRangeAt(0).cloneRange();return range;};
    const select=r=>{range=r.cloneRange();const s=getSelection();s.removeAllRanges();s.addRange(r);};
    const offsets=()=>{const r=selection(),before=document.createRange();before.selectNodeContents(root);before.setEnd(r.startContainer,r.startOffset);return [before.toString().length,r.toString().length];};
    function at(node,offset){const walker=document.createTreeWalker(node,NodeFilter.SHOW_TEXT);while(walker.nextNode()){const text=walker.currentNode;if(offset<=text.length)return [text,offset];offset-=text.length;}return [node,node.childNodes.length];}
    function locate(node,start,length){const r=document.createRange();r.setStart(...at(node,start));r.setEnd(...at(node,start+length));return r;}
    function apply(action){
      const [start,length]=offsets(),buffer=document.createElement('div');buffer.dataset.fleaInlineTool='';buffer.style.cssText='position:fixed;left:-10000px;top:0;opacity:0;pointer-events:none';document.body.append(buffer);
      const editor=new Squire(buffer,{blockTag:'P',addLinks:false,sanitizeToDOMFragment:fragment});
      try{editor.setHTML('<p>'+inlineHTML(root.innerHTML)+'</p>');editor.setSelection(locate(buffer,start,length));action(editor);root.innerHTML=inlineHTML(editor.getHTML());}
      finally{editor.destroy();buffer.remove();}
      root.focus();select(locate(root,start,length));
    }
    const paste=event=>{event.preventDefault();const html=event.clipboardData?.getData('text/html'),text=event.clipboardData?.getData('text/plain')||'';const box=document.createElement('div');if(html)box.innerHTML=inlineHTML(html);else box.textContent=text;const r=selection();r.deleteContents();const frag=document.createDocumentFragment();frag.append(...box.childNodes);const last=frag.lastChild;r.insertNode(frag);if(last){r.setStartAfter(last);r.collapse(true);select(r);}root.dispatchEvent(new Event('input'));};
    const enter=event=>{if(event.inputType==='insertParagraph'||event.inputType==='insertLineBreak'){event.preventDefault();const r=selection();r.deleteContents();const br=document.createElement('br');r.insertNode(br);r.setStartAfter(br);r.collapse(true);select(r);root.dispatchEvent(new Event('input'));}};
    const track=()=>{selection();controls();};
    root.contentEditable='true';root.addEventListener('paste',paste);root.addEventListener('beforeinput',enter);document.addEventListener('selectionchange',track);
    return {
      getHTML:()=>root.innerHTML,setHTML:html=>{root.innerHTML=inlineHTML(html);range=document.createRange();range.selectNodeContents(root);range.collapse(false);},focus:()=>{root.focus();select(range);},getSelection:selection,setSelection:select,
      hasFormat:tag=>{let node=selection().commonAncestorContainer;if(node.nodeType!==1)node=node.parentElement;return !!node.closest(tag);},
      changeFormat:(add,remove)=>apply(e=>e.changeFormat(add,remove)),removeLink:()=>apply(e=>e.removeLink()),makeLink:(url,attrs)=>apply(e=>e.makeLink(url,attrs)),
      addEventListener:(type,fn)=>{if(type==='input')root.addEventListener(type,fn);},
      destroy:()=>{root.removeEventListener('paste',paste);root.removeEventListener('beforeinput',enter);document.removeEventListener('selectionchange',track);}
    };
  }
  function content(r) {
    if(r.editor)return rich(r)?cleanHTML(r.editor.getHTML()):inlineHTML(r.editor.getHTML());
    const span=document.createElement('span');span.textContent=r.node.textContent;return span.innerHTML;
  }
  function controls() {
    const e=editing?active?.editor:null;
    $('#details').disabled=!editing || saving || expired;
    for(const c of all('[data-mark],[data-list],[data-align],#heading,#link'))c.disabled=!e || saving || expired;
    if(active && !rich(active))for(const c of all('[data-list],[data-align],#heading'))c.disabled=true;
    for(const b of all('[data-mark]'))b.setAttribute('aria-pressed',String(!!e?.hasFormat(b.dataset.mark)));
    for(const b of all('[data-list]'))b.setAttribute('aria-pressed',String(!!e?.hasFormat(b.dataset.list)));
    let node=e?.getSelection().startContainer; if(node?.nodeType!==1)node=node?.parentElement;
    const block=node?.closest(blocks);
    $('#heading').value=/^H[1-4]$/.test(block?.tagName)?block.tagName:'P';
    for(const b of all('[data-align]'))b.setAttribute('aria-pressed',String(!!e && b.dataset.align===(block?.style.textAlign||'left')));
    const linked=!!e?.hasFormat('A');$('#link').setAttribute('aria-label',linked?'Edit link':'Add link');$('#link').title=linked?'Edit link':'Add link';$('#link').setAttribute('aria-pressed',String(linked));
    $('#notice').textContent=editing&&active&&shared.has(active.id)?'Shared content: changes apply everywhere.':'';
    host.collectionUI?.updateState();
  }
  function changed(source) {
    if(!editing || saving)return;
    if(source) {
      const value=content(source);
      for(const r of regions.filter(r=>r.id===source.id && r!==source)) {
		if(content(r)===value)continue;
        if(r.editor)r.editor.setHTML(value);else r.node.textContent=source.node.textContent;
      }
    }
    dirty=draft.title!==page.title || draft.description!==page.description || regions.some(r=>r.node.isConnected && content(r)!==r.baseline) || multiDirty() || !!host.mediaSnips?.dirty();
    $('#save').disabled=!dirty || expired;
    $('#cancel').disabled=!dirty;
    $('#status').textContent=saveError || (dirty?'Unsaved changes':'');
    controls();
  }
  const updateToolbarHeight=()=>{const height=host.getBoundingClientRect().height;document.documentElement.style.scrollPaddingTop=height+12+'px';document.documentElement.style.setProperty('--flea-toolbar-height',height+'px');};
  updateToolbarHeight();new ResizeObserver(updateToolbarHeight).observe(host);
  function stop() {
    for(const r of regions) {
      r.editor?.destroy();r.editor=null;r.node.innerHTML=r.html;
      for(const [name,value] of [['contenteditable',r.editable],['tabindex',r.tabindex]])value===null?r.node.removeAttribute(name):r.node.setAttribute(name,value);
      r.node.removeEventListener('input',r.onInput);
      r.outline ? r.node.style.setProperty('outline',r.outline,r.outlinePriority) : r.node.style.removeProperty('outline');
      if(!r.node.getAttribute('style'))r.node.removeAttribute('style');
    }
    resetMultisnips();
    host.mediaSnips?.reset();
    draft={title:page.title,description:page.description};active=null;editing=dirty=false;saveError='';
    $('#cancel').disabled=$('#save').disabled=true;$('#edit').checked=false;$('#status').textContent='';controls();
  }
  const editPreference='flea-editing:'+host.dataset.email;
  let preferredEditing=true;
  try{preferredEditing=sessionStorage.getItem(editPreference)!=='off';}catch{}
  $('#edit').checked=preferredEditing;
  function start(){
    if(editing||saving||expired)return;
    editing=true;$('#edit').checked=true;
    for(const r of regions) startRegion(r);
    beginMultisnips();
    host.mediaSnips?.start();
    changed();
  }
  $('#edit').onclick=()=>{
    if(saving||expired){$('#edit').checked=editing;return;}
    if(editing){if(dirty&&!confirm('Discard your unsaved changes and turn editing off?')){$('#edit').checked=true;return;}stop();}else start();
    preferredEditing=editing;try{sessionStorage.setItem(editPreference,editing?'on':'off');}catch{}
  };
  function startRegion(r) {
      r.onInput=()=>changed(r);
      r.outline=r.node.style.getPropertyValue('outline');r.outlinePriority=r.node.style.getPropertyPriority('outline');
      r.node.style.setProperty('outline','none','important');
      {
        r.editor=rich(r)?new Squire(r.node,{blockTag:'P',addLinks:false,sanitizeToDOMFragment:fragment,undo:{documentSizeThreshold:-1,undoLimit:80}}):inlineEditor(r.node);
        r.editor.setHTML(rich(r)?cleanHTML(r.html):inlineHTML(r.html));r.editor.addEventListener('input',r.onInput);
        r.editor.addEventListener('pathChange',controls);r.editor.addEventListener('select',controls);
      }
      r.baseline=content(r);
  }
  for(const type of ['focusin','pointerdown'])document.addEventListener(type,event=>{
    if(!editing || event.target===host)return;
    if(event.target.closest?.('[data-flea-inline-tool]'))return;
    active=regions.find(r=>r.node.contains(event.target))||null;controls();
  });
  $('#format').addEventListener('mousedown',event=>{if(event.target.closest('button'))event.preventDefault();});
  function command(action) {if(!active?.editor || saving)return;saveError='';active.editor.focus();action(active.editor);changed(active);}
  // Collections retain their actual DOM nodes so Cancel restores source layout,
  // original item order and editor state without serializing the whole page.
  const groups=new Map();
  const libraryPhotos=JSON.parse(host.dataset.photos||'null')||{};
  const photoAttributes=new Map([...document.querySelectorAll('img[data-snip-photo]')].map(img=>[img,['src','alt','data-photo-id','tabindex','role','aria-label','style'].map(name=>[name,img.getAttribute(name)])]));
  function setPhoto(img,id){
    img.dataset.photoId=id;const photo=libraryPhotos[id];
    img.src=photo?(photo.preview||'/media/'+photo.file):host.dataset.photoPlaceholder;img.alt=photo?.alt||'';
  }
  const groupWalker=document.createTreeWalker(document.body,NodeFilter.SHOW_COMMENT);
  while(groupWalker.nextNode()) {
    const start=groupWalker.currentNode,match=/^multisnip-start:([a-z][a-z0-9_]*)$/.exec(start.data);
    if(!match)continue;
    const nodes=[];let end=start.nextSibling;
    while(end && !(end.nodeType===8 && end.data==='multisnip-end:'+match[1])){nodes.push(end);end=end.nextSibling;}
    const prototype=nodes.find(n=>n.nodeType===1 && n.tagName==='TEMPLATE' && n.dataset.multisnip===match[1]);
    if(!end || !prototype)continue;
    const items=nodes.filter(n=>n.nodeType===1 && n!==prototype).map(node=>({node,id:regions.find(r=>node.contains(r.node))?.id.split('/')[1]}));
    if(items.some(item=>!item.id))continue;
    groups.set(match[1],{start,end,prototype,nodes,items,original:items.slice(),baseline:''});
  }
  const label=value=>value.replaceAll('_',' ').replace(/\b\w/g,c=>c.toUpperCase());
  function multiDraft(){
    const result={};
    for(const [name,group] of groups)result[name]=group.items.map(item=>{
      const fields={};for(const r of regions)if(r.id.startsWith(name+'/'+item.id+'/'))fields[r.id.split('/')[2]]=r.editor?content(r):cleanHTML(r.node.innerHTML);
      for(const img of item.node.querySelectorAll('img[data-snip-photo]'))fields[img.dataset.snipPhoto]=img.dataset.photoId||'';
      return {id:item.id,fields};
    });
    return result;
  }
  function multiDirty(){const draft=multiDraft();return [...groups].some(([name,g])=>g.baseline!==JSON.stringify(draft[name]));}
  function beginMultisnips(){
    const draft=multiDraft();for(const [name,g] of groups)g.baseline=JSON.stringify(draft[name]);
    host.collectionUI?.start();
  }
  function resetMultisnips(){
    host.collectionUI?.stop();
    for(const group of groups.values()){
      let node=group.start.nextSibling;while(node && node!==group.end){const next=node.nextSibling;node.remove();node=next;}
      for(const original of group.nodes)group.end.before(original);
      group.items=group.original.slice();
    }
    regions.splice(0,regions.length,...originalRegions);
    for(const [img,attrs] of photoAttributes)for(const [name,value] of attrs)value===null?img.removeAttribute(name):img.setAttribute(name,value);
  }
  function newID(){const alphabet='abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';let id='';while(id.length<10){const bytes=crypto.getRandomValues(new Uint8Array(16));for(const b of bytes)if(b<248 && id.length<10)id+=alphabet[b%62];}return id;}
  function addItem(name){
    const group=groups.get(name);if(group.items.length>=200)return;
    let id;do{id=newID();}while([...groups.values()].some(g=>g.items.some(item=>item.id===id)));
    const node=group.prototype.content.firstElementChild.cloneNode(true);
    const comments=document.createTreeWalker(node,NodeFilter.SHOW_COMMENT);while(comments.nextNode())comments.currentNode.data=comments.currentNode.data.replace('/prototype/','/'+id+'/');
    group.end.before(node);const item={node,id};group.items.push(item);
    const before=regions.length;findRegions(node);
    for(const r of regions.slice(before)){startRegion(r);const text=document.createElement('span');text.textContent=label(r.id.split('/')[2]);r.editor.setHTML(text.innerHTML);}
    for(const img of node.querySelectorAll('img[data-snip-photo]'))setPhoto(img,'');
    changed();return item;
  }
  for(const b of all('[data-mark]'))b.onclick=()=>command(e=>{const mark={tag:b.dataset.mark};e.changeFormat(e.hasFormat(mark.tag)?null:mark,e.hasFormat(mark.tag)?mark:null);});
  for(const b of all('[data-list]'))b.onclick=()=>command(e=>{if(e.hasFormat(b.dataset.list))e.removeList();else e[b.dataset.list==='UL'?'makeUnorderedList':'makeOrderedList']();});
  for(const b of all('[data-align]'))b.onclick=()=>command(e=>e.setTextAlignment(b.dataset.align));
  $('#heading').onchange=()=>{
    const tag=$('#heading').value;
    command(e=>e.modifyBlocks(frag=>{
      for(const node of [...frag.querySelectorAll('p,div,h1,h2,h3,h4')].reverse()) {
        if(node.querySelector('p,div,h1,h2,h3,h4,ul,ol'))continue;
        const next=document.createElement(tag);next.style.textAlign=node.style.textAlign;next.append(...node.childNodes);node.replaceWith(next);
      }
      for(const li of frag.querySelectorAll('li')) {if(li.querySelector('p,h1,h2,h3,h4,ul,ol'))continue;const next=document.createElement(tag);next.append(...li.childNodes);li.append(next);}
      return frag;
    }));
  };
  const animations=new WeakMap();
  function open(dialog){dialog.showModal();if(!matchMedia('(prefers-reduced-motion: reduce)').matches)animations.set(dialog,dialog.animate([{opacity:0,transform:'translateY(-20px)'},{opacity:1,transform:'translateY(0)'}],{duration:280,easing:'ease-out'}));}
  async function close(dialog){
    if(!dialog.open || dialog.classList.contains('closing'))return;
    animations.get(dialog)?.cancel();dialog.classList.add('closing');
    if(!matchMedia('(prefers-reduced-motion: reduce)').matches)await dialog.animate([{opacity:1,transform:'translateY(0)'},{opacity:0,transform:'translateY(-20px)'}],{duration:240,easing:'ease-in',fill:'forwards'}).finished;
    dialog.close();dialog.getAnimations().forEach(a=>a.cancel());dialog.classList.remove('closing');
  }
  for(const dialog of all('dialog'))dialog.addEventListener('cancel',event=>{event.preventDefault();close(dialog);});
  $('#details').onclick=()=>{$('#page-title').value=draft.title;$('#page-description').value=draft.description;open($('#page-details'));};
  $('#details-cancel').onclick=()=>close($('#page-details'));
  $('#page-details form').onsubmit=event=>{event.preventDefault();draft={title:$('#page-title').value,description:$('#page-description').value};changed();close($('#page-details'));};
  let linkRange,linkRegion;
  $('#link').onclick=()=>{
    if(!active?.editor)return;linkRegion=active;linkRange=active.editor.getSelection().cloneRange();
    let n=linkRange.startContainer;if(n.nodeType!==1)n=n.parentElement;const a=n.closest('a');
    if(a)linkRange.selectNodeContents(a);
    $('#link-url').value=a?.getAttribute('href')||'';$('#link-new-tab').checked=a?.getAttribute('target')==='_blank';$('#link-remove').hidden=!a;$('#link-error').textContent='';open($('#link-dialog'));
  };
  $('#link-cancel').onclick=()=>close($('#link-dialog'));
  $('#link-remove').onclick=async()=>{if($('#link-dialog').classList.contains('closing'))return;await close($('#link-dialog'));active=linkRegion;command(e=>{e.setSelection(linkRange);e.removeLink();});};
  $('#link-form').onsubmit=async event=>{
    event.preventDefault();if($('#link-dialog').classList.contains('closing'))return;
    const url=normalizeLink($('#link-url').value),newTab=$('#link-new-tab').checked;
    if(!safeLink(url)){$('#link-error').textContent='Use a safe web, email, phone, or relative address.';return;}
    await close($('#link-dialog'));active=linkRegion;
    command(e=>{e.setSelection(linkRange);e.removeLink();e.makeLink(url,newTab?{target:'_blank',rel:'noopener noreferrer'}:{});});
  };
  $('#cancel').onclick=()=>{if(!saving && (!dirty || confirm('Discard your unsaved changes?'))){stop();start();}};
  $('#theme').onclick=()=>{
    if(saving || expired)return;
    if(editing){if(dirty && !confirm('Discard page changes and open Theme? Choose Cancel to save them first.'))return;stop();}
    host.dispatchEvent(new Event('theme-open'));
  };
  host.addEventListener('theme-closed',()=>{if(preferredEditing)start();});
  $('#save').onclick=async()=>{
    if(!dirty || saving || expired)return;
    const changes=new Map();for(const r of regions)if(!r.id.includes('/') && content(r)!==r.baseline)changes.set(r.id,{id:r.id,content:content(r)});
    const edit={version:page.version,route:page.route,title:draft.title,description:draft.description,changes:[...changes.values()],multisnips:multiDraft()};
    const body=host.mediaSnips?.dirty()||host.collectionUI?.hasUploads()?host.mediaSnips.body(edit):new URLSearchParams({csrf:host.dataset.csrf,edit:JSON.stringify(edit)});
    saving=true;$('#edit').disabled=$('#save').disabled=$('#cancel').disabled=$('#logout').disabled=$('#details').disabled=$('#theme').disabled=true;
    for(const r of regions)r.node.setAttribute('contenteditable','false');controls();$('#status').textContent='Saving…';
    try {
      const response=await fetch('/admin/page',{method:'POST',credentials:'same-origin',redirect:'error',headers:{'X-CSRF-Token':host.dataset.csrf},signal:AbortSignal.timeout(body instanceof FormData?180000:15000),body});
      if(response.status!==204) {
        if(response.status===403)expired=true;
        const validation=response.status===400?(await response.text()).trim().slice(0,500):'';
        throw new Error(expired?'Your session expired. Copy your draft before signing in again.':validation||'Could not save. Your draft is still here. Copy it before reloading; the page may have changed.');
      }
      dirty=false;location.reload();
    } catch(error) {
      saveError=error instanceof TypeError || error.name==='TimeoutError'?'Save could not be confirmed. Your draft is still here. Check the connection and copy it before reloading.':error.message;
      $('#status').textContent=saveError;
      saving=false;$('#edit').disabled=$('#save').disabled=$('#theme').disabled=expired;$('#cancel').disabled=$('#logout').disabled=$('#details').disabled=false;
      for(const r of regions)r.node.setAttribute('contenteditable',r.editor?'true':'plaintext-only');controls();
    }
  };
  document.addEventListener('click',event=>{if(editing && event.target.closest?.('a')?.closest('[contenteditable]'))event.preventDefault();});
  window.addEventListener('beforeunload',event=>{if(dirty){event.preventDefault();event.returnValue='';}});
  $('#logout').onclick=()=>{
    if(dirty && !confirm('Discard your unsaved changes and sign out?'))return;dirty=false;
    try{sessionStorage.removeItem(editPreference);}catch{}
    const form=document.createElement('form');form.method='post';form.action='/admin/logout';const csrf=document.createElement('input');csrf.type='hidden';csrf.name='csrf';csrf.value=host.dataset.csrf;form.append(csrf);document.body.append(form);form.submit();
  };
  host.ownerEditing={isEditing:()=>editing,canEdit:()=>editing&&!saving&&!expired,changed,open,close,editableNodes:()=>regions.filter(r=>r.node.isConnected).map(r=>r.node)};
  host.ownerCollections={groups,addItem,photos:libraryPhotos,setPhoto,clearActive:()=>{active=null;controls();}};
  shadow.querySelector('section').inert=false;
  queueMicrotask(()=>{if(preferredEditing)start();});
  if(host.dataset.revision){
    const key='flea-dev-edit:'+location.pathname;
    try {const saved=JSON.parse(sessionStorage.getItem(key)||'null');sessionStorage.removeItem(key);if(saved && Date.now()-saved.time<15000){requestAnimationFrame(()=>scrollTo(0,saved.scroll));}}catch{}
    let checking=false;
    setInterval(async()=>{
      if(checking || saving || expired || document.visibilityState!=='visible')return;checking=true;
      try {
        const response=await fetch('/_flea/reload',{cache:'no-store',credentials:'omit',redirect:'error',signal:AbortSignal.timeout(5000)});
        if(!response.ok)return;const revision=await response.text();
        if(!/^\d+$/.test(revision) || revision===host.dataset.revision)return;
        if(dirty){$('#status').textContent='Source changed. Your draft is preserved; cancel edits to reload.';return;}
        if(saving || expired || [...all('dialog')].some(dialog=>dialog.open))return;
        if(!host.dispatchEvent(new Event('dev-reload-check',{cancelable:true})))return;
        if(editing){try{sessionStorage.setItem(key,JSON.stringify({time:Date.now(),scroll:scrollY}));}catch{}}
        location.reload();
      }catch{}finally{checking=false;}
    },700);
  }
  let last=Date.now(),pending=false;
  for(const type of ['input','keydown','pointerdown','scroll'])document.addEventListener(type,async event=>{
    if(!event.isTrusted || document.visibilityState!=='visible' || pending || Date.now()-last<60000)return;last=Date.now();pending=true;
    try {const response=await fetch('/admin/activity',{method:'POST',credentials:'same-origin',redirect:'error',signal:AbortSignal.timeout(15000),body:new URLSearchParams({csrf:host.dataset.csrf})});if(!response.ok){expired=true;saveError='Session expired. Copy your draft before signing in again.';$('#status').textContent=saveError;$('#edit').disabled=$('#save').disabled=true;controls();}}
    catch {} finally {pending=false;}
  },{capture:true,passive:true});
})();
