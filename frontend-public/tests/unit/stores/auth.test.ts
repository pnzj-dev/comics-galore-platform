import { get } from 'svelte/store';
import { currentUser, isAuthenticated, login, register, logout } from '$lib/stores/auth';

const mockUser = {
	id: 'user-1',
	email: 'test@example.com',
	role: 'member',
	tier: 'free',
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

		const result = await login('test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(fetch).toHaveBeenCalledWith('/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email: 'test@example.com', password: 'password123', turnstile_token: '' }),
		});
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
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

describe('register', () => {
	it('posts to /auth/register and sets currentUser', async () => {
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse(mockUser) as Response);

		const result = await register('testuser', 'test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(fetch).toHaveBeenCalledWith('/auth/register', expect.objectContaining({ method: 'POST' }));
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
	});
});

describe('logout', () => {
	it('posts to /auth/logout and clears currentUser', async () => {
		currentUser.set(mockUser);
		isAuthenticated.set(true);
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse({ ok: true }) as Response);

		await logout();

		expect(fetch).toHaveBeenCalledWith('/auth/logout', { method: 'POST' });
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});

	it('works even when no user is logged in', async () => {
		vi.mocked(fetch).mockResolvedValue(mockFetchResponse({ ok: true }) as Response);
		await logout();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
