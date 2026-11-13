import { fetchProfile, loginUser, logoutUser, registerUser } from '../api/user-api.js';

export async function authenticate(credentials, registering) {
    if (registering) {
        await registerUser(credentials);
    } else {
        await loginUser(credentials);
    }
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
