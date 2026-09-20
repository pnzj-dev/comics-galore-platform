import { writable } from 'svelte/store';
import type { auth } from '$lib/api/encore-client';

export type User = auth.User;

export const currentUser = writable<User | null>(null);
export const isAuthenticated = writable<boolean>(false);
// Flips true once the client has hydrated and the auth store has been seeded
// from server data. Until then, layouts should prefer the server-provided
// user (data.user) to avoid a logged-out flash; afterwards the store is the
// single source of truth so logout/login update the UI reactively.
export const hydrated = writable<boolean>(false);

// Logto owns authentication. Sign-in/sign-out are GET endpoints that redirect
// to the Logto hosted UI (/login and /logout server routes). `?mode` deep-links
// to a specific screen.
export function login() {
	window.location.href = '/login';
}

export function register() {
	window.location.href = '/login?mode=signup';
}

export function forgotPassword() {
	window.location.href = '/login?mode=forgot';
}

export function logout() {
	currentUser.set(null);
	isAuthenticated.set(false);
	window.location.href = '/logout';
}
