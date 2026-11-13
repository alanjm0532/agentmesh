import { fetchProfile, fetchToken, loginUser, logoutUser, resetToken } from '../api/user-api.js';

export async function authenticate(credentials) {
    await loginUser(credentials);
    // Confirm that the browser accepted the session cookie before showing success.
    return fetchProfile();
}

export async function currentUser() {
    try {
        return await fetchProfile();
    } catch (error) {
        if (error.status === 401) {
            return null;
        }
        throw error;
    }
}

export async function logout() {
    await logoutUser();
}

export async function currentToken() {
    const result = await fetchToken();
    return result.access_token;
}

export async function regenerateToken() {
    const result = await resetToken();
    return result.access_token;
}
