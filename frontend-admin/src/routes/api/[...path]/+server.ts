import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:4000';

export const GET: RequestHandler = (e) => forward(e);
export const POST: RequestHandler = (e) => forward(e);
export const PUT: RequestHandler = (e) => forward(e);
export const PATCH: RequestHandler = (e) => forward(e);
export const DELETE: RequestHandler = (e) => forward(e);

async function forward(event: import('./$types').RequestEvent): Promise<Response> {
	const { request, params, locals, fetch } = event;
	const path = params.path;

	const ctx = await locals.logtoClient.getContext();
	const token = ctx.isAuthenticated ? await locals.logtoClient.getIdToken() : undefined;

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
