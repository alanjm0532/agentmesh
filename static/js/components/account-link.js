import { currentUser } from '../services/user-service.js';
import { setText } from '../utils/i18n.js';

export function initAccountLink() {
    const label = document.getElementById('account-link-label');
    const link = label.closest('a');

    async function refresh() {
        try {
            const user = await currentUser();
            if (user === null) {
                link.href = '/login.html';
                setText(label, 'common.apiAccess');
            } else {
                link.href = '/account.html';
                setText(label, 'home.profile', { username: user.username });
            }
        } catch {
            link.href = '/account.html';
            // Keep the account entry available when session verification fails.
            setText(label, 'access.settings');
        }
    }

    refresh();
    window.addEventListener('focus', refresh);
    window.addEventListener('pageshow', (event) => {
        // Browser back navigation may restore the page without reloading its scripts.
        if (event.persisted) {
            refresh();
        }
    });
}
