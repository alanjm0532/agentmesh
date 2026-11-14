import { currentUser, invalidateCurrentUser } from '../services/user-service.js';
import { pendingCLIRequest, decideCLIRequest } from '../services/token-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';
import { loginForCurrentPage } from '../utils/auth-return.js';
import { formatDateTime } from '../utils/time.js';
import { navigate } from '../utils/router.js';

export function initCLIAuthPanel({ signal }) {
    const panel = document.getElementById('cli-panel');
    const status = document.getElementById('cli-status');
    const form = document.getElementById('cli-code-form');
    const details = document.getElementById('cli-request');
    const code = document.getElementById('cli-code');
    const retry = document.getElementById('cli-retry');
    let ready = false;
    let revision = 0;
    let expiresAt;
    let finished = false;
    signal.addEventListener('abort', () => { revision++; ready = false; }, { once: true });
    code.value = new URLSearchParams(window.location.search).get('user_code') ?? '';

    function busy(value) {
        panel.setAttribute('aria-busy', String(value));
        panel.querySelectorAll('button, input').forEach((control) => { control.disabled = value; });
    }

    function handleError(error) {
        ready = false;
        details.hidden = true;
        form.hidden = false;
        if (error?.status === 401) { invalidateCurrentUser(); navigate(loginForCurrentPage(), { replace: true }); return; }
        retry.hidden = false;
        if (error?.status === 400) { showNotice('cli.invalid', 'danger'); }
        else { showNotice(error?.key, 'danger', error?.params); }
    }

    async function load(refresh = false) {
        const current = ++revision;
        busy(true);
        ready = false;
        details.hidden = true;
        retry.hidden = true;
        clearNotice();
        try {
            const user = await currentUser({ refresh });
            if (current !== revision) return;
            if (!user) { navigate(loginForCurrentPage(), { replace: true }); return; }
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
            time.textContent = formatDateTime(request.expires_at);
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
            finished = true;
            if (approve) { navigate('/', { replace: true }); return; }
            setText(status, 'cli.denied');
            showNotice('cli.denied', 'info');
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
    retry.addEventListener('click', () => load(true));
    window.addEventListener('focus', () => {
        if (!finished && panel.getAttribute('aria-busy') !== 'true') load(true);
    }, { signal });
    window.addEventListener('pagehide', () => { revision++; ready = false; details.hidden = true; }, { signal });
    window.addEventListener('pageshow', (event) => { if (event.persisted && !finished) load(true); }, { signal });
    load();
}
