let renderPage;

function isAppURL(url) {
    return url.origin === window.location.origin && url.pathname === '/' && !url.hash;
}

export function navigate(value, { replace = false } = {}) {
    const url = new URL(value, window.location.href);
    if (!isAppURL(url)) {
        window.location[replace ? 'replace' : 'assign'](url.href);
        return;
    }
    if (url.href === window.location.href) return;
    window.history[replace ? 'replaceState' : 'pushState'](null, '', url);
    renderPage();
}

export function initRouter(render) {
    let renderedURL;
    renderPage = (event) => {
        const url = window.location.pathname + window.location.search;
        if (url === renderedURL) return;
        renderedURL = url;
        render(event);
    };
    document.addEventListener('click', (event) => {
        if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
        const link = event.target instanceof Element ? event.target.closest('a[href]') : null;
        if (!link || link.hasAttribute('download') || (link.target && link.target !== '_self') || link.getAttribute('href').startsWith('#')) return;
        const url = new URL(link.href);
        if (!isAppURL(url)) return;
        event.preventDefault();
        navigate(url);
    });
    window.addEventListener('popstate', renderPage);
    renderPage();
}
