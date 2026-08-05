import { get } from 'svelte/store';
import { currentUser, isAuthenticated, login, register, fetchMe, logout } from '$lib/stores/auth';
import type { User } from '$lib/stores/auth';

vi.mock('$lib/api/client', () => ({
	api: {
		get: vi.fn(),
		post: vi.fn(),
		put: vi.fn(),
		delete: vi.fn()
	},
	ApiError: class ApiError extends Error {
		status: number;
		code: string;
		constructor(status: number, code: string, message: string) {
			super(message);
			this.status = status;
			this.code = code;
		}
	}
}));

const { api, ApiError } = await import('$lib/api/client');

const mockUser: User = {
	id: 'user-1',
	email: 'test@example.com',
	role: 'member',
	tier: 'free',
	created_at: '2024-01-01T00:00:00Z'
};

const mockToken = 'jwt-token-abc123';

function setupLocalStorage() {
	const store: Record<string, string> = {};
	return {
		getItem: vi.fn((key: string) => store[key] ?? null),
		setItem: vi.fn((key: string, value: string) => {
			store[key] = value;
		}),
		removeItem: vi.fn((key: string) => {
			delete store[key];
		})
	};
}

let localStorageMock: ReturnType<typeof setupLocalStorage>;

beforeEach(() => {
	vi.clearAllMocks();
	localStorageMock = setupLocalStorage();
	vi.stubGlobal('localStorage', localStorageMock);
	currentUser.set(null);
	isAuthenticated.set(false);
});

describe('fetchMe', () => {
	it('sets currentUser to null when no token exists', async () => {
		localStorageMock.getItem.mockReturnValue(null);

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(api.get).not.toHaveBeenCalled();
	});

	it('fetches and sets currentUser when token exists', async () => {
		localStorageMock.getItem.mockReturnValue(mockToken);
		vi.mocked(api.get).mockResolvedValue(mockUser);

		const result = await fetchMe();

		expect(result).toEqual(mockUser);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(api.get).toHaveBeenCalledWith('/auth/me');
	});

	it('removes token and sets currentUser to null on 401', async () => {
		localStorageMock.getItem.mockReturnValue(mockToken);
		vi.mocked(api.get).mockRejectedValue(new ApiError(401, 'unauthorized', 'Invalid token'));

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
	});

	it('sets currentUser to null on non-401 error without removing token', async () => {
		localStorageMock.getItem.mockReturnValue(mockToken);
		vi.mocked(api.get).mockRejectedValue(new ApiError(500, 'internal', 'Server error'));

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
		expect(localStorageMock.removeItem).not.toHaveBeenCalled();
	});

	it('handles non-ApiError exceptions gracefully', async () => {
		localStorageMock.getItem.mockReturnValue(mockToken);
		vi.mocked(api.get).mockRejectedValue(new Error('Network failure'));

		const result = await fetchMe();

		expect(result).toBeNull();
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});

describe('login', () => {
	it('stores token in localStorage and sets currentUser', async () => {
		const res = { token: mockToken, user: mockUser };
		vi.mocked(api.post).mockResolvedValue(res);

		const result = await login('test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockToken);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(api.post).toHaveBeenCalledWith('/auth/login', {
			email: 'test@example.com',
			password: 'password123'
		});
	});

	it('propagates errors from the API', async () => {
		const error = new ApiError(401, 'unauthorized', 'Invalid credentials');
		vi.mocked(api.post).mockRejectedValue(error);

		await expect(login('bad@example.com', 'wrong')).rejects.toThrow('Invalid credentials');
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});

describe('register', () => {
	it('stores token in localStorage and sets currentUser', async () => {
		const res = { token: mockToken, user: mockUser };
		vi.mocked(api.post).mockResolvedValue(res);

		const result = await register('test@example.com', 'password123');

		expect(result).toEqual(mockUser);
		expect(localStorageMock.setItem).toHaveBeenCalledWith('token', mockToken);
		expect(get(currentUser)).toEqual(mockUser);
		expect(get(isAuthenticated)).toBe(true);
		expect(api.post).toHaveBeenCalledWith('/auth/register', {
			email: 'test@example.com',
			password: 'password123'
		});
	});
});

describe('logout', () => {
	it('removes token and clears currentUser', () => {
		// Set up logged-in state
		currentUser.set(mockUser);
		isAuthenticated.set(true);
		localStorageMock.setItem('token', mockToken);

		logout();

		expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});

	it('works even when no user is logged in', () => {
		logout();

		expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
		expect(get(currentUser)).toBeNull();
		expect(get(isAuthenticated)).toBe(false);
	});
});
