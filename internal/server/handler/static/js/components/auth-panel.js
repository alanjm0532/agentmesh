import { authenticate, currentUser, logout } from '../services/user-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';

export async function initAuthPanel() {
    const form = document.getElementById('auth-form');
    const status = document.getElementById('auth-status');
    const logoutButton = document.getElementById('auth-logout');

    function setBusy(busy) {
        form.setAttribute('aria-busy', String(busy));
        form.querySelectorAll('input, button').forEach((control) => {
            control.disabled = busy;
        });
        logoutButton.disabled = busy;
    }

    function renderUser(user) {
        const authenticated = user !== null;
        form.hidden = authenticated;
        logoutButton.hidden = !authenticated;
        setText(status, authenticated ? 'access.authenticated' : 'access.signedOut', authenticated ? { username: user.username } : {});
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        clearNotice();
        const data = new FormData(form);
        setBusy(true);
        try {
            await authenticate({
                username: data.get('username'),
                password: data.get('password'),
            });
            form.reset();
            window.location.replace('/');
        } catch (error) {
            showNotice(error.key, 'danger', error.params);
        } finally {
            setBusy(false);
        }
    });

    logoutButton.addEventListener('click', async () => {
        clearNotice();
        setBusy(true);
        try {
            await logout();
            renderUser(null);
            showNotice('access.loggedOut', 'success');
        } catch (error) {
            showNotice(error.key, 'danger', error.params);
        } finally {
            setBusy(false);
        }
    });

    setBusy(true);
    try {
        renderUser(await currentUser());
    } catch (error) {
        setText(status, 'access.unavailable');
        showNotice(error.key, 'danger', error.params);
    } finally {
        setBusy(false);
    }
}
