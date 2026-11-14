// Small, local demonstrations. All examples remain readable without JavaScript.
(function () {
  document.querySelectorAll('[data-tabs]').forEach(function (group) {
    var list = group.querySelector('[data-tab-list]');
    var tabs = Array.from(list.querySelectorAll('[data-tab]'));
    var panels = tabs.map(function (tab) {
      return document.getElementById(tab.getAttribute('href').slice(1));
    });
    if (panels.some(function (panel) { return !panel; })) return;
    var narrow = window.matchMedia('(max-width: 520px)');
    var vertical;
    list.setAttribute('role', 'tablist');
    function orientation() {
      vertical = group.dataset.tabOrientation === 'vertical' && !narrow.matches;
      list.setAttribute('aria-orientation', vertical ? 'vertical' : 'horizontal');
    }
    orientation();
    narrow.addEventListener('change', orientation);
    function select(index) {
      tabs.forEach(function (tab, current) {
        var active = current === index;
        tab.setAttribute('aria-selected', String(active));
        tab.tabIndex = active ? 0 : -1;
        panels[current].hidden = !active;
      });
    }
    tabs.forEach(function (tab, index) {
      tab.setAttribute('role', 'tab');
      tab.setAttribute('aria-controls', panels[index].id);
      panels[index].setAttribute('role', 'tabpanel');
      panels[index].tabIndex = 0;
      tab.addEventListener('click', function (event) {
        event.preventDefault();
        select(index);
      });
      tab.addEventListener('keydown', function (event) {
        var next;
        if (event.key === (vertical ? 'ArrowDown' : 'ArrowRight')) next = (index + 1) % tabs.length;
        if (event.key === (vertical ? 'ArrowUp' : 'ArrowLeft')) next = (index + tabs.length - 1) % tabs.length;
        if (event.key === 'Home') next = 0;
        if (event.key === 'End') next = tabs.length - 1;
        if (next !== undefined) {
          event.preventDefault();
          select(next);
          tabs[next].focus();
        }
        if (event.key === ' ') {
          event.preventDefault();
          select(index);
        }
      });
    });
    // An ordinary fragment link can still open its corresponding panel.
    var requested = panels.findIndex(function (panel) { return '#' + panel.id === location.hash; });
    select(requested < 0 ? 0 : requested);
  });

  var motion = window.matchMedia('(prefers-reduced-motion: reduce)');
  var htmlDemo = document.getElementById('example-html');
  var source = document.querySelector('[data-demo-source]');
  var replay = document.querySelector('[data-demo-replay]');
  var preview = document.querySelector('.example-page');
  var typingFrame;
  var typingCopy;
  var played = false;
  var previewAnimation;

  function refreshPreview() {
    if (previewAnimation) previewAnimation.cancel();
    if (!motion.matches && preview.animate) {
      previewAnimation = preview.animate([
        { opacity: 0.65, filter: 'blur(1px)' },
        { opacity: 1, filter: 'blur(0)' }
      ], { duration: 200, easing: 'cubic-bezier(.16,1,.3,1)' });
    }
  }

  function finishTyping(refresh) {
    cancelAnimationFrame(typingFrame);
    typingFrame = null;
    if (typingCopy) typingCopy.remove();
    typingCopy = null;
    htmlDemo.classList.remove('demo-typing');
    replay.textContent = 'Replay';
    replay.setAttribute('aria-label', 'Replay the HTML typing demo');
    if (refresh) refreshPreview();
  }

  function typeExample() {
    if (motion.matches || htmlDemo.hidden || document.hidden) return;
    finishTyping(false);
    if (previewAnimation) previewAnimation.cancel();
    played = true;
    typingCopy = source.cloneNode(true);
    typingCopy.removeAttribute('data-demo-source');
    typingCopy.className = 'typing-copy';
    typingCopy.setAttribute('aria-hidden', 'true');
    var walker = document.createTreeWalker(typingCopy, NodeFilter.SHOW_TEXT);
    var fragments = [];
    var node;
    var length = 0;
    while ((node = walker.nextNode())) {
      fragments.push({ node: node, text: node.data, start: length });
      length += node.data.length;
      node.data = '';
    }
    var caret = document.createElement('span');
    caret.className = 'typing-caret';
    typingCopy.appendChild(caret);
    source.after(typingCopy);
    htmlDemo.classList.add('demo-typing');
    replay.textContent = 'Skip typing';
    replay.setAttribute('aria-label', 'Skip typing and show the finished example');
    var started;
    function frame(now) {
      if (htmlDemo.hidden || document.hidden) return finishTyping(false);
      if (started === undefined) started = now;
      // A bounded, one-shot sequence; the original code preserves layout and accessibility.
      var count = Math.min(length, Math.floor((now - started) / 2600 * length));
      var last;
      fragments.forEach(function (fragment) {
        var shown = Math.max(0, Math.min(fragment.text.length, count - fragment.start));
        fragment.node.data = fragment.text.slice(0, shown);
        if (shown) last = fragment.node;
      });
      if (last) last.after(caret);
      if (count === length) return finishTyping(true);
      typingFrame = requestAnimationFrame(frame);
    }
    typingFrame = requestAnimationFrame(frame);
  }

  if (source && replay && htmlDemo && preview) {
    replay.hidden = motion.matches;
    replay.setAttribute('aria-label', 'Replay the HTML typing demo');
    replay.addEventListener('click', function () {
      if (typingCopy) finishTyping(true);
      else typeExample();
    });
    source.parentElement.addEventListener('focus', function () { finishTyping(false); });
    preview.addEventListener('focusin', function () { finishTyping(false); });
    document.addEventListener('visibilitychange', function () {
      if (document.hidden) finishTyping(false);
    });
    motion.addEventListener('change', function () {
      replay.hidden = motion.matches;
      if (motion.matches) {
        finishTyping(false);
        if (previewAnimation) previewAnimation.cancel();
      }
    });
    if ('IntersectionObserver' in window) {
      var observer = new IntersectionObserver(function (entries) {
        entries.forEach(function (entry) {
          if (entry.intersectionRatio >= 0.25 && !played) typeExample();
          else if (!entry.isIntersecting && typingCopy) finishTyping(false);
        });
      }, { threshold: 0.25 });
      observer.observe(htmlDemo);
    }
  }

  var examples = {
    home: { title: 'Hello, web.', description: 'A small site, simply made.', file: 'site/index.html', path: '' },
    about: { title: 'About our site.', description: 'A little HTML goes a long way.', file: 'site/about.html', path: '/about' },
    contact: { title: 'Get in touch.', description: 'We would love to hear from you.', file: 'site/contact.html', path: '/contact' }
  };
  var previewButtons = document.querySelectorAll('[data-example-page]');
  previewButtons.forEach(function (button) {
    button.disabled = false;
    button.addEventListener('click', function () {
      var example = examples[button.dataset.examplePage];
      if (!example) return;
      finishTyping(false);
      document.querySelector('[data-example-title]').textContent = example.title;
      document.querySelector('[data-example-source]').textContent = example.title;
      document.querySelector('[data-example-description]').textContent = example.description;
      document.querySelector('[data-example-source-description]').textContent = example.description;
      document.querySelector('[data-example-file]').textContent = example.file;
      document.querySelector('[data-example-address]').textContent = 'http://localhost:3000' + example.path;
      previewButtons.forEach(function (item) {
        if (item === button) item.setAttribute('aria-current', 'page');
        else item.removeAttribute('aria-current');
      });
      refreshPreview();
    });
  });
})();
