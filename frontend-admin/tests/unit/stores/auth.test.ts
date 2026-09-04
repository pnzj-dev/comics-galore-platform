import { get } from 'svelte/store';
import { currentUser, isAuthenticated, login, logout } from '$lib/stores/auth';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const mockUser = {
	id: 'user-1',
	email: 'admin@example.com',
	role: 'admin',
	tier: 'platinum',
	created_at: '',
};

function mockFetchResponse(body: unknown, ok = true, status = 200) {
	return {
		ok,
		status,
		json: vi.fn().mockResolvedValue(body),
	};
}

beforeEach(() => {
	vi.clearAllMocks();
	vi.stubGlobal('fetch', vi.fn());
	currentUser.set(null);
	isAuthenticated.set(false);
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('login', () => {
	it('posts to /auth/login and sets currentUser', async () => {
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse(mockUser) as Response);

		const result = await login('admin@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(fetch).toHaveBeenCalledWith('/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: 'admin@example.com', password: 'password123', turnstile_token: '' }),
		});
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
	});

	it('passes the turnstile token through', async () => {
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse(mockUser) as Response);

		await login('admin@example.com', 'password123', 'test-token');

		expect(fetch).toHaveBeenCalledWith(
			'/auth/login',
			expect.objectContaining({
				body: JSON.stringify({ email: 'admin@example.com', password: 'password123', turnstile_token: 'test-token' }),
			}),
		);
	});

	it('throws a user-facing message on error', async () => {
		vi.mocked(fetch).mockResolvedValue(
			mockFetchResponse({ message: 'invalid email or password' }, false, 401) as Response,
		);

		await expect(login('bad@example.com', 'wrong')).rejects.toThrow('invalid email or password');
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});

describe('logout', () => {
	it('posts to /auth/logout and clears currentUser', () => {
		currentUser.set(mockUser);
		isAuthenticated.set(true);
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse({ ok: true }) as Response);

		logout();

		expect(fetch).toHaveBeenCalledWith('/auth/logout', { method: 'POST' });
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});

	it('works even when no user is logged in', () => {
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse({ ok: true }) as Response);

		logout();

		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
