import { setText } from './i18n.js';

export function showNotice(key, tone = 'danger', params = {}) {
    const notice = document.getElementById('notice');
    notice.className = `alert alert-${tone}`;
    setText(notice, key, params);
}

export function clearNotice() {
    document.getElementById('notice').className = 'alert d-none';
}
