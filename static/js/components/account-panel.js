import { currentUser, logout } from '../services/user-service.js';
import { listTokens, updateTokenName, disableToken } from '../services/token-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';
import { confirmAction } from '../utils/confirm-dialog.js';
import { formatDateTime } from '../utils/time.js';

export async function initAccountPanel() {
    const panel = document.getElementById('account-panel');
    const status = document.getElementById('account-status');
    const content = document.getElementById('account-content');
    const list = document.getElementById('tokens-list');
    const retry = document.getElementById('account-retry');
    let revision = 0;

    function setBusy(value) {
        panel.setAttribute('aria-busy', String(value));
        panel.querySelectorAll('button, input').forEach((control) => {
            control.disabled = value || control.dataset.unavailable === 'true';
        });
    }

    function handleError(error) {
        if (error?.status === 401) {
            content.hidden = true;
            window.location.replace('/?page=login');
            return;
        }
        showNotice(error?.key, 'danger', error?.params);
    }

    function translated(tag, key) {
        const element = document.createElement(tag);
        setText(element, key);
        return element;
    }

    function renderTokens(tokens) {
        list.replaceChildren();
        document.getElementById('tokens-empty').hidden = tokens.length !== 0;
        tokens.forEach((item) => {
            const row = document.createElement('li');
            row.className = 'token-entry';
            const edit = document.createElement('form');
            edit.className = 'token-edit-form';
            const label = translated('label', 'account.name');
            label.htmlFor = `name-${item.id}`;
            label.className = 'visually-hidden';
            const input = document.createElement('input');
            input.id = label.htmlFor;
            input.className = 'form-control';
            input.value = item.name;
            input.required = true;
            input.maxLength = 100;
            input.autocomplete = 'off';
            const save = translated('button', 'account.save');
            save.className = 'btn action-secondary';
            save.type = 'submit';
            edit.append(label, input, save);
            edit.addEventListener('submit', (event) => {
                event.preventDefault();
                const name = input.value.trim();
                if (!name) { input.focus(); return; }
                run(async () => {
                    await updateTokenName(item.id, name);
                    showNotice('account.saved', 'success');
                });
            });
            const badge = translated('span', `account.status.${item.status}`);
            badge.className = `token-badge token-badge-${item.status}`;
            const details = document.createElement('dl');
            details.className = 'token-metadata';
            for (const [key, value] of [['account.createdAt', item.created_at], ['account.expires', item.expires_at]]) {
                const group = document.createElement('div');
                const description = document.createElement('dd');
                const time = document.createElement('time');
                time.dateTime = value;
                time.textContent = formatDateTime(value);
                description.append(time);
                group.append(translated('dt', key), description);
                details.append(group);
            }
            const revoke = translated('button', 'account.revoke');
            revoke.className = 'btn action-secondary token-revoke';
            revoke.type = 'button';
            if (item.status !== 'active') { revoke.disabled = true; revoke.dataset.unavailable = 'true'; }
            revoke.addEventListener('click', async () => {
                const confirmed = await confirmAction({ title: 'account.revoke', message: 'account.revokeConfirm', confirm: 'account.revoke' });
                if (!confirmed) return;
                run(async () => {
                    await disableToken(item.id);
                    showNotice('account.revoked', 'success');
                });
            });
            row.append(edit, revoke, badge, details);
            list.append(row);
        });
    }

    async function run(action) {
        const current = revision;
        clearNotice();
        setBusy(true);
        try {
            await action();
            if (current !== revision) return;
            const tokens = await listTokens();
            if (current !== revision) return;
            renderTokens(tokens);
            retry.hidden = true;
        } catch (error) {
            if (current !== revision) return;
            retry.hidden = false;
            handleError(error);
        } finally {
            if (current === revision) setBusy(false);
        }
    }

    async function loadAccount() {
        const current = ++revision;
        clearNotice();
        retry.hidden = true;
        setBusy(true);
        setText(status, 'access.loading');
        try {
            const user = await currentUser();
            if (current !== revision) return;
            if (user === null) { content.hidden = true; window.location.replace('/?page=login'); return; }
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
            revision++;
            content.hidden = true;
            window.location.replace('/');
        } catch (error) { handleError(error); }
        finally { setBusy(false); }
    });
    retry.addEventListener('click', loadAccount);
    window.addEventListener('pagehide', () => { revision++; });
    window.addEventListener('pageshow', (event) => { if (event.persisted) loadAccount(); });
    await loadAccount();
}
