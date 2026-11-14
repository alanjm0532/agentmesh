// Only the CLI authorization page is allowed as a post-login destination.
export function authReturnURL() {
    const value = new URLSearchParams(window.location.search).get('next');
    if (!value) return '/?page=account';
    try {
        const target = new URL(value, window.location.origin);
        if (target.origin === window.location.origin && target.pathname === '/' &&
            target.searchParams.get('page') === 'cli-login') {
            return target.pathname + target.search;
        }
    } catch {
        // Ignore malformed return URLs.
    }
    return '/?page=account';
}

export function loginForCurrentPage() {
    return `/?page=login&next=${encodeURIComponent(window.location.pathname + window.location.search)}`;
}
