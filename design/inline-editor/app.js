/* Standalone experiment. No authentication, site writes, or production routes. */
(() => {
  'use strict';
  const $ = (selector) => document.querySelector(selector);
  const blocks = 'p,h1,h2,h3,h4,li';
  function normalizeLink(value) {
    const text = value.trim();
    if (/^[^\s/@:]+(?:\.[^\s/@:]+)+(?:[/:?#]|$)/.test(text)) return 'https://' + text;
    return text;
  }
  const safeLink = (value) => {
    const text = value.trim();
    if (!text || /[\\\u0000-\u0020\u007f]/.test(text) || text.startsWith('//')) return false;
    try { return ['http:', 'https:', 'mailto:', 'tel:'].includes(new URL(text, location.href).protocol); }
    catch { return false; }
  };
  DOMPurify.addHook('uponSanitizeAttribute', (node, attribute) => {
    if (attribute.attrName === 'href') attribute.keepAttr = node.nodeName === 'A' && safeLink(attribute.attrValue);
    if (attribute.attrName === 'title') attribute.keepAttr = node.nodeName === 'A';
    if (attribute.attrName === 'target') attribute.keepAttr = node.nodeName === 'A' && attribute.attrValue === '_blank';
    if (attribute.attrName === 'rel') attribute.keepAttr = false;
    if (attribute.attrName === 'style') {
      attribute.keepAttr = node.matches(blocks) && /^\s*text-align\s*:\s*(left|center|right)\s*;?\s*$/i.test(attribute.attrValue);
    }
  });
  function fragment(html) {
    const result = DOMPurify.sanitize(html, {
      ALLOWED_TAGS: ['p', 'div', 'br', 'strong', 'b', 'em', 'i', 'u', 's', 'del', 'a', 'h1', 'h2', 'h3', 'h4', 'ul', 'ol', 'li'],
      ALLOWED_ATTR: ['href', 'title', 'style', 'target'],
      ALLOW_DATA_ATTR: false, ALLOW_ARIA_ATTR: false, RETURN_DOM_FRAGMENT: true,
    });
    for (const link of result.querySelectorAll('a[target="_blank"]')) link.setAttribute('rel', 'noopener noreferrer');
    return result;
  }
  function cleanHTML(html) {
    const root = document.createElement('div');
    root.append(fragment(html));
    // Squire can use divs for pasted/default blocks; store semantic paragraphs.
    for (const div of [...root.querySelectorAll('div')].reverse()) {
      if (div.querySelector('p,div,h1,h2,h3,h4,ul,ol')) div.replaceWith(...div.childNodes);
      else { const p = document.createElement('p'); p.append(...div.childNodes); div.replaceWith(p); }
    }
    return root.innerHTML;
  }
  function createEditor(node, html) {
    const editor = new Squire(node, { blockTag: 'P', addLinks: false, sanitizeToDOMFragment: fragment, undo: { documentSizeThreshold: -1, undoLimit: 80 } });
    editor.setHTML(cleanHTML(html));
    return editor;
  }
  function heading(editor, tag) {
    if (!['P', 'H1', 'H2', 'H3', 'H4'].includes(tag)) return;
    editor.modifyBlocks((frag) => {
      for (const node of [...frag.querySelectorAll('p,div,h1,h2,h3,h4')].reverse()) {
        if (node.querySelector('p,div,h1,h2,h3,h4,ul,ol')) continue;
        const next = document.createElement(tag);
        if (node.style.textAlign) next.style.textAlign = node.style.textAlign;
        next.append(...node.childNodes);
        node.replaceWith(next);
      }
      for (const li of frag.querySelectorAll('li')) {
        if (li.querySelector('p,h1,h2,h3,h4,ul,ol')) continue;
        const next = document.createElement(tag);
        next.append(...li.childNodes);
        li.append(next);
      }
      return frag;
    });
  }
  const pages = {
    '/': { label: 'Homepage', title: 'Little Field Studio', description: 'Thoughtful websites for small businesses.', html: '<h1>A little website.<br>A lot of possibility.</h1><p>We help independent businesses find their voice online. Simple pages, useful words, and room to grow.</p><h2>Made for the everyday</h2><p>Your website should be as easy to care for as it is to visit. Update the words yourself and keep the design you love.</p><ul><li>Clear, readable content</li><li>A layout that feels like you</li><li>No complicated publishing routine</li></ul>' },
    '/about': { label: 'About', title: 'About the studio', description: 'A small studio with a practical approach.', html: '<h1>Small by choice.</h1><p>We are a small team making thoughtful websites. We believe useful things can also be beautiful.</p><h2>How we work</h2><p>Start with a conversation. Decide what matters. Build only what you need.</p><h3>Keep it personal</h3><p>Your story deserves your own words.</p><h4>The little details</h4><p>Good spacing and clear language go a long way.</p>' },
    '/contact': { label: 'Contact', title: 'Say hello', description: 'Start a conversation with Little Field Studio.', html: '<h1>Let’s talk.</h1><p>Tell us a little about your project. We would love to hear from you.</p><p>Email <a href="mailto:hello@example.com">hello@example.com</a>.</p>' },
    '/thank-you': { label: 'Thank you', title: 'Thanks for getting in touch', description: 'Your message has arrived.', html: '<h1>Thanks for saying hello.</h1><p>Your message is with us. We will get back to you soon.</p>' },
    '/privacy': { label: 'Privacy (no text snips)', title: 'Privacy', description: 'Our privacy information.', html: null },
  };
  let shared = '<p>A little website. All yours.</p><p>Little Field Studio · Independent by design.</p>';
  let route = '/', editing = false, active = null, dirty = false, draft = null, signedIn = true;
  const instances = new Map();
  const setStatus = (text) => { $('#status').textContent = text; };
  function markDirty() {
    if (!editing) return;
    dirty = draft.title !== pages[route].title || draft.description !== pages[route].description ||
      [...instances.values()].some(({ editor, baseline }) => cleanHTML(editor.getHTML()) !== baseline);
    $('#save').disabled = !dirty;
    setStatus(dirty ? 'Unsaved changes' : 'Click text to edit.');
  }
  function syncToolbar() {
    const editor = active?.editor;
    if (!editor) {
      $('#block').value = 'P';
      for (const button of document.querySelectorAll('[data-align]')) button.setAttribute('aria-pressed', 'false');
    }
    for (const control of $('#formatting').querySelectorAll('button:not(#details-button),select')) control.disabled = !editor;
    const linked = !!editor?.hasFormat('A');
    $('#link-button').setAttribute('aria-label', linked ? 'Edit link' : 'Add link');
    $('#link-button').title = linked ? 'Edit link' : 'Add link';
    $('#link-button').setAttribute('aria-pressed', String(linked));
    for (const button of document.querySelectorAll('[data-mark]')) button.setAttribute('aria-pressed', String(!!editor?.hasFormat(button.dataset.mark)));
    for (const button of document.querySelectorAll('[data-list]')) button.setAttribute('aria-pressed', String(!!editor?.hasFormat(button.dataset.list)));
    if (editor) {
      let node = editor.getSelection().startContainer;
      if (node.nodeType !== Node.ELEMENT_NODE) node = node.parentElement;
      const block = node.closest('p,h1,h2,h3,h4,li');
      $('#block').value = /^H[1-4]$/.test(block?.nodeName) ? block.nodeName : 'P';
      for (const button of document.querySelectorAll('[data-align]')) button.setAttribute('aria-pressed', String(button.dataset.align === (block?.style.textAlign || 'left')));
    }
    $('#region-label').textContent = active?.shared ? 'Shared content: changes apply everywhere.' : active ? 'Page content' : 'Click editable text to enable formatting.';
  }
  function dispose() {
    for (const { editor, node } of instances.values()) { editor.destroy(); node.removeAttribute('contenteditable'); }
    instances.clear(); active = null;
  }
  function region(parent, id, html, isShared) {
    const node = document.createElement('div');
    node.dataset.region = id;
    node.setAttribute('aria-label', isShared ? 'Shared content: changes apply everywhere' : 'Editable page content');
    node.append(fragment(html));
    parent.replaceChildren(node);
    if (!editing) return;
    node.setAttribute('role', 'textbox'); node.setAttribute('aria-multiline', 'true');
    const editor = createEditor(node, html);
    const entry = { editor, node, shared: isShared, baseline: cleanHTML(editor.getHTML()), canUndo: false, canRedo: false };
    instances.set(id, entry);
    editor.addEventListener('focus', () => {
      active = entry;
      syncToolbar();
    });
    editor.addEventListener('pathChange', syncToolbar);
    editor.addEventListener('select', syncToolbar);
    editor.addEventListener('undoStateChange', (event) => {
      entry.canUndo = event.detail.canUndo; entry.canRedo = event.detail.canRedo;
      if (entry === active) syncToolbar();
    });
    editor.addEventListener('input', markDirty);
  }
  function show(nextRoute, edit, push = true) {
    dispose();
    route = pages[nextRoute] ? nextRoute : '/';
    editing = edit && signedIn; dirty = false; draft = { ...pages[route] };
    document.body.classList.toggle('editing', editing);
    $('#formatting').hidden = !editing;
    $('#save').hidden = $('#cancel').hidden = !editing;
    $('#edit').hidden = editing;
    $('#save').disabled = true;
    $('#details-button').disabled = !editing;
    $('#page-label').textContent = pages[route].label;
    if (pages[route].html === null) {
      $('#content').innerHTML = '<h1>Privacy, in plain language.</h1><p>This developer-owned page has no editable text snips. Its SEO title and description are still editable through Page details.</p>';
    } else region($('#content'), 'page', pages[route].html, false);
    region($('#shared'), 'shared', shared, true);
    document.title = pages[route].title + ' · Flea prototype';
    setStatus(editing ? 'Click text to edit.' : '');
    syncToolbar();
    if (push) history.pushState({ route, editing }, '', '?page=' + encodeURIComponent(route) + (editing ? '&edit=1' : ''));
  }
  function canLeave() { return !dirty || confirm('Discard your unsaved changes?'); }
  let pageTransition;
  let navigationRequest = 0;
  async function navigate(nextRoute, edit = false, push = true) {
    const request = ++navigationRequest;
    if (pageTransition) {
      pageTransition.skipTransition();
      await pageTransition.finished;
    }
    if (request !== navigationRequest) return;
    const update = () => show(nextRoute, edit, push);
    if (nextRoute !== route && document.startViewTransition && !matchMedia('(prefers-reduced-motion: reduce)').matches) {
      const transition = document.startViewTransition(update);
      pageTransition = transition;
      try { await transition.finished; } finally { if (pageTransition === transition) pageTransition = null; }
    }
    else update();
  }
  const modalAnimations = new WeakMap();
  function openDialog(dialog) {
    dialog.classList.remove('closing');
    dialog.showModal();
    if (!matchMedia('(prefers-reduced-motion: reduce)').matches) {
      modalAnimations.set(dialog, dialog.animate([
        { opacity: 0, transform: 'translateY(-20px)' },
        { opacity: 1, transform: 'translateY(0)' },
      ], { duration: 280, easing: 'ease-out' }));
    }
  }
  async function closeDialog(dialog) {
    if (!dialog.open || dialog.classList.contains('closing')) return;
    modalAnimations.get(dialog)?.cancel();
    if (!matchMedia('(prefers-reduced-motion: reduce)').matches) {
      dialog.classList.add('closing');
      await dialog.animate([
        { opacity: 1, transform: 'translateY(0)' },
        { opacity: 0, transform: 'translateY(-20px)' },
      ], { duration: 240, easing: 'ease-in', fill: 'forwards' }).finished;
    }
    dialog.close();
    dialog.getAnimations().forEach((animation) => animation.cancel());
    dialog.classList.remove('closing');
  }
  function clearOutside(target) {
    if (editing && !target.closest('[data-region], #formatting, dialog')) {
      active = null; syncToolbar();
    }
  }
  document.addEventListener('pointerdown', (event) => {
    document.body.classList.remove('keyboard-focus'); clearOutside(event.target);
  });
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Tab') document.body.classList.add('keyboard-focus');
  });
  document.addEventListener('focusin', (event) => clearOutside(event.target));
  $('#edit').addEventListener('click', () => show(route, true));
  document.addEventListener('click', (event) => {
    const link = event.target.closest('a');
    if (!link) return;
    if (editing && link.closest('[data-region]')) { event.preventDefault(); return; }
    if (link.dataset.page) { event.preventDefault(); if (canLeave()) navigate(link.dataset.page); }
  });
  $('#formatting').addEventListener('mousedown', (event) => {
    // Mouse toolbar use must not collapse the selection. Keyboard focus remains usable.
    if (event.target.closest('button')) event.preventDefault();
  });
  function command(action) {
    if (!active) return;
    active.editor.focus(); action(active.editor); markDirty(); syncToolbar();
  }
  for (const button of document.querySelectorAll('[data-mark]')) button.addEventListener('click', () => command((editor) => {
    const mark = { tag: button.dataset.mark };
    editor.changeFormat(editor.hasFormat(mark.tag) ? null : mark, editor.hasFormat(mark.tag) ? mark : null);
  }));
  for (const button of document.querySelectorAll('[data-command]')) button.addEventListener('click', () => command((editor) => editor[button.dataset.command]()));
  for (const button of document.querySelectorAll('[data-list]')) button.addEventListener('click', () => command((editor) => {
    if (editor.hasFormat(button.dataset.list)) editor.removeList();
    else editor[button.dataset.list === 'UL' ? 'makeUnorderedList' : 'makeOrderedList']();
  }));
  $('#block').addEventListener('change', (event) => { const value = event.target.value; command((editor) => heading(editor, value)); });
  for (const button of document.querySelectorAll('[data-align]')) button.addEventListener('click', () => command((editor) => editor.setTextAlignment(button.dataset.align)));
  let linkRange;
  $('#link-button').addEventListener('click', () => {
    if (!active) return;
    linkRange = active.editor.getSelection().cloneRange();
    let element = linkRange.startContainer;
    if (element.nodeType !== Node.ELEMENT_NODE) element = element.parentElement;
    const anchor = element.closest('a');
    // Edit the whole existing link, even when only the caret is inside it.
    if (anchor) linkRange.selectNodeContents(anchor);
    $('#link-url').value = anchor?.getAttribute('href') || '';
    $('#link-new-tab').checked = anchor?.getAttribute('target') === '_blank';
    $('#link-remove').hidden = !anchor;
    $('#link-error').textContent = ''; openDialog($('#link-dialog'));
  });
  $('#link-cancel').addEventListener('click', () => closeDialog($('#link-dialog')));
  $('#link-remove').addEventListener('click', async () => {
    if ($('#link-dialog').classList.contains('closing')) return;
    await closeDialog($('#link-dialog'));
    command((editor) => { editor.setSelection(linkRange); editor.removeLink(); });
  });
  $('#link-dialog').addEventListener('cancel', (event) => { event.preventDefault(); closeDialog($('#link-dialog')); });
  $('#link-form').addEventListener('submit', async (event) => {
    event.preventDefault();
    if ($('#link-dialog').classList.contains('closing')) return;
    const url = normalizeLink($('#link-url').value);
    const newTab = $('#link-new-tab').checked;
    if (!safeLink(url)) { $('#link-error').textContent = 'Use a safe web, email, phone, or relative address.'; return; }
    await closeDialog($('#link-dialog'));
    command((editor) => {
      editor.setSelection(linkRange);
      editor.makeLink(url, newTab ? { target: '_blank', rel: 'noopener noreferrer' } : {});
    });
  });
  let detailsBefore;
  function cancelDetails() {
    draft.title = detailsBefore.title; draft.description = detailsBefore.description;
    markDirty(); closeDialog($('#page-details'));
  }
  $('#details-cancel').addEventListener('click', cancelDetails);
  $('#page-details').addEventListener('cancel', (event) => { event.preventDefault(); cancelDetails(); });
  $('#details-button').addEventListener('click', () => {
    detailsBefore = { title: draft.title, description: draft.description };
    $('#page-title').value = draft.title; $('#page-description').value = draft.description;
    openDialog($('#page-details'));
  });
  $('#page-details form').addEventListener('submit', (event) => { event.preventDefault(); closeDialog($('#page-details')); });
  $('#page-title').addEventListener('input', (event) => { draft.title = event.target.value; markDirty(); });
  $('#page-description').addEventListener('input', (event) => { draft.description = event.target.value; markDirty(); });
  $('#cancel').addEventListener('click', () => { if (canLeave()) show(route, false); });
  $('#save').addEventListener('click', () => {
    // Commit all fields together to this tab's demo model, never a real site.
    const next = { ...draft, html: instances.has('page') ? cleanHTML(instances.get('page').editor.getHTML()) : null };
    shared = cleanHTML(instances.get('shared').editor.getHTML());
    pages[route] = next;
    $('#saved-output').textContent = JSON.stringify({ route, title: next.title, description: next.description, html: next.html, shared }, null, 2);
    show(route, false); setStatus('Saved in this demo tab only.');
  });
  $('#sign-out').addEventListener('click', () => {
    if (!canLeave()) return;
    signedIn = false; show(route, false); $('#admin').hidden = true; $('#sign-in').hidden = false;
  });
  $('#sign-in').addEventListener('click', () => {
    signedIn = true; $('#admin').hidden = false; $('#sign-in').hidden = true; show('/', false);
  });
  window.addEventListener('beforeunload', (event) => { if (dirty) { event.preventDefault(); event.returnValue = ''; } });
  window.addEventListener('popstate', () => {
    if (!canLeave()) { history.pushState({}, '', '?page=' + encodeURIComponent(route) + (editing ? '&edit=1' : '')); return; }
    const query = new URLSearchParams(location.search); navigate(query.get('page'), query.get('edit') === '1', false);
  });
  $('#compare').addEventListener('click', async () => {
    const html = active ? cleanHTML(active.editor.getHTML()) : pages[route].html || shared;
    try {
      const response = await fetch('/sanitize-check', { method: 'POST', body: html });
      if (!response.ok) throw new Error('Compatibility check failed');
      const result = await response.json();
      $('#compatibility').textContent = (result.changed ? 'Current server changes this HTML:\n' : 'Current server preserves this HTML:\n') + result.html;
    } catch (error) { $('#compatibility').textContent = error.message; }
  });
  // Exposed only in this standalone experiment for the repeatable browser checks.
  window.SnipPrototype = { createEditor, heading, cleanHTML, safeLink, normalizeLink, openDialog, closeDialog };
  const query = new URLSearchParams(location.search);
  show(query.get('page'), query.get('edit') === '1', false);
})();
