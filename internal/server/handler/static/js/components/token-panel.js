import { clearToken, hasToken, saveToken } from '../utils/storage.js';
import { showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';

export function initTokenPanel() {
    const form = document.getElementById('token-form');
    const status = document.getElementById('token-status');

    function renderStatus() {
        setText(status, hasToken() ? 'access.configured' : 'access.public');
    }

    form.addEventListener('submit', (event) => {
        event.preventDefault();
        try {
            saveToken(new FormData(form).get('token').trim());
            form.reset();
            renderStatus();
            showNotice('access.saved', 'success');
        } catch {
            showNotice('error.tokenSave');
        }
    });

    document.getElementById('clear-token').addEventListener('click', () => {
        try {
            clearToken();
            form.reset();
            renderStatus();
            showNotice('access.cleared', 'success');
        } catch {
            showNotice('error.tokenClear');
        }
    });

    try {
        renderStatus();
    } catch {
        showNotice('error.storage');
    }
}
