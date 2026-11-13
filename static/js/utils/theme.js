const themeKey = 'agentmesh.theme';
const systemTheme = window.matchMedia('(prefers-color-scheme: dark)');

function applyTheme(mode) {
    document.documentElement.dataset.themeMode = mode;
    document.documentElement.dataset.bsTheme = mode === 'system'
        ? (systemTheme.matches ? 'dark' : 'light')
        : mode;
    document.querySelectorAll('[data-theme]').forEach((button) => {
        button.setAttribute('aria-pressed', String(button.dataset.theme === mode));
    });
}

export function initTheme() {
    applyTheme(document.documentElement.dataset.themeMode);

    document.querySelectorAll('[data-theme]').forEach((button) => {
        button.addEventListener('click', () => {
            const mode = button.dataset.theme;
            applyTheme(mode);
            try {
                localStorage.setItem(themeKey, mode);
            } catch {
                // Theme switching still works for the current page without storage.
            }
        });
    });

    systemTheme.addEventListener('change', () => {
        if (document.documentElement.dataset.themeMode === 'system') {
            applyTheme('system');
        }
    });

    window.addEventListener('storage', (event) => {
        if (event.key === themeKey) {
            applyTheme(['light', 'dark'].includes(event.newValue) ? event.newValue : 'system');
        }
    });
}
