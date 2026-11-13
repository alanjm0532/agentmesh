import { t } from '../utils/i18n.js';

const webPrefix = '/web';

export class ApiError extends Error {
    constructor(key, status = 0) {
        super(t(key, { status }));
        this.name = 'ApiError';
        this.key = key;
        this.params = { status };
        this.status = status;
    }
}

async function request(path, options) {
    let response;
    try {
        response = await fetch(`${webPrefix}${path}`, {
            ...options,
            credentials: 'same-origin',
            headers: {
                Accept: 'application/json',
                ...options.headers,
            },
        });
    } catch {
        throw new ApiError('error.network');
    }

    if (response.status === 401) {
        throw new ApiError('error.unauthorized', 401);
    }

    if (!response.ok) {
        throw new ApiError('error.request', response.status);
    }

    if (response.status === 204) {
        return;
    }

    try {
        return await response.json();
    } catch {
        throw new ApiError('error.json', response.status);
    }
}

export function get(path) {
    return request(path, { method: 'GET' });
}

export function post(path, data) {
    return request(path, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
}
