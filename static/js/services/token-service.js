import { fetchTokens, renameToken, revokeToken, fetchCLIRequest, authorizeCLI } from '../api/token-api.js';

export async function listTokens() {
    const result = await fetchTokens();
    return result.tokens;
}

export const updateTokenName = (id, name) => renameToken(id, name);
export const disableToken = (id) => revokeToken(id);
export const pendingCLIRequest = (code) => fetchCLIRequest(code);
export const decideCLIRequest = (code, approve) => authorizeCLI(code, approve);
