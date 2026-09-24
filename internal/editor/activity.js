(() => {
  const csrf = document.querySelector('input[name="csrf"]')?.value;
  if (!csrf) return;
  let last = Date.now();
  let pending = false;
  let expired = false;
  const activity = async (event) => {
    if (!event.isTrusted || document.visibilityState !== 'visible' || pending || expired) return;
    const now = Date.now();
    if (now - last < 60000) return;
    last = now;
    pending = true;
    try {
      const response = await fetch('/admin/activity', {
        method: 'POST',
        credentials: 'same-origin',
        redirect: 'error',
        body: new URLSearchParams({ csrf }),
      });
      expired = response.status === 403;
    } catch {
      // Retry on a later interaction, never on an idle timer.
    } finally {
      pending = false;
    }
  };
  for (const type of ['input', 'keydown', 'pointerdown', 'scroll']) {
    document.addEventListener(type, activity, { capture: true, passive: true });
  }
})();
