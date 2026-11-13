import { ApiError, get, post } from './http.js';

export async function loginUser(credentials) {
    try {
        return await post('/login', credentials);
    } catch (error) {
        if (error.status === 400) {
            throw new ApiError('error.loginInput', error.status);
        }
        if (error.status === 401) {
            throw new ApiError('error.credentials', error.status);
        }
        throw error;
    }
}

export function fetchProfile() {
    return get('/profile');
}

export function logoutUser() {
    return post('/logout', {});
}

export function fetchToken() {
    return get('/token');
}

export function resetToken() {
    return post('/token/reset', {});
}
