(() => {
    let mode = 'system';
    try {
        const saved = localStorage.getItem('agentmesh.theme');
        if (['system', 'light', 'dark'].includes(saved)) {
            mode = saved;
        }
    } catch {
        // Keep system theme detection available when storage is blocked.
    }

    // Apply the theme before styles load to avoid flashing the wrong background.
    const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    document.documentElement.dataset.themeMode = mode;
    document.documentElement.dataset.bsTheme = mode === 'system' ? systemTheme : mode;
})();
