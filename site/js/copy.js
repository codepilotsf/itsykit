// Progressive enhancement for copy buttons. No build step, no dependencies.
(function () {
  function status(text) {
    var region = document.getElementById('copy-status');
    if (region) region.textContent = text;
  }
  function fallback(text) {
    var field = document.createElement('textarea');
    field.value = text;
    field.setAttribute('readonly', '');
    field.style.position = 'fixed';
    field.style.top = '-1000px';
    document.body.appendChild(field);
    field.select();
    try { document.execCommand('copy'); } catch (error) { /* ignore */ }
    document.body.removeChild(field);
  }
  function done(button) {
    button.classList.add('is-copied');
    button.setAttribute('aria-label', 'Copied');
    status('Copied to clipboard.');
    window.setTimeout(function () {
      button.classList.remove('is-copied');
      button.setAttribute('aria-label', 'Copy the install command');
    }, 2000);
  }
  document.addEventListener('click', function (event) {
    var button = event.target.closest('.copy-button');
    if (!button) return;
    var block = button.closest('.copy-code');
    var code = block && block.querySelector('code');
    if (!code) return;
    var text = code.textContent.trim();
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function () { done(button); }, function () {
        fallback(text);
        done(button);
      });
    } else {
      fallback(text);
      done(button);
    }
  });
})();
