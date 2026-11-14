import { initAccountLink } from './components/account-link.js';
import { initAccountPanel } from './components/account-panel.js';
import { initAuthPanel } from './components/auth-panel.js';
import { initCLIAuthPanel } from './components/cli-auth-panel.js';
import { initI18n, renderTranslations, setText } from './utils/i18n.js';
import { initRouter } from './utils/router.js';
import { initTheme } from './utils/theme.js';

const homeTemplate = document.getElementById('page-content').cloneNode(true);
const pages = {
    login: { template: 'login-page', title: 'access.title', init: initAuthPanel },
    account: { template: 'account-page', title: 'account.title', init: initAccountPanel },
    'cli-login': { template: 'cli-login-page', title: 'cli.title', init: initCLIAuthPanel },
};
let view;
let initialRender = true;

function renderPage(event) {
    view?.abort();
    view = new AbortController();
    const pageName = new URLSearchParams(window.location.search).get('page');
    const page = Object.hasOwn(pages, pageName) ? pages[pageName] : null;
    const content = page
        ? document.getElementById(page.template).content.cloneNode(true)
        : homeTemplate.cloneNode(true);
    document.getElementById('page-content').replaceWith(content);
    setText(document.querySelector('title'), page?.title ?? 'home.title');
    document.getElementById('home-link').hidden = page?.template !== 'login-page';
    renderTranslations();

    if (!initialRender) {
        const heading = document.querySelector('main h1');
        heading.tabIndex = -1;
        heading.focus({ preventScroll: true });
        if (event?.type !== 'popstate') window.scrollTo(0, 0);
    }
    initialRender = false;
    (page?.init ?? initAccountLink)({ signal: view.signal });
}

initI18n();
initTheme();
initRouter(renderPage);
