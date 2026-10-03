(() => {
  'use strict';
  const host=document.getElementById('flea-owner-root');if(!host?.shadowRoot)return;
  const ui=host.shadowRoot,$=s=>ui.querySelector(s),panel=$('#theme-panel'),fields=$('#theme-fields');
  host.addEventListener('dev-reload-check',event=>{if(open || loading || saving)event.preventDefault();});
  let open=false,dirty=false,saving=false,loading=false,version='',data=[],changes={},original=new Map();
  let scales={},scaleChanges={};
  let fontLink=null,fontTimer;
  const systemFonts=new Set('serif,sans-serif,monospace,cursive,fantasy,system-ui,ui-serif,ui-sans-serif,ui-monospace,ui-rounded,emoji,math,fangsong,arial,verdana,georgia,times new roman,courier new,helvetica,tahoma,trebuchet ms,impact,comic sans ms,segoe ui,apple color emoji,inherit,initial,unset,revert,revert-layer'.split(','));
  function fontFamily(value){
    const parts=value.split(',');if(parts.length>10)throw new Error('Too many font fallbacks.');
    const names=parts.map(part=>{let name=part.trim();if((name[0]==='"'||name[0]==="'")&&name.at(-1)===name[0])name=name.slice(1,-1);if(!/^[A-Za-z][A-Za-z0-9 -]{0,99}$/.test(name))throw new Error('Enter a font family, such as "Lora", serif.');return name;});
    return systemFonts.has(names[0].toLowerCase())?'':names[0];
  }
  function clearFonts(){clearTimeout(fontTimer);fontLink?.remove();fontLink=null;}
  function previewFonts(){
    clearTimeout(fontTimer);
    if(!['--font-body','--font-heading'].some(key=>Object.hasOwn(changes,key))){clearFonts();return;}
    fontTimer=setTimeout(()=>{
      const families=[];
      for(const key of ['--font-body','--font-heading']){try{const family=fontFamily(changes[key]??data.find(f=>f.name===key)?.value??'');if(family&&!families.includes(family))families.push(family);}catch{}}
      if(!families.length){clearFonts();return;}
      const query=new URLSearchParams();query.set('display','swap');families.forEach(family=>query.append('family',family));
      const href='https://fonts.googleapis.com/css2?'+query;
      if(fontLink?.href===href)return;
      fontLink?.remove();const link=document.createElement('link');fontLink=link;link.rel='stylesheet';link.href=href;
      link.onerror=()=>{if(open&&fontLink===link)message('Google Fonts could not load. Check the family name and your connection.');};
      document.head.append(link);
    },400);
  }
  const scaleNames={text:['sm','base','lg','xl','2xl','3xl','4xl'],space:['3xs','2xs','xs','sm','md','lg','xl','2xl','3xl','4xl']};
  const ratios=[[1.067,'Minor Second'],[1.125,'Major Second'],[1.2,'Minor Third'],[1.25,'Major Third'],[1.333,'Perfect Fourth'],[1.414,'Augmented Fourth'],[1.5,'Perfect Fifth'],[1.618,'Golden Ratio'],[1.667,'Major Sixth'],[1.778,'Minor Seventh'],[1.875,'Major Seventh'],[2,'Octave']];
  const num=v=>{if(!v)return '0';const factor=10**(3-Math.floor(Math.log10(Math.abs(v))));return String(Math.sign(v)*Math.round(Math.abs(v)*factor)/factor);};
  function clamp(s,step){
    const a=s.minBase*s.minRatio**step,b=s.maxBase*s.maxRatio**step,slope=(b-a)/(s.maxViewport-s.minViewport),intercept=(a-slope*s.minViewport)/16;
    return `clamp(${num(Math.min(a,b)/16)}rem, ${num(intercept)}rem ${slope<0?'-':'+'} ${num(Math.abs(slope)*100)}vw, ${num(Math.max(a,b)/16)}rem)`;
  }
  function updateDirty(){
    dirty=Object.keys(changes).length>0 || Object.keys(scaleChanges).length>0;
    $('#theme-save').disabled=!dirty || !fields.checkValidity();
    message(dirty?'Unsaved theme changes. Only you can see this preview.':'Changes are visible only to you until saved.');preview();
  }
  function restore(name){const old=original.get(name);if(old)old.value?document.documentElement.style.setProperty(name,old.value,old.priority):document.documentElement.style.removeProperty(name);}
  function accordion(title,initial=false){
    const group=document.createElement('details'),summary=document.createElement('summary'),body=document.createElement('div');
    group.open=initial;group.dataset.expanded=String(initial);summary.textContent=title;summary.setAttribute('aria-expanded',String(initial));body.className='theme-group';body.style.height=initial?'auto':'0px';
    let timer,frame;
    summary.onclick=event=>{
      event.preventDefault();clearTimeout(timer);cancelAnimationFrame(frame);
      const expanded=group.dataset.expanded!=='true',height=body.getBoundingClientRect().height;
      if(expanded)for(const sibling of group.parentElement.children){if(sibling!==group && sibling.tagName==='DETAILS' && sibling.dataset.expanded==='true')sibling.querySelector('summary').click();}
      group.open=true;body.style.height=height+'px';group.dataset.expanded=String(expanded);summary.setAttribute('aria-expanded',String(expanded));
      const finish=()=>{body.style.height=expanded?'auto':'0px';group.open=expanded;};
      if(matchMedia('(prefers-reduced-motion: reduce)').matches){finish();return;}
      void body.offsetHeight;
      frame=requestAnimationFrame(()=>{body.style.height=expanded?body.scrollHeight+'px':'0px';timer=setTimeout(finish,250);});
    };
    group.append(summary,body);return group;
  }
  function scaleControls(kind,body){
    const info=scales[kind],s={...info.settings},controls=document.createElement('fieldset'),samples=document.createElement('div');
    controls.className='theme-scale-controls';samples.className='theme-scale-samples';
    let replaced=false;
    if(info.custom){
      const notice=document.createElement('p'),button=document.createElement('button');notice.textContent='Custom scale';button.type='button';button.textContent='Replace with generated scale';controls.disabled=true;
      button.onclick=()=>{if(!confirm('Replace the custom sizes with a generated scale?'))return;replaced=true;controls.disabled=false;button.remove();notice.remove();update();};body.append(notice,button);
    }
    const inputs={};
    function update(){
      for(const input of Object.values(inputs))input.setCustomValidity('');
      if(s.maxViewport<=s.minViewport)inputs.maxViewport.setCustomValidity('Large screen width must exceed small screen width.');
      if(!fields.checkValidity()){$('#theme-save').disabled=true;message('Check the highlighted value.');return;}
      const changed=replaced || Object.keys(s).some(key=>s[key]!==info.settings[key]);
      if(changed)scaleChanges[kind]={settings:{...s},replace:replaced};else delete scaleChanges[kind];
      samples.replaceChildren();
      scaleNames[kind].forEach((name,i)=>{
        const prop=`--${kind}-${name}`,value=clamp(s,i-(kind==='text'?1:3));
        if(changed)document.documentElement.style.setProperty(prop,value,'important');else restore(prop);
        const sample=document.createElement('div');sample.title=value;
        if(kind==='text'){sample.textContent=name+' Aa';sample.style.fontSize=value;sample.style.lineHeight='1.2';}
        else {const caption=document.createElement('span'),bar=document.createElement('i');caption.textContent=name;bar.style.width=value;sample.append(caption,bar);}
        samples.append(sample);
      });updateDirty();
    }
    const heading=document.createElement('div');heading.className='theme-scale-columns theme-scale-heading';heading.innerHTML='<span>Small screen</span><span>Large screen</span>';controls.append(heading);
    function pair(title,key,min,max,step,parent){
      const row=document.createElement('div'),caption=document.createElement('span');row.className='theme-scale-columns';caption.textContent=title;row.append(caption);
      for(const prefix of ['min','max']){
        const name=prefix+key,input=document.createElement(key==='Ratio'?'select':'input');input.setAttribute('aria-label',`${kind==='text'?'Text':'Spacing'} ${prefix==='min'?'small':'large'} screen ${title.toLowerCase()}`);
        if(key==='Ratio'){
          for(const [value,label] of [...ratios,...(ratios.some(r=>r[0]===s[name])?[]:[[s[name],'Custom']])]){const option=document.createElement('option');option.value=value;option.textContent=value+' · '+label;input.append(option);}
          const wrap=document.createElement('div');wrap.className='theme-select';wrap.append(input);row.append(wrap);
        }else{input.type='number';input.required=true;input.min=min;input.max=max;input.step=step;
          if(key==='Viewport'){const wrapper=document.createElement('label');wrapper.textContent=prefix==='min'?'Min':'Max';input.setAttribute('aria-label',`${kind==='text'?'Text':'Spacing'} viewport ${prefix} (px)`);wrapper.append(input);row.append(wrapper);}else row.append(input);
        }
        input.value=s[name];inputs[name]=input;input.oninput=()=>{s[name]=Number(input.value);update();};
      }parent.append(row);
    }
    pair(kind==='text'?'Base size (px)':'Small spacing (px)','Base',1,200,0.1,controls);
    pair('Scale ratio','Ratio',1,2,0.001,controls);
    const advanced=accordion('Viewport width');pair('Pixels','Viewport',100,10000,1,advanced.querySelector('.theme-group'));controls.append(advanced);
    body.append(controls,samples);
    if(!info.custom)update();
  }
  const labels={'--bg-color':'Background Color','--max-w-text':'Text Max Width','--max-w-form':'Form Max Width','--container-max':'Container Max Width','--font-body':'Body Font','--font-heading':'Heading Font'};
  const label=name=>labels[name] || name.replace(/^--/,'').replaceAll('-',' ').replace(/\b\w/g,c=>c.toUpperCase());
  const help={
    '--container-max':'Maximum width of content using the container class. On smaller screens it still fits the available width.',
    '--container-padding':'Space inside containers. Two values mean vertical then horizontal padding.',
    '--max-w-text':'Maximum width for text using max-w-text. The ch unit is based on the width of a zero in the current font.',
    '--max-w-form':'Maximum width for forms using max-w-form. One rem is the root font size.',
    '--font-body':'Paste a Google Fonts family, such as "Source Sans 3", sans-serif. Preview loads it automatically; Save adds the import to theme.css. System fonts need no import.',
    '--font-heading':'Paste a Google Fonts family for headings, such as "Lora", serif. Do not paste a stylesheet link or @import. The regular face loads automatically; browsers can synthesize bold and italic.',
    '--font-normal':'Default text weight. Most fonts use 400. The chosen font must support this weight.',
    '--font-semibold':'Semibold text weight, usually 600. The chosen font must support this weight.',
    '--font-bold':'Bold text weight, usually 700. The chosen font must support this weight.',
    '--field-height':'Default height of form fields. Textareas can grow taller.',
    '--field-padding':'Space between a form field’s border and its text.'
  };
  const message=text=>{$('#theme-message').textContent=text;};
  const property=f=>f.kind==='color'?'color':f.kind==='font'?'font-family':f.name.startsWith('--shadow')?'box-shadow':f.name==='--field-border'?'border':f.name.startsWith('--font-')?'font-weight':['--container-padding','--field-padding'].includes(f.name)?'padding':'width';
  const sampleColor=value=>{
    const canvas=document.createElement('canvas'),ctx=canvas.getContext('2d');canvas.width=canvas.height=1;
    ctx.fillStyle=value;ctx.fillRect(0,0,1,1);return '#'+[...ctx.getImageData(0,0,1,1).data].slice(0,3).map(v=>v.toString(16).padStart(2,'0')).join('');
  };
  function preview(){
    const computed=getComputedStyle(document.documentElement),palette=$('#theme-palette');palette.replaceChildren();
    for(const name of ['--brand-1','--brand-2','--brand-3','--brand-4','--black','--white']){const chip=document.createElement('span');chip.style.background=computed.getPropertyValue(name);chip.title=label(name);palette.append(chip);}
  }
  function change(f,input){
    const value=input.value.trim();input.setCustomValidity(CSS.supports(property(f),value)?'':'Enter a valid value.');
    if(f.kind==='font'){try{fontFamily(value);}catch(error){input.setCustomValidity(error.message);}}
    if(!input.validity.valid){$('#theme-save').disabled=true;message('Check the highlighted value.');return;}
    if(value===f.value)delete changes[f.name];else changes[f.name]=value;
    const originalValue=original.get(f.name);
    if(value===f.value){originalValue.value?document.documentElement.style.setProperty(f.name,originalValue.value,originalValue.priority):document.documentElement.style.removeProperty(f.name);}
    else document.documentElement.style.setProperty(f.name,value,'important');
    updateDirty();
    if(f.kind==='font')previewFonts();
  }
  function build(){
    fields.replaceChildren();const groups=new Map();
    for(const f of data)original.set(f.name,{value:document.documentElement.style.getPropertyValue(f.name),priority:document.documentElement.style.getPropertyPriority(f.name)});
    const order=['Brand colors','Neutral colors','Semantic colors','Page colors','Typography','Text sizes','Spacing','Layout','Shape & shadows','Forms','Other'];
    for(const f of [...data].sort((a,b)=>order.indexOf(a.group)-order.indexOf(b.group))){
      let group=groups.get(f.group);if(!group){
        group=accordion(f.group,false);groups.set(f.group,group);fields.append(group);
      }
      const kind=f.name.startsWith('--text-')?'text':f.name.startsWith('--space-')?'space':null;
      if(kind && scales[kind] && scaleNames[kind].includes(f.name.slice(kind.length+3))){
        if(!group.dataset.scale){group.dataset.scale=kind;scaleControls(kind,group.querySelector('.theme-group'));}continue;
      }
      const row=document.createElement('div'),title=document.createElement('label'),controls=document.createElement('div');row.className='theme-field';title.textContent=label(f.name);controls.className='theme-control';row.append(title,controls);
      const fontColor=['--font-color','--bg-color'].includes(f.name);
      const input=document.createElement(fontColor?'select':'input');input.id='theme-value'+f.name;title.htmlFor=input.id;input.setAttribute('aria-label',label(f.name));
      if(fontColor){
        for(const category of ['Neutral colors','Brand colors']){
          const group=document.createElement('optgroup');group.label=category;
          for(const color of data.filter(c=>c.group===category && c.kind==='color')){const option=document.createElement('option');option.value=`var(${color.name})`;option.textContent=label(color.name);group.append(option);}
          if(group.children.length)input.append(group);
        }
        if(![...input.options].some(option=>option.value===f.value)){const current=document.createElement('option');current.value=f.value;current.textContent='Choose a theme color';current.disabled=true;input.prepend(current);}
      }else{input.type='text';input.spellcheck=false;input.autocomplete='off';}
      if(help[f.name]){
        const head=document.createElement('div'),button=document.createElement('button'),bubble=document.createElement('p');head.className='theme-label';button.type='button';button.className='theme-info';button.textContent='i';button.setAttribute('aria-label','About '+label(f.name));button.setAttribute('aria-expanded','false');bubble.id='theme-help'+f.name;bubble.className='theme-help';bubble.textContent=help[f.name];bubble.hidden=true;button.setAttribute('aria-controls',bubble.id);input.setAttribute('aria-describedby',bubble.id);
        button.onclick=()=>{bubble.hidden=!bubble.hidden;button.setAttribute('aria-expanded',String(!bubble.hidden));};head.append(title,button);row.prepend(head);row.insertBefore(bubble,controls);
      }
      input.value=f.value;input.oninput=()=>change(f,input);
      if(f.kind==='color' && !fontColor){
        const swatch=document.createElement('input');swatch.type='color';swatch.setAttribute('aria-label',label(f.name)+' color picker');
        swatch.value=sampleColor(getComputedStyle(document.documentElement).getPropertyValue(f.name).trim());
        swatch.oninput=()=>{input.value=swatch.value;change(f,input);};controls.append(swatch);
      }
      if(fontColor){const wrap=document.createElement('div');wrap.className='theme-select';wrap.append(input);controls.append(wrap);}else controls.append(input);
      group.querySelector('.theme-group').append(row);
    }
    preview();
  }
  $('#theme-cancel').replaceChildren($('#collection-remove').content.cloneNode(true));
  let exitAnimation=null;
  function close(){
    if(saving || loading)return false;
    if(dirty && !confirm('Discard theme changes? Choose Cancel to save them first.'))return false;
    clearFonts();
    $('#cancel').hidden=$('#save').hidden=false;$('#edit').disabled=false;
    for(const [name,old] of original){old.value?document.documentElement.style.setProperty(name,old.value,old.priority):document.documentElement.style.removeProperty(name);}
    original.clear();dirty=open=false;changes={};scaleChanges={};
    panel.inert=true;
    const finish=()=>{panel.hidden=true;panel.inert=false;};
    if(matchMedia('(prefers-reduced-motion: reduce)').matches)finish();
    else {const animation=panel.animate([{transform:getComputedStyle(panel).transform},{transform:'translateX(100%)'}],{duration:220,easing:'ease-in',fill:'forwards'});exitAnimation=animation;animation.finished.then(()=>{if(exitAnimation===animation){finish();animation.cancel();exitAnimation=null;}},()=>{});}
    $('#theme').setAttribute('aria-expanded','false');$('#theme').removeAttribute('aria-pressed');host.dispatchEvent(new Event('theme-closed'));$('#theme').focus();return true;
  }
  host.addEventListener('theme-open',async()=>{
    if(open){close();return;}if(loading || saving)return;
    exitAnimation?.cancel();exitAnimation=null;panel.inert=false;
    panel.hidden=true;void panel.offsetWidth;
    open=loading=true;dirty=false;changes={};scaleChanges={};panel.hidden=false;$('#theme').setAttribute('aria-expanded','true');$('#theme').setAttribute('aria-pressed','true');$('#theme-save').disabled=true;fields.replaceChildren();message('Loading theme…');
    $('#edit').disabled=true;$('#cancel').hidden=$('#save').hidden=true;$('#theme-cancel').hidden=$('#theme-save').hidden=false;
    host.style.setProperty('--owner-height',host.getBoundingClientRect().bottom+'px');$('#theme-title').focus();
    try{
      const response=await fetch('/admin/theme',{credentials:'same-origin',redirect:'error',signal:AbortSignal.timeout(15000)});
      if(!response.ok)throw new Error('Theme unavailable. Ask your developer to check css/theme.css.');
      const result=await response.json();version=result.version;data=result.fields;scales=result.scales;build();message('Changes are visible only to you until saved.');
    }catch(error){message(error.message);}finally{loading=false;}
  });
  $('#theme-cancel').onclick=close;
  $('#logout').addEventListener('click',event=>{if(open && !close()){event.preventDefault();event.stopImmediatePropagation();}},true);
  fields.onsubmit=event=>event.preventDefault();
  $('#theme-save').onclick=async()=>{
    if(!dirty || saving || !fields.reportValidity())return;saving=true;$('#theme-save').disabled=$('#theme-cancel').disabled=$('#theme').disabled=$('#edit').disabled=$('#logout').disabled=true;fields.inert=true;message('Saving theme…');
    try{
      const response=await fetch('/admin/theme',{method:'POST',credentials:'same-origin',redirect:'error',signal:AbortSignal.timeout(15000),body:new URLSearchParams({csrf:host.dataset.csrf,edit:JSON.stringify({version,changes,scales:scaleChanges})})});
      if(response.status!==204)throw new Error(response.status===403?'Your session expired. Copy your changes before signing in again.':'Could not save. The theme may have changed. Your preview is still here; copy your changes before reloading.');
      dirty=false;location.reload();
    }catch(error){message(error.message);saving=false;fields.inert=false;$('#theme-save').disabled=$('#theme-cancel').disabled=$('#theme').disabled=$('#logout').disabled=false;$('#edit').disabled=true;}
  };
  document.addEventListener('keydown',event=>{if(open && event.key==='Escape' && !event.defaultPrevented){event.preventDefault();close();}});
  window.addEventListener('beforeunload',event=>{if(dirty){event.preventDefault();event.returnValue='';}});
})();
