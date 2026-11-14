import { currentUser } from '../services/user-service.js';
import { setText } from '../utils/i18n.js';

export function initAccountLink({ signal }) {
    const label = document.getElementById('account-link-label');
    const link = label.closest('a');

    async function refresh(force = false) {
        try {
            const user = await currentUser({ refresh: force });
            if (signal.aborted) return;
            if (user === null) {
                link.href = '/?page=login';
                setText(label, 'common.apiAccess');
            } else {
                link.href = '/?page=account';
                setText(label, 'home.profile', { username: user.username });
            }
        } catch {
            if (signal.aborted) return;
            link.href = '/?page=account';
            // Keep the account entry available when session verification fails.
            setText(label, 'access.settings');
        }
    }

    refresh();
    window.addEventListener('focus', () => refresh(true), { signal });
    window.addEventListener('pageshow', (event) => {
        // Browser back navigation may restore the page without reloading its scripts.
        if (event.persisted) {
            refresh(true);
        }
    }, { signal });
}
