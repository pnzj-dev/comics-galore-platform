import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie } from '$lib/server/session';

export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json();
	const client = getEncoreClient(undefined);

	try {
		const res = await client.auth.Register({ email: body.email, password: body.password, username: body.username, turnstile_token: body.turnstile_token || '' });
		setSessionCookie(cookies, res.token);
		return json(res.user);
	} catch (e) {
		const status = (e as { status?: number }).status ?? 400;
		return json({ message: (e as Error).message || 'registration failed' }, { status });
	}
};
