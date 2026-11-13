const authKey = 'agentmesh.auth';

export function getAuthHeaders() {
    return { ...JSON.parse(sessionStorage.getItem(authKey)) };
}

export function saveToken(token) {
    sessionStorage.setItem(authKey, JSON.stringify({ Authorization: `Bearer ${token}` }));
}

export function clearToken() {
    sessionStorage.removeItem(authKey);
}

export function hasToken() {
    return Object.hasOwn(getAuthHeaders(), 'Authorization');
}
