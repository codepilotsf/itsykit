/* Developer-designated media regions share the page's Save/Cancel transaction. */
(() => {
  'use strict';
  const host=document.getElementById('flea-owner-root');
  if(!host?.ownerEditing)return;
  const ui=host.shadowRoot,owner=host.ownerEditing;
  const entries=JSON.parse(host.dataset.media||'null')||{};
  const drafts=new Map(),urls=new Set();
  const nodes=[...document.querySelectorAll('img[data-snip-media],video[data-snip-media]')].filter(node=>{
    const entry=entries[node.dataset.snipMedia];
    return entry && entry.type===(node.tagName==='IMG'?'image':'video') && !node.closest('button,form,script,template,[inert]');
  });
  const original=new Map(nodes.map(node=>[node,{
    attrs:['src','alt','poster','tabindex','role','aria-label'].map(name=>[name,node.getAttribute(name)]),
    cursor:node.style.getPropertyValue('cursor'),priority:node.style.getPropertyPriority('cursor'),
    outline:node.style.getPropertyValue('outline'),outlinePriority:node.style.getPropertyPriority('outline')
  }]));
  const dialog=document.createElement('dialog');dialog.id='media-dialog';dialog.setAttribute('aria-labelledby','media-title');
  dialog.innerHTML=`<form id="media-form">
    <h2 id="media-title">Edit image</h2>
    <p id="media-help">Changes are saved with the page.</p>
    <label for="media-name">Name</label><input id="media-name" maxlength="200">
    <div id="media-alt-field"><label for="media-alt">Alt text</label><input id="media-alt" maxlength="500"></div>
    <label for="media-file">Replace file</label><input id="media-file" type="file">
    <div id="media-poster-field" hidden>
      <label for="media-poster">Replace poster image</label><input id="media-poster" type="file" accept="image/jpeg,image/png">
      <label class="switch"><input id="media-clear-poster" type="checkbox" role="switch">Remove poster</label>
    </div>
    <p id="media-error" role="alert"></p>
    <div class="actions"><button id="media-cancel" type="button" class="text-button">Cancel</button><button type="submit">Done</button></div>
  </form>`;
  ui.append(dialog);
  const $=selector=>dialog.querySelector(selector);
  let selected='',sourceNode=null;
  const initial=id=>{const entry=entries[id],value=entry.image||entry.video;return {id,name:value.name||'',alt:entry.image?.alt||'',clear_poster:false};};
  function isChanged(draft){const base=initial(draft.id);return !!(draft.file||draft.poster||draft.clear_poster||draft.name!==base.name||draft.alt!==base.alt);}
  function objectURL(file){const url=URL.createObjectURL(file);urls.add(url);return url;}
  function restoreAttributes(node){for(const [name,value] of original.get(node).attrs)value===null?node.removeAttribute(name):node.setAttribute(name,value);}
  function preview(id,draft){
    for(const node of nodes.filter(node=>node.dataset.snipMedia===id)){
      const before=original.get(node).attrs;
      const value=name=>before.find(pair=>pair[0]===name)?.[1];
      node.setAttribute('src',draft?.url||value('src'));
      if(node.tagName==='IMG')node.alt=draft?.alt??value('alt')??'';
      else {
        const poster=draft?.clear_poster?null:draft?.posterURL||value('poster');
        poster===null?node.removeAttribute('poster'):node.setAttribute('poster',poster);
        node.load();
      }
    }
  }
  function show(node){
    if(!owner.canEdit()||dialog.open)return;
    selected=node.dataset.snipMedia;sourceNode=node;
    const entry=entries[selected],draft=drafts.get(selected)||initial(selected),image=entry.type==='image';
    $('#media-title').textContent=image?'Edit image':'Edit video';
    $('#media-help').textContent=image?'JPEG or PNG, up to 8 MiB. Changes are saved with the page.':'Optimized H.264 MP4, up to 25 MiB. Changes are saved with the page.';
    $('#media-name').value=draft.name;$('#media-alt').value=draft.alt;
    $('#media-alt-field').hidden=!image;$('#media-poster-field').hidden=image;
    $('#media-file').accept=image?'image/jpeg,image/png':'video/mp4';
    $('#media-file').value=$('#media-poster').value='';
    $('#media-clear-poster').checked=draft.clear_poster;
    $('#media-error').textContent='';
    owner.open(dialog);
  }
  function activate(event){
    const node=event.target.closest?.('img[data-snip-media],video[data-snip-media]');
    if(!nodes.includes(node)||!owner.isEditing())return;
    if(event.type==='keydown' && !['Enter',' '].includes(event.key))return;
    event.preventDefault();event.stopImmediatePropagation();show(node);
  }
  document.addEventListener('click',activate,true);
  document.addEventListener('keydown',activate,true);
  const cancel=()=>owner.close(dialog);
  $('#media-cancel').onclick=cancel;
  dialog.addEventListener('cancel',event=>{event.preventDefault();cancel();});
  $('#media-poster').onchange=()=>{if($('#media-poster').files.length)$('#media-clear-poster').checked=false;};
  $('#media-clear-poster').onchange=()=>{if($('#media-clear-poster').checked)$('#media-poster').value='';};
  $('#media-form').onsubmit=async event=>{
    event.preventDefault();if(!owner.canEdit()||dialog.classList.contains('closing'))return;
    const entry=entries[selected],image=entry.type==='image';
    const prior=drafts.get(selected),file=$('#media-file').files[0]||prior?.file;
    const clear=$('#media-clear-poster').checked;
    const poster=clear?null:$('#media-poster').files[0]||prior?.poster;
    if(file && (file.size>(image?8:25)*1048576 || !(image?/\.(jpe?g|png)$/i:/\.mp4$/i).test(file.name))){$('#media-error').textContent=image?'Choose a JPEG or PNG up to 8 MiB.':'Choose an optimized MP4 up to 25 MiB.';return;}
    if(poster && (poster.size>8*1048576 || !/\.(jpe?g|png)$/i.test(poster.name))){$('#media-error').textContent='Choose a JPEG or PNG poster up to 8 MiB.';return;}
    const draft={id:selected,name:$('#media-name').value.trim(),alt:image?$('#media-alt').value:'',clear_poster:!image&&clear,file,poster};
    const total=[...drafts.values()].filter(d=>d.id!==selected).concat(draft).reduce((n,d)=>n+(d.file?.size||0)+(d.poster?.size||0),0);
    if(total>38*1048576){$('#media-error').textContent='Save these changes before selecting more files (38 MiB per save).';return;}
    if(file)draft.url=file===prior?.file?prior.url:objectURL(file);
    if(poster)draft.posterURL=poster===prior?.poster?prior.posterURL:objectURL(poster);
    if(isChanged(draft))drafts.set(selected,draft);else drafts.delete(selected);
    preview(selected,drafts.get(selected));
    for(const url of [prior?.url,prior?.posterURL])if(url && url!==draft.url && url!==draft.posterURL){URL.revokeObjectURL(url);urls.delete(url);}
    owner.changed();await owner.close(dialog);sourceNode?.focus();
  };
  host.mediaSnips={
    commit(saved){
      Object.assign(entries,saved);
      for(const node of nodes){
        const entry=saved[node.dataset.snipMedia];if(!entry)continue;
        const values={src:'/media/'+(entry.image||entry.video).file};
        if(node.tagName==='IMG')values.alt=entry.image.alt||'';else values.poster=entry.poster?'/media/'+entry.poster.file:null;
        for(const pair of original.get(node).attrs)if(pair[0] in values){pair[1]=values[pair[0]];pair[1]===null?node.removeAttribute(pair[0]):node.setAttribute(pair[0],pair[1]);}
        if(node.tagName==='VIDEO')node.load();
      }
      drafts.clear();for(const url of urls)URL.revokeObjectURL(url);urls.clear();
    },
    dirty:()=>drafts.size>0,
    start(){for(const node of nodes){node.tabIndex=0;node.setAttribute('role','button');node.setAttribute('aria-label',node.tagName==='IMG'?'Edit image':'Edit video');node.style.cursor='pointer';node.style.setProperty('outline','none','important');if(node.tagName==='VIDEO')node.pause();}},
    reset(){
      drafts.clear();for(const node of nodes){restoreAttributes(node);const state=original.get(node);state.cursor?node.style.setProperty('cursor',state.cursor,state.priority):node.style.removeProperty('cursor');state.outline?node.style.setProperty('outline',state.outline,state.outlinePriority):node.style.removeProperty('outline');if(!node.getAttribute('style'))node.removeAttribute('style');if(node.tagName==='VIDEO')node.load();}
      for(const url of urls)URL.revokeObjectURL(url);urls.clear();
    },
    body(edit){
      edit.media=[...drafts.values()].map(({id,name,alt,clear_poster})=>({id,name,alt,clear_poster}));
      const body=new FormData();body.set('csrf',host.dataset.csrf);body.set('edit',JSON.stringify(edit));
      for(const draft of drafts.values()){if(draft.file)body.set('media-'+draft.id,draft.file);if(draft.poster)body.set('poster-'+draft.id,draft.poster);}
      host.collectionUI?.appendUploads(edit,body);
      body.set('edit',JSON.stringify(edit));
      return body;
    }
  };
  if(owner.isEditing())host.mediaSnips.start();
})();
