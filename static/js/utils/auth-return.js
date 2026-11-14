// Only the CLI authorization page is allowed as a post-login destination.
export function authReturnURL() {
    const value = new URLSearchParams(window.location.search).get('next');
    if (!value) return '/account.html';
    try {
        const target = new URL(value, window.location.origin);
        if (target.origin === window.location.origin && target.pathname === '/cli-login.html') {
            return target.pathname + target.search;
        }
    } catch {
        // Ignore malformed return URLs.
    }
    return '/account.html';
}

export function loginForCurrentPage() {
    return `/login.html?next=${encodeURIComponent(window.location.pathname + window.location.search)}`;
}
