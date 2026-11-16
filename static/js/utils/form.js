export function setControlsBusy(container, busy) {
    container.setAttribute('aria-busy', String(busy));
    container.querySelectorAll('button, input').forEach((control) => {
        control.disabled = busy || control.dataset.unavailable === 'true';
    });
}
