import { get } from 'svelte/store';
import { currentUser, isAuthenticated, login, register, fetchMe, logout } from '$lib/stores/auth';

vi.mock('$lib/api/encore', () => ({
	encore: {
		auth: {
			Login: vi.fn(),
			Register: vi.fn(),
			Me: vi.fn(),
		},
	},
}));

const { encore } = await import('$lib/api/encore');

const mockUser = {
	id: 'user-1',
	email: 'test@example.com',
	role: 'member',
	tier: 'free',
	created_at: '',
};

const mockToken = 'jwt-token-abc123';

const mockAuthResponse = { token: mockToken, user: mockUser };

function setupCookie(token: string | null = null) {
	let value = token ? `token=${token}` : '';
	return {
		get value() { return value; },
		set value(v: string) { value = v; }
	};
}

let cookieStore: ReturnType<typeof setupCookie>;

beforeEach(() => {
	vi.clearAllMocks();
	cookieStore = setupCookie();
	Object.defineProperty(document, 'cookie', {
		get: vi.fn(() => cookieStore.value),
		set: vi.fn((v: string) => { cookieStore.value = v; }),
		configurable: true
	});
	currentUser.set(null);
	isAuthenticated.set(false);
});

describe('fetchMe', () => {
	it('sets currentUser to null when no token exists', async () => {
		cookieStore.value = '';
		const result = await fetchMe();
		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(encore.auth.Me).not.toHaveBeenCalled();
	});

	it('fetches and sets currentUser when token exists', async () => {
		cookieStore.value = `token=${mockToken}`;
		vi.mocked(encore.auth.Me).mockResolvedValue(mockUser);

		const result = await fetchMe();

		expect(result).toEqual(mockUser);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(encore.auth.Me).toHaveBeenCalled();
	});

	it('removes token and sets currentUser to null on 401', async () => {
		cookieStore.value = `token=${mockToken}`;
		const err = { status: 401 };
		vi.mocked(encore.auth.Me).mockRejectedValue(err);

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(cookieStore.value).not.toContain(mockToken);
	});

	it('sets currentUser to null on non-401 error without removing token', async () => {
		cookieStore.value = `token=${mockToken}`;
		vi.mocked(encore.auth.Me).mockRejectedValue(new Error('Server error'));

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(cookieStore.value).toContain(mockToken);
	});

	it('handles non-ApiError exceptions gracefully', async () => {
		cookieStore.value = `token=${mockToken}`;
		vi.mocked(encore.auth.Me).mockRejectedValue(new Error('Network failure'));

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});

describe('login', () => {
	it('stores token in cookie and sets currentUser', async () => {
		vi.mocked(encore.auth.Login).mockResolvedValue(mockAuthResponse);

		const result = await login('test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(cookieStore.value).toContain(mockToken);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(encore.auth.Login).toHaveBeenCalledWith({
			email: 'test@example.com',
			password: 'password123'
		});
	});

	it('propagates errors from the API', async () => {
		const error = new Error('Invalid credentials');
		vi.mocked(encore.auth.Login).mockRejectedValue(error);

		await expect(login('bad@example.com', 'wrong')).rejects.toThrow('Invalid credentials');
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});

describe('register', () => {
	it('stores token in cookie and sets currentUser', async () => {
		vi.mocked(encore.auth.Register).mockResolvedValue(mockAuthResponse);

		const result = await register('test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(cookieStore.value).toContain(mockToken);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(encore.auth.Register).toHaveBeenCalledWith({
			email: 'test@example.com',
			password: 'password123'
		});
	});
});

describe('logout', () => {
	it('removes token and clears currentUser', () => {
		currentUser.set(mockUser);
		isAuthenticated.set(true);
		cookieStore.value = `token=${mockToken}`;

		logout();

		expect(cookieStore.value).not.toContain(mockToken);
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});

	it('works even when no user is logged in', () => {
		logout();

		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
