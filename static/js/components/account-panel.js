import { currentUser, invalidateCurrentUser, logout } from '../services/user-service.js';
import { listTokens, renameToken, revokeToken } from '../services/token-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { renderTranslations, setText } from '../utils/i18n.js';
import { setControlsBusy } from '../utils/form.js';
import { confirmAction } from '../utils/confirm-dialog.js';
import { formatDateTime } from '../utils/time.js';
import { navigate } from '../utils/router.js';

export function initAccountPanel({ signal }) {
    const panel = document.getElementById('account-panel');
    const status = document.getElementById('account-status');
    const content = document.getElementById('account-content');
    const list = document.getElementById('tokens-list');
    const retry = document.getElementById('account-retry');
    const template = document.getElementById('token-item-template');
    const setBusy = (value) => setControlsBusy(panel, value);
    let accountID;
    let revision = 0;
    signal.addEventListener('abort', () => { revision++; }, { once: true });

    function handleError(error) {
        if (error?.status === 401) {
            invalidateCurrentUser();
            content.hidden = true;
            navigate('/?page=login', { replace: true });
            return;
        }
        showNotice(error?.key, 'danger', error?.params);
    }

    function renderTokens(tokens, savedTokenID) {
        // Preserve edits across refreshes, except for the name just saved.
        const drafts = new Map();
        list.querySelectorAll('.token-entry').forEach((row) => {
            const input = row.querySelector('input');
            if (row.dataset.tokenId !== savedTokenID && input.value !== input.defaultValue) {
                drafts.set(row.dataset.tokenId, input.value);
            }
        });
        list.replaceChildren();
        document.getElementById('tokens-empty').hidden = tokens.length !== 0;
        tokens.forEach((item) => {
            const row = template.content.firstElementChild.cloneNode(true);
            row.dataset.tokenId = item.id;
            const edit = row.querySelector('form');
            const label = row.querySelector('label');
            label.htmlFor = `name-${item.id}`;
            const input = row.querySelector('input');
            input.id = label.htmlFor;
            input.defaultValue = item.name;
            input.value = drafts.get(item.id) ?? item.name;
            edit.addEventListener('submit', (event) => {
                event.preventDefault();
                run(() => renameToken(item.id, input.value.trim()), 'account.saved', item.id);
            });
            const badge = row.querySelector('.token-badge');
            setText(badge, `account.status.${item.status}`);
            badge.className = `token-badge token-badge-${item.status}`;
            row.querySelectorAll('time').forEach((time) => {
                const value = item[time.dataset.field];
                time.dateTime = value;
                time.textContent = formatDateTime(value);
            });
            const revoke = row.querySelector('.token-revoke');
            if (item.status !== 'active') { revoke.disabled = true; revoke.dataset.unavailable = 'true'; }
            revoke.addEventListener('click', async () => {
                const confirmed = await confirmAction({ title: 'account.revoke', message: 'account.revokeConfirm', confirm: 'account.revoke', signal });
                if (!confirmed || signal.aborted) return;
                run(() => revokeToken(item.id), 'account.revoked');
            });
            renderTranslations(row);
            list.append(row);
        });
    }

    async function run(action, successKey, savedTokenID) {
        const current = revision;
        clearNotice();
        setBusy(true);
        try {
            await action();
            if (current !== revision) return;
            const tokens = await listTokens();
            if (current !== revision) return;
            renderTokens(tokens, savedTokenID);
            showNotice(successKey, 'success');
            retry.hidden = true;
        } catch (error) {
            if (current !== revision) return;
            retry.hidden = false;
            handleError(error);
        } finally {
            if (current === revision) setBusy(false);
        }
    }

    async function loadAccount(refresh = false) {
        const current = ++revision;
        clearNotice();
        retry.hidden = true;
        setBusy(true);
        setText(status, 'access.loading');
        try {
            const user = await currentUser({ refresh });
            if (current !== revision) return;
            if (user === null) { content.hidden = true; navigate('/?page=login', { replace: true }); return; }
            if (accountID !== user.id) {
                // Never carry token drafts into a different browser session's account.
                list.replaceChildren();
                content.hidden = true;
                accountID = user.id;
            }
            setText(status, 'access.authenticated', { username: user.username });
            const tokens = await listTokens();
            if (current !== revision) return;
            renderTokens(tokens);
            content.hidden = false;
        } catch (error) {
            if (current !== revision) return;
            setText(status, 'account.unavailable');
            retry.hidden = false;
            handleError(error);
        } finally {
            if (current === revision) setBusy(false);
        }
    }

    document.getElementById('account-logout').addEventListener('click', async () => {
        setBusy(true);
        try {
            await logout();
            if (signal.aborted) return;
            content.hidden = true;
            navigate('/', { replace: true });
        } catch (error) { if (!signal.aborted) handleError(error); }
        finally { if (!signal.aborted) setBusy(false); }
    });
    retry.addEventListener('click', () => loadAccount(true));
    window.addEventListener('focus', () => {
        if (panel.getAttribute('aria-busy') !== 'true') loadAccount(true);
    }, { signal });
    window.addEventListener('pagehide', () => { revision++; }, { signal });
    window.addEventListener('pageshow', (event) => { if (event.persisted) loadAccount(true); }, { signal });
    loadAccount();
}
