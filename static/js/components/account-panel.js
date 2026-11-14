import { currentUser, logout } from '../services/user-service.js';
import { listTokens, issueToken, updateTokenName, disableToken } from '../services/token-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText, t } from '../utils/i18n.js';
import { confirmAction } from '../utils/confirm-dialog.js';

export async function initAccountPanel() {
    const panel = document.getElementById('account-panel');
    const status = document.getElementById('account-status');
    const content = document.getElementById('account-content');
    const controls = document.getElementById('token-controls');
    const token = document.getElementById('access-token');
    const reveal = document.getElementById('token-reveal');
    const form = document.getElementById('token-create-form');
    const list = document.getElementById('tokens-list');
    const retry = document.getElementById('account-retry');
    let revision = 0;

    function setBusy(value) {
        panel.setAttribute('aria-busy', String(value));
        panel.querySelectorAll('button, input, select').forEach((control) => {
            control.disabled = value || control.dataset.unavailable === 'true';
        });
    }

    function visibility(visible) {
        token.type = visible ? 'text' : 'password';
        reveal.setAttribute('aria-pressed', String(visible));
        const key = visible ? 'account.hide' : 'account.show';
        reveal.dataset.i18nAriaLabel = key;
        reveal.setAttribute('aria-label', t(key));
        reveal.querySelector('use').setAttribute('href', `/static/images/icons.svg#${visible ? 'eye-off' : 'eye'}`);
    }

    function clearToken() {
        token.value = '';
        controls.hidden = true;
        visibility(false);
    }

    function handleError(error) {
        if (error?.status === 401) {
            clearToken();
            content.hidden = true;
            window.location.replace('/login.html');
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
            label.className = 'form-label';
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
                const description = document.createElement('dd');
                const time = document.createElement('time');
                time.dateTime = value;
                time.textContent = new Date(value).toISOString().replace('T', ' ').replace(/\.\d+Z$/, ' UTC');
                description.append(time);
                details.append(translated('dt', key), description);
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
            row.append(edit, badge, details, revoke);
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
            if (user === null) { clearToken(); content.hidden = true; window.location.replace('/login.html'); return; }
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

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const name = document.getElementById('token-name').value.trim();
        if (!name) return;
        if (token.value) { showNotice('account.saveFirst', 'warning'); token.focus(); return; }
        const days = Number(document.getElementById('token-lifetime').value);
        const current = revision;
        run(async () => {
            const result = await issueToken(name, days);
            if (current !== revision) return;
            token.value = result.access_token;
            controls.hidden = false;
            visibility(false);
            form.reset();
            token.focus();
            showNotice('account.createdNotice', 'success');
        });
    });
    reveal.addEventListener('click', () => visibility(token.type === 'password'));
    document.getElementById('token-copy').addEventListener('click', async () => {
        clearNotice();
        try {
            await navigator.clipboard.writeText(token.value);
            showNotice('account.copied', 'success');
        } catch {
            visibility(true);
            token.focus();
            token.select();
            showNotice('account.copyFailed', 'warning');
        }
    });
    document.getElementById('token-dismiss').addEventListener('click', clearToken);
    document.getElementById('account-logout').addEventListener('click', async () => {
        setBusy(true);
        try {
            await logout();
            revision++;
            clearToken();
            content.hidden = true;
            window.location.replace('/');
        } catch (error) { handleError(error); }
        finally { setBusy(false); }
    });
    retry.addEventListener('click', loadAccount);
    window.addEventListener('pagehide', () => { revision++; clearToken(); });
    window.addEventListener('pageshow', (event) => { if (event.persisted) loadAccount(); });
    await loadAccount();
}
