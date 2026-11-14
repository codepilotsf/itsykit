;(()=>{
 const form=document.getElementById('blog-editor');if(!form)return;
 const field=name=>form.elements.namedItem(name),message=document.getElementById('blog-message');
 const textarea=field('body'),editorHost=document.createElement('div');
 editorHost.className='blog-rich-editor';editorHost.setAttribute('role','textbox');editorHost.setAttribute('aria-label','Body');editorHost.setAttribute('aria-multiline','true');
 textarea.before(editorHost);textarea.hidden=true;
 const policy={ALLOWED_TAGS:['span','p','br','hr','strong','b','em','i','u','s','del','mark','small','sub','sup','h1','h2','h3','h4','h5','h6','ul','ol','li','blockquote','pre','code','a'],ALLOWED_ATTR:['href','title','style'],ALLOW_DATA_ATTR:false,ALLOW_ARIA_ATTR:false};
 DOMPurify.addHook('uponSanitizeAttribute',(_node,attribute)=>{if(attribute.attrName==='style'&&!/^\s*text-align\s*:\s*(left|center|right)\s*;?\s*$/i.test(attribute.attrValue))attribute.keepAttr=false});
 const clean=source=>DOMPurify.sanitize(source,policy);
 const editor=new Squire(editorHost,{blockTag:'P',addLinks:false,sanitizeToDOMFragment:source=>DOMPurify.sanitize(source,{...policy,RETURN_DOM_FRAGMENT:true}),undo:{documentSizeThreshold:-1,undoLimit:80}});
 editor.setHTML(clean(textarea.value));
 const tools=document.createElement('div');tools.className='blog-format';tools.setAttribute('role','group');tools.setAttribute('aria-label','Body formatting');editorHost.before(tools);
 const formatBlock=tag=>editor.modifyBlocks(f=>{for(const node of [...f.querySelectorAll('p,div,h1,h2,h3,h4')].reverse()){if(node.querySelector('p,div,h1,h2,h3,h4,ul,ol'))continue;const next=document.createElement(tag);next.append(...node.childNodes);node.replaceWith(next)}return f});
 const buttons=[['Bold',()=>editor.bold()],['Italic',()=>editor.italic()],['Heading',()=>formatBlock('h2')],['Paragraph',()=>formatBlock('p')],['Bullet list',()=>editor.makeUnorderedList()],['Numbered list',()=>editor.makeOrderedList()],['Link',()=>{const url=prompt('Link URL (https:// or a website path)');if(url&&/^(https?:\/\/|\/(?!\/))/.test(url))editor.makeLink(url)}]];
 for(const [label,action]of buttons){const b=document.createElement('button');b.type='button';b.textContent=label;b.addEventListener('mousedown',e=>e.preventDefault());b.addEventListener('click',()=>{action();editor.focus()});tools.append(b)}
 let dirty=false,busy=false,objectURL='',uploadID='';const mark=()=>dirty=true;
 form.addEventListener('input',mark);form.addEventListener('change',mark);editor.addEventListener('input',mark);
 const setState=()=>{document.getElementById('blog-state').textContent='Status: '+form.dataset.status+(form.dataset.status==='published'?'. Saving updates the public post immediately.':'');form.querySelector('[value="publish"]').hidden=form.dataset.status==='published';form.querySelector('[value="unpublish"]').hidden=form.dataset.status!=='published'};setState();
 const preview=document.createElement('img');preview.id='cover-preview';preview.className='blog-cover-preview';preview.alt='';preview.hidden=true;document.getElementById('cover-preview-slot').append(preview);
 const select=field('cover'),file=field('cover_file');select.value=field('saved_cover').value;
 const clearURL=()=>{if(objectURL){URL.revokeObjectURL(objectURL);objectURL=''}};
 const updatePreview=()=>{const option=select.selectedOptions[0];preview.hidden=!option?.dataset.src;if(option?.dataset.src)preview.src=option.dataset.src;else preview.removeAttribute('src');preview.alt=option?.dataset.alt||''};updatePreview();
 select.addEventListener('change',()=>{file.value='';uploadID='';clearURL();updatePreview()});
 file.addEventListener('change',()=>{clearURL();const photo=file.files[0];if(!photo){updatePreview();return}if(photo.size>8*1024*1024||!['image/jpeg','image/png'].includes(photo.type)){message.textContent='Choose a JPEG or PNG up to 8 MiB.';file.value='';updatePreview();return}uploadID=Array.from(crypto.getRandomValues(new Uint8Array(10)),(v,i)=>(i?'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789':'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz')[v%(i?62:52)]).join('');objectURL=URL.createObjectURL(photo);preview.src=objectURL;preview.alt=field('cover_alt').value;preview.hidden=false;select.value=''});
 const dialog=document.getElementById('cover-library');document.getElementById('choose-cover').addEventListener('click',()=>dialog.showModal());document.getElementById('close-library').addEventListener('click',()=>dialog.close());
 dialog.addEventListener('click',e=>{const button=e.target.closest('[data-photo]');if(!button)return;select.value=button.dataset.photo;select.dispatchEvent(new Event('change',{bubbles:true}));dialog.close()});
 let autoSlug=!field('slug').value;field('slug').addEventListener('input',()=>autoSlug=false);
 field('title').addEventListener('input',()=>{if(autoSlug)field('slug').value=field('title').value.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'')});
 window.addEventListener('beforeunload',e=>{if(dirty){e.preventDefault();e.returnValue=''}});
 document.getElementById('blog-cancel').addEventListener('click',e=>{if(busy||dirty&&!confirm('Discard your unsaved changes?')){e.preventDefault();return}dirty=false;clearURL()});
 form.addEventListener('submit',async e=>{
  e.preventDefault();if(busy)return;const action=e.submitter?.value||'save';
  const post={id:field('id').value,title:field('title').value,slug:field('slug').value,excerpt:field('excerpt').value,body:editor.getHTML(),status:form.dataset.status,published_at:field('published_at').value,cover:uploadID||select.value};
  const edit={version:field('version').value,post,action};const data=new FormData();data.set('csrf',field('csrf').value);
  if(file.files[0]){edit.upload={id:uploadID,name:file.files[0].name,alt:field('cover_alt').value};data.set('cover',file.files[0])}data.set('edit',JSON.stringify(edit));
  busy=true;form.inert=true;message.textContent='Saving…';
  try{const response=await fetch('/admin/blog/save',{method:'POST',body:data,headers:{'X-CSRF-Token':field('csrf').value},signal:AbortSignal.timeout(240000)});if(!response.ok)throw Error(await response.text());const result=await response.json();field('version').value=result.version;field('published_at').value=result.post.published_at||'';textarea.value=result.post.body;editor.setHTML(clean(result.post.body));form.dataset.status=result.post.status;setState();dirty=false;history.replaceState(null,'','/admin/blog/edit?id='+encodeURIComponent(result.post.id));message.removeAttribute('role');message.setAttribute('role','status');message.textContent=action==='unpublish'?'Unpublished. This post is now private.':result.post.status==='published'?'Saved and published.':'Draft saved.';
   if(uploadID){const option=document.createElement('option');option.value=uploadID;option.textContent=file.files[0].name;option.dataset.src='/media/'+result.cover.file;option.dataset.alt=result.cover.alt;select.append(option);select.value=uploadID;uploadID='';file.value='';field('cover_alt').value='';clearURL();updatePreview()}
  }catch(error){message.textContent=error.message||'Could not save. Your draft is still here.';message.setAttribute('role','alert')}
  finally{busy=false;form.inert=false}
 });
})();
