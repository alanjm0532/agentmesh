import { ApiError, get, post } from './http.js';

export async function registerUser(credentials) {
    try {
        return await post('/register', credentials);
    } catch (error) {
        if (error.status === 409) {
            throw new ApiError('error.userExists', error.status);
        }
        if (error.status === 400) {
            throw new ApiError('error.passwordLength', error.status);
        }
        throw error;
    }
}

export async function loginUser(credentials) {
    try {
        return await post('/login', credentials);
    } catch (error) {
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
