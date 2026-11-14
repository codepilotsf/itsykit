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

  var examples = {
    home: { title: 'Hello, web.', file: 'site/index.html', path: '' },
    about: { title: 'About our site.', file: 'site/about.html', path: '/about' },
    contact: { title: 'Get in touch.', file: 'site/contact.html', path: '/contact' }
  };
  var previewButtons = document.querySelectorAll('[data-example-page]');
  previewButtons.forEach(function (button) {
    button.disabled = false;
    button.addEventListener('click', function () {
      var example = examples[button.dataset.examplePage];
      if (!example) return;
      document.querySelector('[data-example-title]').textContent = example.title;
      document.querySelector('[data-example-source]').textContent = example.title;
      document.querySelector('[data-example-file]').textContent = example.file;
      document.querySelector('[data-example-address]').textContent = 'http://localhost:3000' + example.path;
      previewButtons.forEach(function (item) {
        if (item === button) item.setAttribute('aria-current', 'page');
        else item.removeAttribute('aria-current');
      });
    });
  });
})();
