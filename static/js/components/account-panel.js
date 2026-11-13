import { currentToken, currentUser, logout, regenerateToken } from '../services/user-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText, t } from '../utils/i18n.js';
import { confirmAction } from '../utils/confirm-dialog.js';

export async function initAccountPanel() {
    const panel = document.getElementById('account-panel');
    const status = document.getElementById('account-status');
    const controls = document.getElementById('token-controls');
    const token = document.getElementById('access-token');
    const reveal = document.getElementById('token-reveal');
    const copy = document.getElementById('token-copy');
    const reset = document.getElementById('token-reset');
    const retry = document.getElementById('account-retry');
    const logoutButton = document.getElementById('account-logout');
    let revision = 0;

    function setBusy(value) {
        panel.setAttribute('aria-busy', String(value));
        panel.querySelectorAll('button').forEach((button) => {
            button.disabled = value;
        });
    }

    function setTokenVisibility(visible) {
        token.type = visible ? 'text' : 'password';
        reveal.setAttribute('aria-pressed', String(visible));
        const key = visible ? 'account.hide' : 'account.show';
        reveal.dataset.i18nAriaLabel = key;
        reveal.dataset.i18nTitle = key;
        reveal.setAttribute('aria-label', t(key));
        reveal.title = t(key);
        reveal.querySelector('use').setAttribute('href', `/static/images/icons.svg#${visible ? 'eye-off' : 'eye'}`);
    }

    function clearToken() {
        revision++;
        token.value = '';
        controls.hidden = true;
        reset.hidden = true;
        setTokenVisibility(false);
    }

    function handleError(error) {
        if (error?.status === 401) {
            clearToken();
            window.location.replace('/login.html');
            return;
        }
        showNotice(error?.key, 'danger', error?.params);
    }

    async function loadAccount() {
        clearToken();
        const requestRevision = revision;
        retry.hidden = true;
        clearNotice();
        setBusy(true);
        setText(status, 'access.loading');
        try {
            const user = await currentUser();
            if (requestRevision !== revision) return;
            if (user === null) {
                window.location.replace('/login.html');
                return;
            }
            setText(status, 'access.authenticated', { username: user.username });
            const value = await currentToken();
            if (requestRevision !== revision) return;
            token.value = value;
            controls.hidden = false;
            reset.hidden = false;
        } catch (error) {
            if (requestRevision !== revision) return;
            setText(status, 'account.unavailable');
            retry.hidden = false;
            handleError(error);
        } finally {
            if (requestRevision === revision) setBusy(false);
        }
    }

    reveal.addEventListener('click', () => {
        setTokenVisibility(token.type === 'password');
    });

    copy.addEventListener('click', async () => {
        clearNotice();
        try {
            await navigator.clipboard.writeText(token.value);
            showNotice('account.copied', 'success');
        } catch {
            setTokenVisibility(true);
            token.focus();
            token.select();
            showNotice('account.copyFailed', 'warning');
        }
    });

    reset.addEventListener('click', async () => {
        const confirmed = await confirmAction({
            title: 'account.reset',
            message: 'account.resetConfirm',
            confirm: 'account.reset',
        });
        if (!confirmed) return;
        clearNotice();
        setBusy(true);
        const requestRevision = revision;
        try {
            const value = await regenerateToken();
            if (requestRevision !== revision) return;
            token.value = value;
            setTokenVisibility(false);
            showNotice('account.resetDone', 'success');
        } catch (error) {
            if (requestRevision === revision) handleError(error);
        } finally {
            if (requestRevision === revision) setBusy(false);
        }
    });

    logoutButton.addEventListener('click', async () => {
        clearNotice();
        setBusy(true);
        try {
            await logout();
            clearToken();
            window.location.replace('/');
        } catch (error) {
            handleError(error);
        } finally {
            setBusy(false);
        }
    });

    retry.addEventListener('click', loadAccount);
    window.addEventListener('pagehide', clearToken);
    window.addEventListener('pageshow', (event) => {
        if (event.persisted) loadAccount();
    });
    await loadAccount();
}
