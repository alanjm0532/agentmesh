import { authenticate, currentUser, logout } from '../services/user-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';
import { setText } from '../utils/i18n.js';

export async function initAuthPanel() {
    const form = document.getElementById('auth-form');
    const status = document.getElementById('auth-status');
    const password = document.getElementById('password');
    const submit = document.getElementById('auth-submit');
    const switchMode = document.getElementById('auth-switch');
    const logoutButton = document.getElementById('auth-logout');
    let registering = true;

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
            const user = await authenticate({
                username: data.get('username'),
                password: data.get('password'),
            }, registering);
            form.reset();
            renderUser(user);
            showNotice(registering ? 'access.registered' : 'access.loggedIn', 'success');
        } catch (error) {
            showNotice(error.key, 'danger', error.params);
        } finally {
            setBusy(false);
        }
    });

    switchMode.addEventListener('click', () => {
        registering = !registering;
        password.autocomplete = registering ? 'new-password' : 'current-password';
        setText(submit, registering ? 'access.register' : 'access.login');
        setText(switchMode, registering ? 'access.toLogin' : 'access.toRegister');
        clearNotice();
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
