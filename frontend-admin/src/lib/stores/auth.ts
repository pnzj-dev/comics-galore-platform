import { writable } from 'svelte/store';
import { encore } from '$lib/api/encore';
import type { User } from '$lib/api/encore-client';

export type { User };

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

function setToken(value: string) {
	document.cookie = `token=${value}; path=/; SameSite=Lax; max-age=2592000`;
}
function getToken(): string | null {
	const m = document.cookie.match(/(?:^|;\s*)token=([^;]*)/);
	return m ? m[1] : null;
}
function clearToken() {
	document.cookie = 'token=; path=/; max-age=0';
}

export async function login(email: string, password: string): Promise<User> {
	const res = await encore.auth.Login({ email, password });
	setToken(res.token);
	currentUser.set(res.user);
	isAuthenticated.set(true);
	return res.user;
}

export async function register(email: string, password: string): Promise<User> {
	const res = await encore.auth.Register({ email, password });
	setToken(res.token);
	currentUser.set(res.user);
	isAuthenticated.set(true);
	return res.user;
}

export async function fetchMe(): Promise<User | null> {
	const token = getToken();
	if (!token) {
		currentUser.set(null);
		isAuthenticated.set(false);
		return null;
	}

	try {
		const user = await encore.auth.Me();
		currentUser.set(user);
		isAuthenticated.set(true);
		return user;
	} catch (e: any) {
		if (e?.status === 401) {
			clearToken();
		}
		currentUser.set(null);
		isAuthenticated.set(false);
		return null;
	}
}

export function logout() {
	clearToken();
	currentUser.set(null);
	isAuthenticated.set(false);
}
