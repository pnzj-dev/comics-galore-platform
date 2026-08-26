import { writable } from 'svelte/store';
import type { auth } from '$lib/api/encore-client';
import { isProtectedPath } from '$lib/utils/protected-routes';

export type User = auth.User;

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);
// Flips true once the client has hydrated and the auth store has been seeded
// from server data. Until then, layouts should prefer the server-provided
// user (data.user) to avoid a logged-out flash; afterwards the store is the
// single source of truth so logout/login update the UI reactively.
export const hydrated = writable<boolean>(false);

// The session cookie is HttpOnly and set/cleared server-side. These helpers
// POST to the SvelteKit server endpoints which exchange credentials for a
// session cookie and return the user.
async function post(path: string, body: unknown): Promise<User> {
	const res = await fetch(path, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body),
	});
	if (!res.ok) {
		let message = 'request failed';
		try {
			const data = await res.json();
			message = data.message || message;
		} catch {
			/* ignore */
		}
		throw new Error(message);
	}
	const user = (await res.json()) as User;
	currentUser.set(user);
	isAuthenticated.set(true);
	return user;
}

export async function register(username: string, email: string, password: string, turnstileToken?: string): Promise<User> {
	return post('/auth/register', { username, email, password, turnstile_token: turnstileToken || '' });
}

// login can return a User (password-only, session issued) or a TOTP challenge
// ({ requires_totp, mfa_token }) when 2FA is enabled. Only the former mutates
// the auth store.
export async function login(email: string, password: string, turnstileToken?: string): Promise<User | { requires_totp: true; mfa_token: string }> {
	const res = await fetch('/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password, turnstile_token: turnstileToken || '' }),
	});
	if (!res.ok) {
		let message = 'request failed';
		try {
			const data = await res.json();
			message = data.message || message;
		} catch {
			/* ignore */
		}
		throw new Error(message);
	}
	const data = await res.json();
	if (data.requires_totp) {
		return data as { requires_totp: true; mfa_token: string };
	}
	currentUser.set(data as User);
	isAuthenticated.set(true);
	return data as User;
}

// verifyTotpLogin completes the TOTP login step, exchanging the short-lived
// challenge + authenticator code for a session.
export async function verifyTotpLogin(mfaToken: string, code: string): Promise<User> {
	const res = await fetch('/auth/login/totp', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ mfa_token: mfaToken, code }),
	});
	if (!res.ok) {
		let message = 'request failed';
		try {
			const data = await res.json();
			message = data.message || message;
		} catch {
			/* ignore */
		}
		throw new Error(message);
	}
	const user = (await res.json()) as User;
	currentUser.set(user);
	isAuthenticated.set(true);
	return user;
}

export async function logout(redirectTo?: string) {
	const redirect = redirectTo || (isProtectedPath(window.location.pathname) ? '/' : '');
	try {
		await fetch('/auth/logout', { method: 'POST' });
	} catch {
		/* ignore */
	}
	currentUser.set(null);
	isAuthenticated.set(false);
	if (redirect) {
		window.location.assign(redirect);
	}
}
