import { get } from './http.js';

export function fetchGreeting(name) {
    const query = new URLSearchParams({ name });
    return get(`/hello?${query}`);
}
