import { writable } from 'svelte/store';
import { api, ApiError } from '$lib/api/client';

export interface User {
	id: string;
	email: string;
	role: string;
	tier: string;
	created_at: string;
}

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

export async function login(email: string, password: string): Promise<User> {
	const res = await api.post<{ token: string; user: User }>('/auth/login', { email, password });
	localStorage.setItem('token', res.token);
	currentUser.set(res.user);
	isAuthenticated.set(true);
	return res.user;
}

export async function register(email: string, password: string): Promise<User> {
	const res = await api.post<{ token: string; user: User }>('/auth/register', { email, password });
	localStorage.setItem('token', res.token);
	currentUser.set(res.user);
	isAuthenticated.set(true);
	return res.user;
}

export async function fetchMe(): Promise<User | null> {
	const token = localStorage.getItem('token');
	if (!token) {
		currentUser.set(null);
		isAuthenticated.set(false);
		return null;
	}

	try {
		const user = await api.get<User>('/auth/me');
		currentUser.set(user);
		isAuthenticated.set(true);
		return user;
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			localStorage.removeItem('token');
		}
		currentUser.set(null);
		isAuthenticated.set(false);
		return null;
	}
}

export function logout() {
	localStorage.removeItem('token');
	currentUser.set(null);
	isAuthenticated.set(false);
}
