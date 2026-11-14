import { authenticate, currentUser } from '../services/user-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';
import { authReturnURL } from '../utils/auth-return.js';
import { navigate } from '../utils/router.js';

export function initAuthPanel({ signal }) {
    const form = document.getElementById('auth-form');
    const status = document.getElementById('auth-status');
    let revision = 0;
    let submitting = false;
    signal.addEventListener('abort', () => { revision++; }, { once: true });

    function setBusy(busy) {
        form.setAttribute('aria-busy', String(busy));
        form.querySelectorAll('input, button').forEach((control) => {
            control.disabled = busy;
        });
    }

    function renderUser(user) {
        const authenticated = user !== null;
        form.hidden = authenticated;
        if (authenticated) {
            navigate(authReturnURL(), { replace: true });
        }
        setText(status, authenticated ? 'access.authenticated' : 'access.signedOut', authenticated ? { username: user.username } : {});
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        clearNotice();
        const data = new FormData(form);
        const current = ++revision;
        submitting = true;
        setBusy(true);
        try {
            await authenticate({
                username: data.get('username'),
                password: data.get('password'),
            });
            if (signal.aborted || current !== revision) return;
            form.reset();
            navigate(authReturnURL(), { replace: true });
        } catch (error) {
            if (!signal.aborted && current === revision) showNotice(error?.key, 'danger', error?.params);
        } finally {
            submitting = false;
            if (!signal.aborted && current === revision) setBusy(false);
        }
    });

    async function load(refresh = false) {
        if (submitting) return;
        const current = ++revision;
        setBusy(true);
        try {
            const user = await currentUser({ refresh });
            if (signal.aborted || current !== revision) return;
            renderUser(user);
        } catch (error) {
            if (signal.aborted || current !== revision) return;
            setText(status, 'access.unavailable');
            showNotice(error?.key, 'danger', error?.params);
        } finally {
            if (!signal.aborted && current === revision) setBusy(false);
        }
    }

    window.addEventListener('focus', () => load(true), { signal });
    window.addEventListener('pageshow', (event) => { if (event.persisted) load(true); }, { signal });
    load();
}
