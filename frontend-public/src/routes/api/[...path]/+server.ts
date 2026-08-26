import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { SESSION_COOKIE } from '$lib/server/session';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:4000';

// Same-origin proxy for authenticated browser calls. The browser cannot read
// the HttpOnly session cookie, so authenticated mutations are routed through
// this endpoint; the server reads the cookie and forwards the request to the
// Encore backend with the session as a Bearer token.
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
	// The browser may echo content-type etc.; strip host/origin so the backend
	// sees a clean server-to-server request.
	headers.delete('host');
	headers.delete('origin');
	headers.delete('content-length');
	if (token) headers.set('Authorization', `Bearer ${token}`);

	const body = ['GET', 'HEAD'].includes(request.method) ? undefined : request.body;
	const init: RequestInit & { duplex?: 'half' } = {
		method: request.method,
		headers,
		body,
	};
	// Node's fetch (undici) requires duplex to stream a web ReadableStream body
	// instead of buffering it — important for large archive uploads.
	if (body) init.duplex = 'half';

	const upstream = await fetch(url, init);

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

	// Non-JSON responses (raw endpoints, streams) pass through.
	return new Response(upstream.body, {
		status: upstream.status,
		headers: upstream.headers,
	});
}
