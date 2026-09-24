/* Browser checks for the prototype, deliberately not shipped in Flea. */
document.querySelector('#run-checks').addEventListener('click', async () => {
  const output = document.querySelector('#check-results');
  const button = document.querySelector('#run-checks');
  button.disabled = true;
  const results = [];
  const { createEditor, heading, cleanHTML, safeLink } = window.FleaPrototype;
  const tick = () => new Promise((resolve) => requestAnimationFrame(resolve));
  const assert = (value, message) => { if (!value) throw new Error(message); };
  async function check(name, test) {
    const root = document.createElement('div'); root.className = 'check-fixture'; document.body.append(root);
    const editor = createEditor(root, '<p>Alpha beta</p>');
    const select = (selector = 'p') => { const range = document.createRange(); range.selectNodeContents(root.querySelector(selector)); editor.setSelection(range); };
    try { await tick(); await test(editor, root, select); results.push('PASS  ' + name); }
    catch (error) { results.push('FAIL  ' + name + ': ' + error.message); }
    finally { editor.destroy(); root.remove(); }
    output.textContent = results.join('\n');
  }
  for (const tag of ['H1', 'H2', 'H3', 'H4']) {
    await check(tag + ' conversion, undo, and redo', async (editor, root, select) => {
      select(); heading(editor, tag); await tick();
      assert(root.querySelector(tag)?.textContent === 'Alpha beta', 'heading not created');
      editor.undo(); await tick(); assert(root.querySelector('p')?.textContent === 'Alpha beta', 'undo failed');
      editor.redo(); await tick(); assert(root.querySelector(tag)?.textContent === 'Alpha beta', 'redo failed');
      select(tag); heading(editor, 'P'); await tick(); assert(root.querySelector('p'), 'normal text failed');
    });
  }
  await check('Undo availability follows Squire history events', async (editor, root, select) => {
    let history;
    editor.addEventListener('undoStateChange', (event) => { history = event.detail; });
    select(); editor.bold(); await tick();
    assert(history?.canUndo && !history.canRedo, 'edit history unavailable');
    editor.undo(); await tick();
    assert(!history.canUndo && history.canRedo, 'undo availability incorrect');
    editor.redo(); await tick();
    assert(history.canUndo && !history.canRedo, 'redo availability incorrect');
  });
  for (const tag of ['B', 'I', 'U', 'S']) {
    await check(tag + ' formatting round trip', async (editor, root, select) => {
      select(); editor.changeFormat({ tag }); await tick();
      const html = cleanHTML(editor.getHTML());
      editor.setHTML(html); assert(root.querySelector(tag)?.textContent === 'Alpha beta', 'format lost');
    });
  }
  for (const kind of ['Unordered', 'Ordered']) {
    await check(kind + ' lists, removal and undo', async (editor, root, select) => {
      select(); editor['make' + kind + 'List'](); await tick();
      assert(root.querySelector(kind === 'Ordered' ? 'ol li' : 'ul li'), 'list missing');
      editor.removeList(); await tick(); assert(!root.querySelector('li'), 'list not removed');
      editor.undo(); await tick(); assert(root.querySelector('li'), 'list undo failed');
    });
  }
  await check('Heading inside a list remains valid', async (editor, root, select) => {
    select(); editor.makeUnorderedList(); await tick(); select('li'); heading(editor, 'H2'); await tick();
    assert(root.querySelector('ul > li > h2'), 'heading escaped list item');
    const html = cleanHTML(editor.getHTML()); editor.setHTML(html);
    assert(root.querySelector('ul > li > h2'), 'list heading lost during round trip');
  });
  await check('Links and unsafe URLs', async (editor, root, select) => {
    select(); editor.makeLink('https://example.com/about'); await tick();
    assert(root.querySelector('a')?.getAttribute('href') === 'https://example.com/about', 'link missing');
    for (const url of ['javascript:alert(1)', 'data:text/html,x', '//evil.test', 'https:\\evil.test']) assert(!safeLink(url), 'unsafe URL accepted');
    for (const url of ['/about', '#contact', 'mailto:hello@example.com', 'tel:+15551234567']) assert(safeLink(url), 'safe URL rejected');
  });
  await check('Left, center and right alignment round trip', async (editor, root, select) => {
    for (const alignment of ['left', 'center', 'right']) {
      select(); editor.setTextAlignment(alignment); await tick();
      editor.setHTML(cleanHTML(editor.getHTML()));
      assert(root.querySelector('p').style.textAlign === alignment, 'alignment lost');
    }
  });
  await check('Pasted HTML removes scripts, media, colors and H5/H6', async (editor, root, select) => {
    select(); editor.insertHTML('<h2>Safe heading</h2><p style="color:red" onclick="alert(1)"><b>Bold</b> <i>italic</i></p><ul><li>Item</li></ul><h5>Extra heading</h5><img src="x" onerror="alert(1)"><video src="x"></video><script>alert(1)</script><a href="javascript:alert(1)">Bad URL</a>', true); await tick();
    const html = cleanHTML(editor.getHTML());
    assert(!/<(?:script|img|video|h5|h6)\b|onclick|onerror|color:|javascript:/i.test(html), 'unwanted markup retained');
    assert(root.querySelector('h2') && root.querySelector('b') && root.querySelector('ul li'), 'useful formatting lost');
  });
  await check('Undo and redo pasted text', async (editor, root, select) => {
    select(); editor.insertPlainText('Replacement', true); await tick();
    assert(root.textContent.includes('Replacement'), 'paste failed');
    editor.undo(); await tick(); assert(root.textContent.includes('Alpha beta'), 'paste undo failed');
    editor.redo(); await tick(); assert(root.textContent.includes('Replacement'), 'paste redo failed');
  });
  await check('Four headings preserve site typography', async (editor, root) => {
    const html = '<h1>One</h1><h2>Two</h2><h3>Three</h3><h4>Four</h4><p>Paragraph</p>';
    const reference = document.createElement('div'); reference.innerHTML = html; root.after(reference);
    try {
      editor.setHTML(html); await tick();
      for (const tag of ['h1', 'h2', 'h3', 'h4', 'p']) {
        const before = getComputedStyle(reference.querySelector(tag)), after = getComputedStyle(root.querySelector(tag));
        for (const property of ['fontSize', 'fontFamily', 'fontWeight', 'lineHeight']) assert(before[property] === after[property], tag + ' ' + property + ' changed');
      }
    } finally { reference.remove(); }
  });
  await check('Independent regions retain independent undo history', async (editor, root, select) => {
    const second = document.createElement('div'); root.after(second); const other = createEditor(second, '<p>Footer</p>');
    try {
      select(); editor.bold(); await tick(); editor.undo(); await tick();
      assert(other.getHTML().includes('Footer') && !second.querySelector('b'), 'history crossed regions');
    } finally { other.destroy(); second.remove(); }
  });
  await check('Current Go sanitizer gap is detected without changing production', async () => {
    const response = await fetch('/sanitize-check', { method: 'POST', body: '<h1>Title</h1><p style="text-align: center">Centered</p>' });
    assert(response.ok, 'comparison endpoint failed');
    const result = await response.json();
    assert(result.changed && !result.html.includes('<h1>') && !result.html.includes('style='), 'production contract changed unexpectedly');
  });
  await check('Dialog exit finishes before closing, and reopens cleanly', async () => {
    const dialog = document.createElement('dialog');
    dialog.textContent = 'Animation check'; document.body.append(dialog);
    try {
      window.FleaPrototype.openDialog(dialog);
      assert(dialog.open, 'dialog did not open');
      const closing = window.FleaPrototype.closeDialog(dialog);
      if (!matchMedia('(prefers-reduced-motion: reduce)').matches) assert(dialog.open && dialog.classList.contains('closing'), 'closed before exit animation');
      await closing;
      assert(!dialog.open && !dialog.classList.contains('closing'), 'close did not finish');
      window.FleaPrototype.openDialog(dialog);
      assert(dialog.open, 'reopen failed');
      await window.FleaPrototype.closeDialog(dialog);
    } finally { dialog.remove(); }
  });
  await check('Link protocol normalization and new-tab round trip', async (editor, root, select) => {
    const normalize = window.FleaPrototype.normalizeLink;
    assert(normalize('example.com/about') === 'https://example.com/about', 'bare domain not HTTPS');
    for (const url of ['http://example.com', 'https://example.com', '/about', './page.html', '../about', '#contact', 'mailto:hello@example.com', 'tel:+15551234567']) assert(normalize(url) === url, 'explicit URL changed');
    select(); editor.makeLink('https://example.com', { target: '_blank', rel: 'noopener noreferrer' }); await tick();
    editor.setHTML(cleanHTML(editor.getHTML()));
    let link = root.querySelector('a');
    assert(link.target === '_blank' && link.rel === 'noopener noreferrer', 'new-tab settings lost');
    select('a'); editor.makeLink('https://example.com/updated', {}); await tick();
    link = root.querySelector('a');
    assert(!link.hasAttribute('target'), 'same-tab toggle retained target');
    assert(!cleanHTML('<a href="/" target="named" rel="opener">x</a>').includes('target='), 'arbitrary target allowed');
  });
  const failures = results.filter((line) => line.startsWith('FAIL')).length;
  output.textContent = `${results.length - failures}/${results.length} passed\n\n` + results.join('\n');
  button.disabled = false;
});
