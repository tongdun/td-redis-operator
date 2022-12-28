const wr = type => {
  const orig = window.history[type];
  return (...args) => {
    const rv = orig.apply(this, args);
    const e = new Event(type);
    e.arguments = args;
    window.dispatchEvent(e);
    return rv;
  };
};
window.history.pushState = wr('pushState');
window.history.replaceState = wr('replaceState');

export function addHistoryListener(fn) {
  window.addEventListener('hashchange', fn);
  window.addEventListener('popstate', fn);
  window.addEventListener('pushState', fn);
  window.addEventListener('replaceState', fn);
}

export function removeHistoryListener(fn) {
  window.removeEventListener('hashchange', fn);
  window.removeEventListener('popstate', fn);
  window.removeEventListener('pushState', fn);
  window.removeEventListener('replaceState', fn);
}
