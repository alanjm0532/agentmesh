import { fetchTokens } from '../api/token-api.js';

export { renameToken, revokeToken, fetchCLIRequest, authorizeCLI } from '../api/token-api.js';

export async function listTokens() {
    const result = await fetchTokens();
    return result.tokens;
}
