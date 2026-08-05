import { api, ApiError } from '$lib/api/client';

function setupLocalStorage(token: string | null = null) {
	const store: Record<string, string> = {};
	if (token) store['token'] = token;
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

function mockFetchResponse(status: number, body: unknown) {
	return vi.fn().mockResolvedValue({
		ok: status >= 200 && status < 300,
		status,
		json: vi.fn().mockResolvedValue(body)
	});
}

beforeEach(() => {
	vi.clearAllMocks();
	vi.stubGlobal('localStorage', setupLocalStorage());
});

describe('api.get', () => {
	it('sends request with Bearer token in Authorization header', async () => {
		const token = 'jwt-token-xyz';
		vi.stubGlobal('localStorage', setupLocalStorage(token));
		vi.stubGlobal('fetch', mockFetchResponse(200, { id: 1, name: 'Test' }));

		await api.get('/comics');

		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/comics'),
			expect.objectContaining({
				method: 'GET',
				headers: expect.objectContaining({
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				})
			})
		);
	});

	it('sends request without Authorization header when no token', async () => {
		vi.stubGlobal('fetch', mockFetchResponse(200, { id: 1 }));

		await api.get('/comics');

		const callArgs = vi.mocked(fetch).mock.calls[0];
		const headers = callArgs[1]?.headers as Record<string, string>;
		expect(headers['Authorization']).toBeUndefined();
	});

	it('returns parsed JSON on success', async () => {
		const data = { id: 1, title: 'My Comic' };
		vi.stubGlobal('fetch', mockFetchResponse(200, data));

		const result = await api.get<{ id: number; title: string }>('/comics/1');

		expect(result).toEqual(data);
	});

	it('uses the correct API base URL', async () => {
		vi.stubGlobal('fetch', mockFetchResponse(200, {}));

		await api.get('/comics');

		expect(fetch).toHaveBeenCalledWith(
			expect.stringMatching(/\/comics$/),
			expect.any(Object)
		);
	});
});

describe('api.post', () => {
	it('sends JSON body with correct Content-Type header', async () => {
		const token = 'jwt-token-abc';
		vi.stubGlobal('localStorage', setupLocalStorage(token));
		vi.stubGlobal('fetch', mockFetchResponse(201, { id: 'new-1' }));

		const payload = { title: 'New Comic', description: 'A new release' };
		await api.post('/comics', payload);

		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/comics'),
			expect.objectContaining({
				method: 'POST',
				body: JSON.stringify(payload),
				headers: expect.objectContaining({
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				})
			})
		);
	});

	it('sends request without body when data is undefined', async () => {
		vi.stubGlobal('fetch', mockFetchResponse(200, {}));

		await api.post('/reset');

		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/reset'),
			expect.objectContaining({
				method: 'POST',
				body: undefined
			})
		);
	});

	it('returns parsed JSON on success', async () => {
		const data = { token: 'tkn', user: { id: 'u1', email: 'a@b.com' } };
		vi.stubGlobal('fetch', mockFetchResponse(200, data));

		const result = await api.post<{ token: string; user: object }>('/auth/login', {
			email: 'a@b.com',
			password: 's3cret'
		});

		expect(result).toEqual(data);
	});
});

describe('api.put', () => {
	it('sends PUT request with JSON body', async () => {
		vi.stubGlobal('fetch', mockFetchResponse(200, { updated: true }));

		const payload = { title: 'Updated Title' };
		await api.put('/comics/1', payload);

		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/comics/1'),
			expect.objectContaining({
				method: 'PUT',
				body: JSON.stringify(payload)
			})
		);
	});
});

describe('api.delete', () => {
	it('sends DELETE request', async () => {
		vi.stubGlobal('fetch', mockFetchResponse(204, null));

		await api.delete('/comics/1');

		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/comics/1'),
			expect.objectContaining({
				method: 'DELETE'
			})
		);
	});
});

describe('ApiError', () => {
	it('is thrown on non-OK response with parsed body', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue({
				ok: false,
				status: 404,
				json: vi.fn().mockResolvedValue({
					code: 'not_found',
					message: 'Comic not found'
				})
			})
		);

		await expect(api.get('/comics/999')).rejects.toThrow('Comic not found');

		try {
			await api.get('/comics/999');
		} catch (e) {
			expect(e).toBeInstanceOf(ApiError);
			expect((e as ApiError).status).toBe(404);
			expect((e as ApiError).code).toBe('not_found');
		}
	});

	it('uses fallback message when response body cannot be parsed', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue({
				ok: false,
				status: 500,
				json: vi.fn().mockRejectedValue(new Error('Invalid JSON'))
			})
		);

		try {
			await api.get('/broken');
		} catch (e) {
			expect(e).toBeInstanceOf(ApiError);
			expect((e as ApiError).status).toBe(500);
			expect((e as ApiError).code).toBe('unknown');
			expect((e as ApiError).message).toBe('Request failed');
		}
	});

	it('is thrown with 401 status', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue({
				ok: false,
				status: 401,
				json: vi.fn().mockResolvedValue({
					code: 'unauthorized',
					message: 'Invalid token'
				})
			})
		);

		try {
			await api.get('/auth/me');
		} catch (e) {
			expect(e).toBeInstanceOf(ApiError);
			expect((e as ApiError).status).toBe(401);
			expect((e as ApiError).code).toBe('unauthorized');
		}
	});
});
