import { writable } from 'svelte/store';
import type { auth } from '$lib/api/encore-client';

export type User = auth.User;

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

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

export async function login(email: string, password: string): Promise<User> {
	return post('/auth/login', { email, password });
}

export function logout(redirectTo?: string) {
	fetch('/auth/logout', { method: 'POST' }).catch(() => {});
	currentUser.set(null);
	isAuthenticated.set(false);
	if (redirectTo) {
		window.location.assign(redirectTo);
	}
}
