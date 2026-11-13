import { fetchGreeting } from '../api/hello-api.js';

export async function greet(name) {
    const response = await fetchGreeting(name.trim());
    return response.message;
}
