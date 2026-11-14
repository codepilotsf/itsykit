document.querySelectorAll('[data-guide-navigation]').forEach((navigation, index) => {
  const currentPath = navigation.dataset.currentGuide;

  navigation.querySelectorAll('.guide-section').forEach((guide) => {
    guide.name = `docs-guides-${index}`;
    const current = guide.dataset.guidePath === currentPath;
    guide.open = current;
    if (current) {
      guide.dataset.current = '';
      guide.querySelector('.guide-overview').setAttribute('aria-current', 'page');
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

  const markChapter = () => {
    navigation.querySelectorAll('.guide-chapters a:not(.guide-overview)').forEach((link) => {
      const target = new URL(link.href);
      if (target.pathname === currentPath && target.hash === location.hash && location.hash) {
        link.setAttribute('aria-current', 'location');
      } else {
        link.removeAttribute('aria-current');
      }
    });
  };

  markChapter();
  window.addEventListener('hashchange', markChapter);
});
