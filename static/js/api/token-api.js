import { get, post } from './http.js';

export const fetchTokens = () => get('/tokens');
export const createToken = (data) => post('/tokens', data);
export const renameToken = (id, name) => post(`/tokens/${encodeURIComponent(id)}/rename`, { name });
export const revokeToken = (id) => post(`/tokens/${encodeURIComponent(id)}/revoke`, {});
export const fetchCLIRequest = (code) => get(`/cli/request?user_code=${encodeURIComponent(code)}`);
export const authorizeCLI = (user_code, approve) => post('/cli/authorize', { user_code, approve });
