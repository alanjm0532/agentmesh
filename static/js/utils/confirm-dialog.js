import { setText } from './i18n.js';

let dialog;

export function confirmAction({ title, message, confirm = 'common.confirm', signal }) {
    if (signal?.aborted) return Promise.resolve(false);
    if (!dialog) {
        dialog = document.createElement('dialog');
        dialog.className = 'confirm-dialog';
        dialog.setAttribute('aria-labelledby', 'confirm-dialog-title');
        dialog.setAttribute('aria-describedby', 'confirm-dialog-message');
        dialog.innerHTML = `
            <form method="dialog">
                <h2 id="confirm-dialog-title"></h2>
                <p id="confirm-dialog-message"></p>
                <div class="confirm-dialog-actions">
                    <button class="btn action-secondary" value="cancel" data-i18n="common.cancel" autofocus></button>
                    <button class="btn action-primary" value="confirm"></button>
                </div>
            </form>
        `;
        document.body.append(dialog);
        dialog.addEventListener('click', (event) => {
            if (event.target !== dialog) return;
            const bounds = dialog.getBoundingClientRect();
            if (event.clientX < bounds.left || event.clientX > bounds.right ||
                event.clientY < bounds.top || event.clientY > bounds.bottom) {
                dialog.close('cancel');
            }
        });
    }
    if (dialog.open) return Promise.resolve(false);

    setText(dialog.querySelector('h2'), title);
    setText(dialog.querySelector('p'), message);
    setText(dialog.querySelector('[value="cancel"]'), 'common.cancel');
    setText(dialog.querySelector('[value="confirm"]'), confirm);
    dialog.returnValue = 'cancel';

    return new Promise((resolve) => {
        const dismiss = () => dialog.close('cancel');
        dialog.addEventListener('close', () => {
            window.removeEventListener('pagehide', dismiss);
            signal?.removeEventListener('abort', dismiss);
            resolve(dialog.returnValue === 'confirm');
        }, { once: true });
        window.addEventListener('pagehide', dismiss, { once: true });
        signal?.addEventListener('abort', dismiss, { once: true });
        dialog.showModal();
    });
}
