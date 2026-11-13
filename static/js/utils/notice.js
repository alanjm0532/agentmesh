import { hasTranslation, setText } from './i18n.js';

export function showNotice(key, tone = 'danger', params = {}) {
    const notice = document.getElementById('notice');
    notice.setAttribute('role', tone === 'danger' ? 'alert' : 'status');
    notice.setAttribute('aria-live', tone === 'danger' ? 'assertive' : 'polite');
    notice.className = `alert alert-${tone}`;
    setText(notice, hasTranslation(key) ? key : 'error.unexpected', params);
}

export function clearNotice() {
    const notice = document.getElementById('notice');
    notice.className = 'alert d-none';
    notice.textContent = '';
    delete notice.dataset.i18n;
    delete notice.dataset.i18nParams;
}
