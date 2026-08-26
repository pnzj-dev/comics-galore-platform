import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie } from '$lib/server/session';

// Passkey login verify must run server-side so the resulting session token can
// be stored in the HttpOnly cookie. The browser builds the credential via
// navigator.credentials.get() and POSTs the response here.
export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json();
	const client = getEncoreClient(undefined);

	try {
		const res = await client.auth.PasskeyLoginVerify({ response: body });
		setSessionCookie(cookies, res.token);
		return json(res.user);
	} catch (e) {
		const status = (e as { status?: number }).status ?? 401;
		return json({ message: (e as Error).message || 'passkey login failed' }, { status });
	}
};
