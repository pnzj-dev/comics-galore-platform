const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:4000';

export class ApiError extends Error {
	status: number;
	code: string;

	constructor(status: number, code: string, message: string) {
		super(message);
		this.status = status;
		this.code = code;
	}
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(options.headers as Record<string, string>)
	};

	const token = localStorage.getItem('token');
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers
	});

	if (!res.ok) {
		const body = await res.json().catch(() => ({}));
		throw new ApiError(res.status, body.code || 'unknown', body.message || 'Request failed');
	}

	return res.json();
}

export const api = {
	get: <T>(path: string) => request<T>(path, { method: 'GET' }),
	post: <T>(path: string, data?: unknown) =>
		request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
	put: <T>(path: string, data?: unknown) =>
		request<T>(path, { method: 'PUT', body: data ? JSON.stringify(data) : undefined }),
	delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
	patch: <T>(path: string, data?: unknown) =>
		request<T>(path, { method: 'PATCH', body: data ? JSON.stringify(data) : undefined })
};
