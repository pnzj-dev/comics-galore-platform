import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie } from '$lib/server/session';

export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json();
	const client = getEncoreClient(undefined);

	try {
		const res = await client.auth.Login({ email: body.email, password: body.password, turnstile_token: body.turnstile_token || '' });
		setSessionCookie(cookies, res.token);
		return json(res.user);
	} catch (e) {
		return json({ message: (e as Error).message || 'invalid email or password' }, { status: 401 });
	}
};
