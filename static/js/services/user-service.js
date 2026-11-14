import { fetchProfile, loginUser, logoutUser } from '../api/user-api.js';
import { ApiError } from '../api/http.js';

const profileLifetime = 30_000;
let cachedUser;
let checkedAt = 0;
let pendingProfile;
let revision = 0;

export function invalidateCurrentUser() {
    revision++;
    cachedUser = undefined;
    checkedAt = 0;
    pendingProfile = null;
}

export async function authenticate(credentials) {
    await loginUser(credentials);
    invalidateCurrentUser();
    // Confirm that the browser accepted the session cookie before showing success.
    const user = await currentUser({ refresh: true });
    if (user === null) throw new ApiError('error.unauthorized', 401);
    return user;
}

export async function currentUser({ refresh = false } = {}) {
    if (!refresh && cachedUser !== undefined && Date.now() - checkedAt < profileLifetime) {
        return cachedUser;
    }
    if (pendingProfile) return pendingProfile;
    const current = revision;
    const request = fetchProfile().catch((error) => {
        if (error.status === 401) return null;
        throw error;
    }).then((user) => {
        if (current === revision) {
            cachedUser = user;
            checkedAt = Date.now();
        }
        return user;
    });
    pendingProfile = request;
    try {
        return await request;
    } finally {
        if (pendingProfile === request) pendingProfile = null;
    }
}

export async function logout() {
    await logoutUser();
    invalidateCurrentUser();
    cachedUser = null;
    checkedAt = Date.now();
}
