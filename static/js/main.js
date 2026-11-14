import { initAccountLink } from './components/account-link.js';
import { initAccountPanel } from './components/account-panel.js';
import { initAuthPanel } from './components/auth-panel.js';
import { initCLIAuthPanel } from './components/cli-auth-panel.js';
import { initI18n } from './utils/i18n.js';
import { initTheme } from './utils/theme.js';

// Mount only the selected page so IDs stay unique.
const pages = {
    login: { template: 'login-page', title: 'access.title', init: initAuthPanel },
    account: { template: 'account-page', title: 'account.title', init: initAccountPanel },
    'cli-login': { template: 'cli-login-page', title: 'cli.title', init: initCLIAuthPanel },
};
const pageName = new URLSearchParams(window.location.search).get('page');
const page = Object.hasOwn(pages, pageName) ? pages[pageName] : null;
if (page) {
    document.getElementById('page-content').replaceWith(
        document.getElementById(page.template).content.cloneNode(true),
    );
    document.querySelector('title').dataset.i18n = page.title;
    document.getElementById('home-link').hidden = page.template !== 'login-page';
}

initI18n();
initTheme();
if (page) {
    page.init();
} else {
    initAccountLink();
}
