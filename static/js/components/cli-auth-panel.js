import { currentUser } from '../services/user-service.js';
import { pendingCLIRequest, decideCLIRequest } from '../services/token-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';
import { loginForCurrentPage } from '../utils/auth-return.js';

export async function initCLIAuthPanel() {
    const panel = document.getElementById('cli-panel');
    const status = document.getElementById('cli-status');
    const form = document.getElementById('cli-code-form');
    const details = document.getElementById('cli-request');
    const code = document.getElementById('cli-code');
    const retry = document.getElementById('cli-retry');
    let ready = false;
    let revision = 0;
    let expiresAt;
    code.value = new URLSearchParams(window.location.search).get('user_code') ?? '';

    function busy(value) {
        panel.setAttribute('aria-busy', String(value));
        panel.querySelectorAll('button, input').forEach((control) => { control.disabled = value; });
    }

    function handleError(error) {
        ready = false;
        details.hidden = true;
        form.hidden = false;
        if (error?.status === 401) { window.location.replace(loginForCurrentPage()); return; }
        retry.hidden = false;
        if (error?.status === 400) { showNotice('cli.invalid', 'danger'); }
        else { showNotice(error?.key, 'danger', error?.params); }
    }

    async function load() {
        const current = ++revision;
        busy(true);
        ready = false;
        details.hidden = true;
        retry.hidden = true;
        clearNotice();
        try {
            const user = await currentUser();
            if (current !== revision) return;
            if (!user) { window.location.replace(loginForCurrentPage()); return; }
            setText(status, 'access.authenticated', { username: user.username });
            if (!code.value.trim()) { form.hidden = false; return; }
            const request = await pendingCLIRequest(code.value.trim().toLowerCase());
            if (current !== revision) return;
            code.value = request.user_code;
            document.getElementById('cli-name').textContent = request.name;
            document.getElementById('cli-user-code').textContent = request.user_code;
            const time = document.getElementById('cli-expires');
            expiresAt = new Date(request.expires_at);
            time.dateTime = request.expires_at;
            time.textContent = expiresAt.toISOString().replace('T', ' ').replace(/\.\d+Z$/, ' UTC');
            details.hidden = false;
            form.hidden = true;
            ready = true;
        } catch (error) { if (current === revision) handleError(error); }
        finally { if (current === revision) busy(false); }
    }

    async function decide(approve) {
        if (!ready) return;
        if (expiresAt <= new Date()) { showNotice('cli.invalid', 'danger'); details.hidden = true; form.hidden = false; ready = false; return; }
        const current = revision;
        busy(true);
        clearNotice();
        try {
            await decideCLIRequest(code.value, approve);
            if (current !== revision) return;
            ready = false;
            details.hidden = true;
            form.hidden = true;
            setText(status, approve ? 'cli.authorized' : 'cli.denied');
            showNotice(approve ? 'cli.authorized' : 'cli.denied', approve ? 'success' : 'info');
        } catch (error) { if (current === revision) handleError(error); }
        finally { if (current === revision) busy(false); }
    }

    form.addEventListener('submit', (event) => {
        event.preventDefault();
        const target = new URL(window.location.href);
        target.searchParams.set('user_code', code.value.trim().toLowerCase());
        window.history.replaceState(null, '', target.pathname + target.search);
        load();
    });
    document.getElementById('cli-approve').addEventListener('click', () => decide(true));
    document.getElementById('cli-deny').addEventListener('click', () => decide(false));
    retry.addEventListener('click', load);
    window.addEventListener('pagehide', () => { revision++; ready = false; details.hidden = true; });
    window.addEventListener('pageshow', (event) => { if (event.persisted) load(); });
    await load();
}
