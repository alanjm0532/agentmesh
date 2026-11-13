import { greet } from '../services/hello-service.js';
import { clearNotice, showNotice } from '../utils/notice.js';

export function initHelloPanel() {
    const form = document.getElementById('hello-form');
    const submit = document.getElementById('hello-submit');
    const greeting = document.getElementById('greeting');

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        clearNotice();
        submit.disabled = true;
        greeting.setAttribute('aria-busy', 'true');
        try {
            greeting.textContent = await greet(new FormData(form).get('name'));
            // API response content is displayed as returned by the server.
            greeting.removeAttribute('data-i18n');
        } catch (error) {
            showNotice(error.key, 'danger', error.params);
        } finally {
            submit.disabled = false;
            greeting.setAttribute('aria-busy', 'false');
        }
    });
}
