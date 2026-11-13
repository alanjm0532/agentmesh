import { currentUser } from '../services/user-service.js';
import { setText } from '../utils/i18n.js';

export function initAccountLink() {
    const label = document.getElementById('account-link-label');

    async function refresh() {
        try {
            const user = await currentUser();
            if (user === null) {
                setText(label, 'common.apiAccess');
            } else {
                setText(label, 'home.account', { username: user.username });
            }
        } catch {
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
