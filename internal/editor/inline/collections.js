/* Arrange the real collection nodes; editor chrome never enters saved content. */
(() => {
  'use strict';
  const host=document.getElementById('flea-owner-root');
  if(!host?.ownerCollections)return;
  const ui=host.shadowRoot,$=s=>ui.querySelector(s),owner=host.ownerEditing;
  const {groups,photos,setPhoto,addItem,clearActive}=host.ownerCollections;
  const uploads=new Map();
  const usedUploads=()=>[...uploads].filter(([id])=>[...groups.values()].some(group=>group.items.some(item=>[...item.node.querySelectorAll('img[data-snip-photo]')].some(image=>image.dataset.photoId===id))));
  const layer=$('#collection-overlays'),status=$('#collection-status');
  const outlines=new Map(),tiles=new Map(),photoHints=new Map(),photoState=new Map();
  let arranging=true,drag=null,frame=0,selectedPhoto=null;
  const regionOutlines=new Map();
  let outlineFade=null,exitLayer=null;
  function clearOutlineFade(){outlineFade?.cancel();outlineFade=null;exitLayer?.remove();exitLayer=null;}
  function fadeOutlinesOut(){
    const opacity=getComputedStyle(layer).opacity;clearOutlineFade();
    if(matchMedia('(prefers-reduced-motion: reduce)').matches)return;
    const outgoing=layer.cloneNode(false);outgoing.removeAttribute('id');outgoing.className='collection-outlines-exit';outgoing.hidden=false;outgoing.inert=true;outgoing.setAttribute('aria-hidden','true');
    for(const outline of layer.querySelectorAll('.editable-outline,.collection-outline'))outgoing.append(outline.cloneNode(false));
    if(!outgoing.children.length)return;
    ui.append(outgoing);exitLayer=outgoing;
    const animation=outgoing.animate([{opacity},{opacity:0}],{duration:160,easing:'ease-out',fill:'forwards'});
    outlineFade=animation;animation.finished.then(()=>{outgoing.remove();if(exitLayer===outgoing){exitLayer=null;outlineFade=null;}},()=>outgoing.remove());
  }
  const settling=new Set(),movement=new Map();
  function finishAnimations(){for(const cleanup of [...settling])cleanup();for(const animation of movement.values())animation.cancel();movement.clear();}
  const label=name=>name.replaceAll('_',' ').replace(/\b\w/g,c=>c.toUpperCase());
  const icon=name=>$('#collection-'+name).content.cloneNode(true);
  const announce=text=>{status.textContent=text;};
  function button(title,action,iconName){
    const node=document.createElement('button');node.type='button';node.title=title;node.setAttribute('aria-label',title);
    if(iconName)node.append(icon(iconName));
    if(action)node.onclick=()=>{if(owner.canEdit())action();};
    return node;
  }
  function place(control,node){
    const r=node.getBoundingClientRect();control.hidden=r.width===0||r.height===0;
    Object.assign(control.style,{left:r.left+'px',top:r.top+'px',width:r.width+'px',height:r.height+'px'});
  }
  function position(){
    frame=0;if(!owner.isEditing())return;
    layer.style.clipPath='inset('+Math.max(0,host.getBoundingClientRect().bottom)+'px 0 0)';
    for(const [item,outline] of outlines)place(outline,item.node);
    const editable=new Set([...owner.editableNodes(),...[...document.querySelectorAll('img[data-snip-photo],img[data-snip-media],video[data-snip-media]')].filter(node=>!node.closest('[data-flea-inline-tool]'))]);
    for(const [node,outline]of regionOutlines)if(!editable.has(node)){outline.remove();regionOutlines.delete(node);}
    for(const node of editable){let outline=regionOutlines.get(node);if(!outline){outline=document.createElement('div');outline.className='editable-outline';layer.append(outline);regionOutlines.set(node,outline);}place(outline,node);outline.classList.toggle('focused',node.contains(document.activeElement));}
    for(const {node,control} of tiles.values())place(control,node);
    for(const [image,hint] of photoHints){
      const r=image.getBoundingClientRect();hint.hidden=!!image.dataset.photoId||!image.isConnected||r.width===0;
      Object.assign(hint.style,{left:(r.left+r.width/2)+'px',top:(r.bottom-14)+'px',maxWidth:Math.max(r.width,120)+'px'});
    }
  }
  function schedule(){if(!frame && owner.isEditing())frame=requestAnimationFrame(position);}
  const observer=new ResizeObserver(schedule);
  window.addEventListener('resize',schedule);
  document.addEventListener('scroll',schedule,true);
  document.addEventListener('load',schedule,true);
  document.addEventListener('input',schedule,true);
  document.addEventListener('focusin',schedule,true);
  document.addEventListener('pointerover',event=>{
    const target=event.composedPath()[0];
    for(const [item,outline]of outlines)outline.classList.toggle('hover',item.node.contains(target)||outline.contains(target));
    for(const [node,outline]of regionOutlines)outline.classList.toggle('hover',node.contains(event.target));
  },true);
  const reveal=node=>{const r=node.getBoundingClientRect();if(r.top<host.getBoundingClientRect().bottom+24 || r.bottom>innerHeight)node.scrollIntoView({block:'nearest',inline:'nearest'});schedule();};
  function order(group){const before=tiles.get(group)?.node||group.end;for(const item of group.items)before.before(item.node);}
  function dragTargets(group){return new Map(group.items.map(item=>{const r=item.node.getBoundingClientRect();return [item,{left:r.left+scrollX,right:r.right+scrollX,top:r.top+scrollY,bottom:r.bottom+scrollY}];}));}
  function orderControls(group){
    const focused=ui.activeElement;
    for(const item of group.items){const outline=outlines.get(item);if(outline)layer.append(outline);}
    if(focused?.isConnected && ui.activeElement!==focused)focused.focus({preventScroll:true});
  }
  function move(group,item,to){
    const from=group.items.indexOf(item);to=Math.max(0,Math.min(group.items.length-1,to));
    if(from<0||from===to)return;
    const before=new Map(group.items.map(other=>[other,other.node.getBoundingClientRect()]));
    for(const animation of movement.values())animation.cancel();movement.clear();
    group.items.splice(from,1);group.items.splice(to,0,item);order(group);if(!drag)orderControls(group);owner.changed();position();
    if(drag)drag.targets=dragTargets(group);
    if(!matchMedia('(prefers-reduced-motion: reduce)').matches)for(const other of group.items){if(other===item&&drag)continue;const old=before.get(other),now=other.node.getBoundingClientRect();movement.set(other,other.node.animate([{translate:`${old.left-now.left}px ${old.top-now.top}px`},{translate:'0 0'}],{duration:180,easing:'ease-out'}));}
    announce('Item moved to position '+(to+1)+' of '+group.items.length+'.');
  }
  function finishDrag(cancel=false){
    if(!drag)return;
    const current=drag;drag=null;cancelAnimationFrame(current.frame);
    if(cancel){finishAnimations();current.group.items=current.before;order(current.group);owner.changed();announce('Reordering cancelled.');}
    current.outline.classList.remove('dragging');current.handle.setAttribute('aria-pressed','false');
    if(current.ghost){
      const node=current.item.node,r=node.getBoundingClientRect();
      let animation,cleaned=false;
      const cleanup=()=>{if(cleaned)return;cleaned=true;settling.delete(cleanup);animation?.cancel();current.ghost.remove();current.visibility?node.style.setProperty('visibility',current.visibility,current.visibilityPriority):node.style.removeProperty('visibility');if(!node.getAttribute('style'))node.removeAttribute('style');};
      if(cancel||matchMedia('(prefers-reduced-motion: reduce)').matches)cleanup();
      else{settling.add(cleanup);animation=current.ghost.animate([{transform:'scale(1.02)'},{transform:`translate(${r.left-parseFloat(current.ghost.style.left)}px,${r.top-parseFloat(current.ghost.style.top)}px) scale(1)`}],{duration:160,easing:'ease-out'});animation.finished.then(cleanup,cleanup);}
    }
    if(current.handle.hasPointerCapture?.(current.pointer))current.handle.releasePointerCapture(current.pointer);
    orderControls(current.group);
    schedule();
  }
  function trackDrag(){
    if(!drag)return;
    const {x,y,group,item}=drag;
    if(drag.moved){
      if(!drag.ghost){
        const node=item.node,r=node.getBoundingClientRect(),ghost=node.cloneNode(true);
        const source=[node,...node.querySelectorAll('*')],copies=[ghost,...ghost.querySelectorAll('*')];
        for(let i=0;i<source.length;i++){const style=getComputedStyle(source[i]);for(const prop of style)copies[i].style.setProperty(prop,style.getPropertyValue(prop));for(const attr of [...copies[i].attributes])if(['id','contenteditable','autofocus'].includes(attr.name)||attr.name.startsWith('on'))copies[i].removeAttribute(attr.name);}
        ghost.querySelectorAll('script,iframe,object,embed').forEach(n=>n.remove());
        ghost.inert=true;ghost.setAttribute('aria-hidden','true');ghost.dataset.fleaInlineTool='drag-preview';
        Object.assign(ghost.style,{position:'fixed',margin:'0',width:r.width+'px',height:r.height+'px',pointerEvents:'none',zIndex:'2147483646',transform:'scale(1.02)',boxShadow:'0 18px 45px #0f172a40',backgroundColor:getComputedStyle(node).backgroundColor==='rgba(0, 0, 0, 0)'?'white':getComputedStyle(node).backgroundColor,borderRadius:'8px',overflow:'hidden'});
        drag.offsetX=drag.startX-r.left;drag.offsetY=drag.startY-r.top;drag.visibility=node.style.getPropertyValue('visibility');drag.visibilityPriority=node.style.getPropertyPriority('visibility');node.style.setProperty('visibility','hidden','important');document.body.append(ghost);drag.ghost=ghost;
      }
      drag.ghost.style.left=(x-drag.offsetX)+'px';drag.ghost.style.top=(y-drag.offsetY)+'px';
      const edge=host.getBoundingClientRect().bottom+48;
      if(y<edge)window.scrollBy(0,-12);else if(y>innerHeight-48)window.scrollBy(0,12);
      // Hit-test final layout slots, not the animated cards passing underneath.
      const target=group.items.find(other=>{const r=drag.targets.get(other);return other!==item&&x+scrollX>=r.left&&x+scrollX<=r.right&&y+scrollY>=r.top&&y+scrollY<=r.bottom;});
      if(target)move(group,item,group.items.indexOf(target));
    }
    drag.frame=requestAnimationFrame(trackDrag);
  }
  function addOutline(name,group,item){
    const outline=document.createElement('div');outline.className='collection-outline';outline.dataset.collection=name;outline.dataset.item=item.id;
    const handle=button('Reorder item',null,'grip');handle.dataset.action='drag';handle.setAttribute('aria-pressed','false');
    handle.setAttribute('aria-description','Drag to reorder, or use arrow keys. Home and End move to the first or last position.');
    const remove=button('Remove item',()=>{
      const index=group.items.indexOf(item);group.items.splice(index,1);item.node.remove();clearActive();owner.changed();render();
      const next=group.items[Math.min(index,group.items.length-1)];
      (next?outlines.get(next)?.querySelector('button'):tiles.get(group)?.control)?.focus();
      announce('Item removed. Save changes to keep it, or Cancel to restore it.');
    },'remove');remove.dataset.action='remove';
    handle.onfocus=()=>reveal(item.node);remove.onfocus=()=>reveal(item.node);
    handle.onkeydown=event=>{
      if(!owner.canEdit())return;
      const from=group.items.indexOf(item),to={ArrowUp:from-1,ArrowLeft:from-1,ArrowDown:from+1,ArrowRight:from+1,Home:0,End:group.items.length-1}[event.key];
      if(to!==undefined){event.preventDefault();move(group,item,to);reveal(item.node);}
    };
    handle.onpointerdown=event=>{
      if(!owner.canEdit()||event.button!==0||drag)return;
      event.preventDefault();finishAnimations();handle.focus();
      drag={group,item,outline,handle,pointer:event.pointerId,before:group.items.slice(),targets:dragTargets(group),x:event.clientX,y:event.clientY,startX:event.clientX,startY:event.clientY,moved:false};
      if(event.isTrusted)handle.setPointerCapture?.(event.pointerId);handle.setAttribute('aria-pressed','true');outline.classList.add('dragging');trackDrag();
    };
    handle.onpointermove=event=>{if(drag?.handle!==handle||event.pointerId!==drag.pointer)return;drag.x=event.clientX;drag.y=event.clientY;drag.moved ||= Math.hypot(drag.x-drag.startX,drag.y-drag.startY)>4;};
    handle.onpointerup=event=>{if(event.pointerId===drag?.pointer)finishDrag();};handle.onpointercancel=event=>{if(event.pointerId===drag?.pointer)finishDrag(true);};
    handle.onlostpointercapture=()=>{if(drag?.handle===handle)finishDrag(true);};
    outline.append(handle,remove);layer.append(outline);outlines.set(item,outline);observer.observe(item.node);
  }
  function preparePhotos(){
    for(const group of groups.values())for(const item of group.items)for(const image of item.node.querySelectorAll('img[data-snip-photo]')){
      if(!photoState.has(image))photoState.set(image,['tabindex','role','aria-label','style'].map(name=>[name,image.getAttribute(name)]));
      image.tabIndex=0;image.setAttribute('role','button');image.setAttribute('aria-label',image.dataset.photoId?'Change photo':'Add photo');image.style.cursor='pointer';image.style.setProperty('outline','none','important');
      if(!photoHints.has(image)){
        const hint=button('Add photo',()=>showPhoto(image),'add');hint.className='collection-photo-hint';hint.append(document.createTextNode('Add photo'));hint.onfocus=()=>reveal(image);layer.append(hint);photoHints.set(image,hint);
      }
    }
  }
  function clearArrangement(){
    finishDrag();finishAnimations();for(const outline of outlines.values())outline.remove();outlines.clear();
    for(const {node,control} of tiles.values()){node.remove();control.remove();}tiles.clear();
  }
  function render(){
    clearArrangement();observer.disconnect();observer.observe(document.body);preparePhotos();
    if(arranging)for(const [name,group] of groups){
      const prototype=group.prototype.content.firstElementChild;
      const node=document.createElement(prototype.tagName);node.className=prototype.className;
      node.dataset.fleaCollectionAdd=name;node.setAttribute('aria-hidden','true');
      Object.assign(node.style,{visibility:'hidden',minHeight:'160px',pointerEvents:'none'});group.end.before(node);
      const control=button('Add '+label(name).toLowerCase()+' item',()=>{
        const item=addItem(name);if(!item)return;order(group);render();reveal(item.node);
        const field=item.node.querySelector('[contenteditable]');field?.focus();
        if(field){const range=document.createRange();range.selectNodeContents(field);const selection=getSelection();selection.removeAllRanges();selection.addRange(range);}
        announce('Item added. Edit its text and photo on the page.');
      },'add');
      control.className='collection-add';control.dataset.collection=name;
      const text=document.createElement('span');text.textContent='Add item';control.append(text);control.onfocus=()=>reveal(node);
      tiles.set(group,{node,control});layer.append(control);observer.observe(node);
      for(const item of group.items)addOutline(name,group,item);
    }
    updateState();position();
  }
  function updateState(){
    const disabled=!owner.canEdit();
    for(const control of layer.querySelectorAll('button'))control.disabled=disabled;
    for(const [group,{control}] of tiles)control.disabled ||= group.items.length>=200;
    if(disabled)finishDrag();
  }
  let dialogPreview='';
  function clearDialogPreview(){if(dialogPreview)URL.revokeObjectURL(dialogPreview);dialogPreview='';}
  function previewChoice(photo,isUpload=false){
    $('#collection-photo-selection').hidden=!photo;
    const preview=$('#collection-photo-preview');
    if(photo)preview.src=photo.preview||'/media/'+photo.file;else preview.removeAttribute('src');
    $('#collection-photo-name').textContent=photo?.name||'Selected photo';
    $('#collection-photo-alt').value=photo?.alt||'';
    $('#collection-photo-alt').readOnly=!isUpload;
    $('#collection-photo-help').textContent=isUpload?'Describe the photo, or leave blank if decorative.':'Saved photo description. Choosing this photo does not change it.';
    for(const tile of $('#collection-photo-grid').children)tile.setAttribute('aria-pressed',String(!isUpload&&tile.dataset.photoId===$('#collection-photo').value));
  }
  function showPhoto(image){
    if(!owner.canEdit()||!image.isConnected)return;
    clearDialogPreview();selectedPhoto=image;const select=$('#collection-photo'),grid=$('#collection-photo-grid');grid.replaceChildren();
    for(const [id,photo] of [['',null],...Object.entries(photos)]){
      const tile=button(photo?.name||'No photo',()=>{clearDialogPreview();select.value=id;$('#collection-photo-file').value='';$('#collection-photo-error').textContent='';previewChoice(photo);});
      tile.dataset.photoId=id;tile.className='photo-tile';
      const thumb=document.createElement('img');thumb.alt='';thumb.src=photo?(photo.preview||'/media/'+photo.file):host.dataset.photoPlaceholder;tile.append(thumb);
      const caption=document.createElement('span');caption.textContent=photo?.name||'No photo';tile.append(caption);grid.append(tile);
    }
    select.value=image.dataset.photoId||'';$('#collection-photo-empty').hidden=Object.keys(photos).length>0;
    $('#collection-photo-file').value='';$('#collection-photo-alt').value='';$('#collection-photo-error').textContent='';
    previewChoice(photos[select.value]);owner.open($('#collection-photo-dialog'));
  }
  document.addEventListener('click',event=>{
    const image=event.target.closest?.('img[data-snip-photo]');
    if(!owner.canEdit()||!photoState.has(image))return;
    event.preventDefault();showPhoto(image);
  },true);
  document.addEventListener('keydown',event=>{
    if(event.key==='Escape'&&drag){event.preventDefault();finishDrag(true);return;}
    const image=event.target.closest?.('img[data-snip-photo]');
    if(['Enter',' '].includes(event.key)&&owner.canEdit()&&photoState.has(image)){event.preventDefault();showPhoto(image);}
  },true);
  $('#collection-photo-cancel').onclick=()=>owner.close($('#collection-photo-dialog'));
  $('#collection-photo-dialog').addEventListener('close',()=>{if($('#collection-photo-dialog').open)return;clearDialogPreview();$('#collection-photo-preview').removeAttribute('src');$('#collection-photo-file').value='';});
  $('#collection-photo-upload').onclick=()=>$('#collection-photo-file').click();
  $('#collection-photo-file').onchange=()=>{
    clearDialogPreview();const file=$('#collection-photo-file').files[0];if(!file){previewChoice(photos[$('#collection-photo').value]);return;}
    if(file.size===0||file.size>8*1048576||! /\.(jpe?g|png)$/i.test(file.name)){$('#collection-photo-error').textContent='Choose a JPEG or PNG up to 8 MiB.';$('#collection-photo-file').value='';previewChoice(photos[$('#collection-photo').value]);return;}
    $('#collection-photo-error').textContent='';dialogPreview=URL.createObjectURL(file);previewChoice({name:file.name,preview:dialogPreview,alt:''},true);
  };
  $('#collection-photo-dialog form').onsubmit=event=>{
    event.preventDefault();if(!owner.canEdit()||!selectedPhoto?.isConnected)return;
    let id=$('#collection-photo').value;
    const file=$('#collection-photo-file').files[0];
    if(file){
      if(file.size===0||file.size>8*1048576||! /\.(jpe?g|png)$/i.test(file.name)){$('#collection-photo-error').textContent='Choose a JPEG or PNG up to 8 MiB.';return;}
      if(usedUploads().reduce((size,[,upload])=>size+upload.file.size,0)+file.size>38*1048576){$('#collection-photo-error').textContent='Save these changes before adding more photos.';return;}
      do{id='';while(id.length<10)for(const n of crypto.getRandomValues(new Uint8Array(16)))if(n<248&&id.length<10)id+='abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'[n%62];}while(photos[id]);
      const photo={name:file.name,alt:$('#collection-photo-alt').value,preview:URL.createObjectURL(file)};
      photos[id]=photo;uploads.set(id,{file,...photo});
    }
    setPhoto(selectedPhoto,id);preparePhotos();owner.changed();schedule();owner.close($('#collection-photo-dialog'));
  };
  host.collectionUI={
    hasUploads:()=>usedUploads().length>0,
    appendUploads(edit,body){for(const [id,photo] of usedUploads()){edit.media.push({id,name:photo.name,alt:photo.alt,library:true});body.set('media-'+id,photo.file);}},
    updateState,
    start(){clearOutlineFade();arranging=true;layer.hidden=false;render();if(!matchMedia('(prefers-reduced-motion: reduce)').matches)outlineFade=layer.animate([{opacity:0},{opacity:1}],{duration:160,easing:'ease-out'});},
    stop(){
      fadeOutlinesOut();
      clearDialogPreview();
      for(const [id,photo] of uploads){URL.revokeObjectURL(photo.preview);delete photos[id];}uploads.clear();
      finishDrag(true);arranging=false;clearArrangement();observer.disconnect();cancelAnimationFrame(frame);frame=0;
      for(const [image,attrs] of photoState)for(const [name,value] of attrs)value===null?image.removeAttribute(name):image.setAttribute(name,value);
      photoState.clear();photoHints.clear();regionOutlines.clear();layer.replaceChildren();layer.hidden=true;
    }
  };
  if(owner.isEditing())host.collectionUI.start();
})();
