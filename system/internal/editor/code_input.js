(() => {
  const original = document.querySelector('form[action="/admin/verify"] input[name="code"]');
  if (!original) return;
  const group = document.createElement('span');
  group.className = 'code-input';
  group.setAttribute('role', 'group');
  group.setAttribute('aria-label', 'Sign-in code');
  const boxes = Array.from({length: 6}, (_, i) => {
    const box = document.createElement('input');
    box.type = 'text';
    box.inputMode = 'numeric';
    box.autocomplete = i === 0 ? 'one-time-code' : 'off';
    box.pattern = '[0-9]';
    box.required = true;
    box.setAttribute('aria-label', `Digit ${i + 1} of 6`);
    group.append(box);
    return box;
  });
  const sync = () => { original.value = boxes.map(box => box.value).join(''); };
  const fill = (text, index) => {
    const digits = text.replace(/\D/g, '').slice(0, 6);
    if (!digits) return;
    if (digits.length === 6) { index = 0; boxes.forEach(box => { box.value = ''; }); }
    for (let n = 0; n < digits.length && index + n < 6; n++) boxes[index + n].value = digits[n];
    sync();
    boxes[Math.min(index + digits.length, 5)].focus();
  };
  boxes.forEach((box, i) => {
    box.addEventListener('focus', () => box.select());
    box.addEventListener('input', () => {
      const value = box.value;
      box.value = value.replace(/\D/g, '').slice(0, 1);
      if (value.replace(/\D/g, '')) fill(value, i); else sync();
    });
    box.addEventListener('paste', event => {
      event.preventDefault();
      const text = event.clipboardData.getData('text');
      fill(text, i);
      if (/^\d{6}$/.test(text.trim())) original.form.requestSubmit();
    });
    box.addEventListener('keydown', event => {
      if (event.key === 'Backspace' && !box.value && i > 0) {
        event.preventDefault(); boxes[i - 1].value = ''; sync(); boxes[i - 1].focus();
      } else if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
        event.preventDefault(); boxes[Math.max(0, Math.min(5, i + (event.key === 'ArrowLeft' ? -1 : 1)))].focus();
      }
    });
  });
  const saved = original.value;
  original.type = 'hidden';
  original.required = false;
  original.removeAttribute('autofocus');
  original.after(group);
  // The group supplies individual labels rather than one label for six controls.
  if (group.parentElement.tagName === 'LABEL') {
    const label = group.parentElement;
    const wrapper = document.createElement('div');
    wrapper.className = label.className;
    wrapper.append(...label.childNodes);
    label.replaceWith(wrapper);
  }
  if (saved) fill(saved, 0); else boxes[0].focus();
})();
