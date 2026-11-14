document.querySelectorAll('[data-guide-navigation]').forEach((navigation, index) => {
  const currentPath = navigation.dataset.currentGuide;

  navigation.querySelectorAll('.guide-section').forEach((guide) => {
    guide.name = `docs-guides-${index}`;
    const current = guide.dataset.guidePath === currentPath;
    guide.open = current;
    // Following a guide title opens its page; the current title can still fold it.
    guide.querySelector('summary a').addEventListener('click', (event) => {
      if (!current || location.hash || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
      event.preventDefault();
      guide.open = !guide.open;
    });
    if (current) {
      guide.dataset.current = '';
      guide.querySelector('summary a').setAttribute('aria-current', 'page');
    }
  });

  const revealCurrentGuide = () => {
    const current = navigation.querySelector('[data-current]');
    const scroller = navigation.closest('.docs-sidebar') || navigation;
    if (!current || !scroller.clientHeight) return;
    const offset = current.getBoundingClientRect().top - scroller.getBoundingClientRect().top;
    if (offset < 0 || offset > scroller.clientHeight / 2) {
      scroller.scrollTop += offset - 12;
    }
  };

  // Settle the initial selection before enabling transitions for user actions.
  revealCurrentGuide();
  navigation.dataset.ready = '';
  const mobileMenu = navigation.closest('.docs-mobile');
  if (mobileMenu) {
    mobileMenu.addEventListener('toggle', () => {
      if (mobileMenu.open) revealCurrentGuide();
    });
  }

  const chapters = [...navigation.querySelectorAll('[data-current] .guide-chapters a')]
    .map((link) => ({ link, target: document.getElementById(new URL(link.href).hash.slice(1)) }))
    .filter(({ target }) => target);

  const markChapter = () => {
    // Match native anchor positioning, including the sticky header's scroll padding.
    const padding = parseFloat(getComputedStyle(document.documentElement).scrollPaddingTop) || 0;
    let active;
    chapters.forEach((chapter) => {
      const margin = parseFloat(getComputedStyle(chapter.target).scrollMarginTop) || 0;
      if (chapter.target.getBoundingClientRect().top <= padding + margin + 1) active = chapter;
    });
    // The final chapter may be too short to reach the reading line.
    if (window.scrollY > 0 && window.scrollY + window.innerHeight >= document.documentElement.scrollHeight - 1) {
      active = chapters[chapters.length - 1];
    }

    chapters.forEach((chapter) => {
      const { link } = chapter;
      if (chapter === active) {
        link.setAttribute('aria-current', 'location');
      } else {
        link.removeAttribute('aria-current');
      }
    });
  };

  let scheduled = false;
  const scheduleChapter = () => {
    if (scheduled) return;
    scheduled = true;
    requestAnimationFrame(() => {
      scheduled = false;
      markChapter();
    });
  };

  markChapter();
  window.addEventListener('scroll', scheduleChapter, { passive: true });
  window.addEventListener('resize', scheduleChapter);
  window.addEventListener('hashchange', scheduleChapter);
  window.addEventListener('pageshow', scheduleChapter);
  window.addEventListener('load', scheduleChapter);
});
