import { writable } from 'svelte/store';
import type { auth } from '$lib/api/encore-client';

export type User = auth.User;

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);

// Logto owns authentication. Sign-in happens via the /login page's form
// action; sign-out navigates to /logout which ends the Logto session.
export function logout() {
	currentUser.set(null);
	isAuthenticated.set(false);
	window.location.href = '/logout';
}
