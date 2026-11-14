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
      group.dispatchEvent(new CustomEvent('demo:select', { bubbles: true }));
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
  var sequences = [];

  function refreshPreview(preview) {
    if (!motion.matches && preview.animate) {
      preview.getAnimations().forEach(function (animation) { animation.cancel(); });
      preview.animate([
        { opacity: 0.65, filter: 'blur(1px)' },
        { opacity: 1, filter: 'blur(0)' }
      ], { duration: 200, easing: 'cubic-bezier(.16,1,.3,1)' });
    }
  }

  // Each view tells the next part of the same story, once, with explicit replay.
  // Completing an interrupted sequence leaves its final, readable HTML in place.
  function sequence(id, duration, scene) {
    var panel = document.getElementById(id);
    if (!panel) return;
    var replay = panel.querySelector('[data-demo-replay]');
    var frameId;
    var running = false;
    var played = false;
    var label = panel.getAttribute('aria-labelledby');
    var name = document.getElementById(label).textContent.trim();
    function finish(animate) {
      if (!running) return;
      cancelAnimationFrame(frameId);
      running = false;
      panel.classList.remove('demo-playing');
      scene.finish(animate);
      replay.textContent = 'Replay';
      replay.setAttribute('aria-label', 'Replay the ' + name + ' demo');
    }
    function play() {
      if (motion.matches || panel.hidden || document.hidden) return;
      finish(false);
      played = true;
      running = true;
      panel.classList.add('demo-playing');
      scene.start();
      replay.textContent = id === 'example-html' ? 'Skip typing' : 'Skip demo';
      replay.setAttribute('aria-label', 'Skip to the finished ' + name + ' example');
      var started;
      function frame(now) {
        if (panel.hidden || document.hidden) return finish(false);
        if (started === undefined) started = now;
        var elapsed = now - started;
        scene.frame(Math.min(elapsed, duration));
        if (elapsed >= duration) return finish(true);
        frameId = requestAnimationFrame(frame);
      }
      frameId = requestAnimationFrame(frame);
    }
    function enter() {
      var rect = panel.getBoundingClientRect();
      var visible = Math.min(rect.bottom, innerHeight) - Math.max(rect.top, 0);
      if (!played && !panel.hidden && visible >= Math.min(rect.height, innerHeight) * 0.25) play();
    }
    replay.hidden = motion.matches;
    replay.setAttribute('aria-label', 'Replay the ' + name + ' demo');
    replay.addEventListener('click', function () {
      if (running) finish(true);
      else play();
    });
    panel.querySelectorAll('pre,.example-page').forEach(function (element) {
      element.addEventListener('focusin', function () { finish(false); });
    });
    var controller = { panel: panel, finish: finish, enter: enter, replay: replay };
    sequences.push(controller);
    if ('IntersectionObserver' in window) {
      var observer = new IntersectionObserver(function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) enter();
          else finish(false);
        });
      }, { threshold: [0, 0.25] });
      observer.observe(panel);
    } else enter();
    return controller;
  }

  var htmlDemo = document.getElementById('example-html');
  var source = document.querySelector('[data-demo-source]');
  var preview = htmlDemo && htmlDemo.querySelector('.example-page');
  var typingCopy;
  var fragments;
  var caret;
  var length;
  var htmlSequence;
  if (htmlDemo && source && preview) {
    htmlSequence = sequence('example-html', 2600, {
      start: function () {
        typingCopy = source.cloneNode(true);
        // The original remains the accessible source; the overlay is visual only.
        typingCopy.querySelectorAll('*').forEach(function (element) {
          Array.from(element.attributes).forEach(function (attribute) {
            if (attribute.name.indexOf('data-') === 0) element.removeAttribute(attribute.name);
          });
        });
        typingCopy.removeAttribute('data-demo-source');
        typingCopy.className = 'typing-copy';
        typingCopy.setAttribute('aria-hidden', 'true');
        var walker = document.createTreeWalker(typingCopy, NodeFilter.SHOW_TEXT);
        fragments = [];
        length = 0;
        var node;
        while ((node = walker.nextNode())) {
          fragments.push({ node: node, text: node.data, start: length });
          length += node.data.length;
          node.data = '';
        }
        caret = document.createElement('span');
        caret.className = 'typing-caret';
        typingCopy.appendChild(caret);
        source.after(typingCopy);
        htmlDemo.classList.add('demo-typing');
      },
      frame: function (elapsed) {
        var count = Math.floor(elapsed / 2600 * length);
        var last;
        fragments.forEach(function (fragment) {
          var shown = Math.max(0, Math.min(fragment.text.length, count - fragment.start));
          fragment.node.data = fragment.text.slice(0, shown);
          if (shown) last = fragment.node;
        });
        if (last) last.after(caret);
      },
      finish: function (animate) {
        if (typingCopy) typingCopy.remove();
        typingCopy = null;
        htmlDemo.classList.remove('demo-typing');
        if (animate) refreshPreview(preview);
      }
    });
  }

  var editTitle = document.querySelector('[data-edit-title]');
  var localTitle = document.querySelector('[data-develop-title]');
  var localPreview = document.querySelector('[data-develop-preview]');
  var localStatus = document.querySelector('[data-develop-status]');
  var editCopy;
  var saved;
  if (editTitle && localTitle && localPreview && localStatus) {
    sequence('example-develop', 3000, {
      start: function () {
        saved = false;
        editCopy = document.createElement('span');
        editCopy.className = 'edit-copy';
        editCopy.setAttribute('aria-hidden', 'true');
        editTitle.appendChild(editCopy);
        editTitle.classList.add('edit-original');
        editCopy.textContent = 'Hello, web.';
        localTitle.textContent = 'Hello, web.';
        localStatus.textContent = 'Editing site/index.html';
      },
      frame: function (elapsed) {
        if (elapsed < 500) return;
        var count = Math.min(13, Math.floor((elapsed - 500) / 1100 * 13));
        editCopy.textContent = 'Hello, again.'.slice(0, count);
        if (elapsed >= 1800 && !saved) {
          saved = true;
          editTitle.classList.remove('edit-original');
          editCopy.remove();
          localStatus.textContent = 'Saved · reloading browser…';
        }
        if (elapsed >= 2200 && localTitle.textContent !== 'Hello, again.') {
          localTitle.textContent = 'Hello, again.';
          localStatus.textContent = 'Saved · browser reloaded';
          refreshPreview(localPreview);
        }
      },
      finish: function () {
        if (editCopy) editCopy.remove();
        editTitle.classList.remove('edit-original');
        localTitle.textContent = 'Hello, again.';
        localStatus.textContent = 'Saved · browser reloaded';
      }
    });
  }

  var publishPanel = document.getElementById('example-publish');
  if (publishPanel) {
    var publishSteps = Array.from(publishPanel.querySelectorAll('[data-publish-step]'));
    var publishStatus = publishPanel.querySelector('[data-publish-status]');
    var publishAddress = publishPanel.querySelector('[data-publish-address]');
    var publishedPreview = publishPanel.querySelector('[data-publish-preview]');
    var published;
    sequence('example-publish', 3400, {
      start: function () {
        published = false;
        publishAddress.textContent = 'http://localhost:3000';
        publishStatus.textContent = 'Set up your Ubuntu server over SSH';
      },
      frame: function (elapsed) {
        var step = elapsed < 900 ? 0 : elapsed < 1700 ? 1 : 2;
        publishSteps.forEach(function (element, index) { element.classList.toggle('is-current', index <= step); });
        publishStatus.textContent = elapsed < 900 ? 'Set up your Ubuntu server over SSH' : elapsed < 1700 ? 'Commit the updated page' : 'Deploy committed files to your server';
        if (elapsed >= 2700 && !published) {
          published = true;
          publishPanel.classList.remove('demo-playing');
          publishAddress.textContent = 'https://example.com';
          refreshPreview(publishedPreview);
        }
        if (published) publishStatus.textContent = 'Published · automatic HTTPS';
      },
      finish: function () {
        publishSteps.forEach(function (element) { element.classList.remove('is-current'); });
        publishAddress.textContent = 'https://example.com';
        publishStatus.textContent = 'Published · automatic HTTPS';
      }
    });
  }

  document.addEventListener('demo:select', function () {
    sequences.forEach(function (controller) {
      if (controller.panel.hidden) controller.finish(false);
      else controller.enter();
    });
  });
  document.querySelectorAll('[data-demo-next]').forEach(function (link) {
    link.addEventListener('click', function (event) {
      var tab = document.querySelector('[data-tab][href="' + link.getAttribute('href') + '"]');
      if (!tab) return;
      event.preventDefault();
      tab.click();
      tab.focus();
    });
  });
  document.addEventListener('visibilitychange', function () {
    if (document.hidden) sequences.forEach(function (controller) { controller.finish(false); });
  });
  motion.addEventListener('change', function () {
    sequences.forEach(function (controller) {
      controller.replay.hidden = motion.matches;
      if (motion.matches) {
        controller.finish(false);
        controller.panel.getAnimations({ subtree: true }).forEach(function (animation) { animation.cancel(); });
      }
    });
  });

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
      if (htmlSequence) htmlSequence.finish(false);
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
      refreshPreview(preview);
    });
  });
})();
