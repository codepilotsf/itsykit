// Reveal the light header divider and translucency after scrolling.
(function () {
  var header = document.querySelector('.site-header');
  if (!header) return;

  function update() {
    header.classList.toggle('is-scrolled', window.scrollY > 0);
  }

  update();
  window.addEventListener('scroll', update, { passive: true });
  window.addEventListener('pageshow', update);
})();
