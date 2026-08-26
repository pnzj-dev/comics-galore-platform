import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { SESSION_COOKIE } from '$lib/server/session';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:4000';

export const GET: RequestHandler = ({ request, params, cookies, fetch }) =>
	forward(request, params.path, cookies, fetch);

export const POST: RequestHandler = ({ request, params, cookies, fetch }) =>
	forward(request, params.path, cookies, fetch);

export const PUT: RequestHandler = ({ request, params, cookies, fetch }) =>
	forward(request, params.path, cookies, fetch);

export const PATCH: RequestHandler = ({ request, params, cookies, fetch }) =>
	forward(request, params.path, cookies, fetch);

export const DELETE: RequestHandler = ({ request, params, cookies, fetch }) =>
	forward(request, params.path, cookies, fetch);

async function forward(
	request: Request,
	path: string,
	cookies: import('@sveltejs/kit').Cookies,
	fetch: typeof globalThis.fetch,
): Promise<Response> {
	const token = cookies.get(SESSION_COOKIE);
	const url = `${BACKEND_URL}/${path}${new URL(request.url).search}`;

	const headers = new Headers(request.headers);
	headers.delete('host');
	headers.delete('origin');
	headers.delete('content-length');
	if (token) headers.set('Authorization', `Bearer ${token}`);

	const upstream = await fetch(url, {
		method: request.method,
		headers,
		body: ['GET', 'HEAD'].includes(request.method) ? undefined : await request.arrayBuffer(),
	});

	const contentType = upstream.headers.get('content-type') || '';
	if (contentType.includes('application/json')) {
		const text = await upstream.text();
		let body: unknown = {};
		if (text) {
			try {
				body = JSON.parse(text);
			} catch {
				// empty/invalid JSON body (e.g. void endpoints) — treat as empty
			}
		}
		if (!upstream.ok) {
			error(upstream.status, (body as { message?: string })?.message ?? 'request failed');
		}
		return json(body);
	}

	return new Response(upstream.body, {
		status: upstream.status,
		headers: upstream.headers,
	});
}
