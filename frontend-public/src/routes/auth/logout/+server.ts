import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { clearSessionCookie, SESSION_COOKIE } from '$lib/server/session';

export const POST: RequestHandler = async ({ cookies }) => {
	const token = cookies.get(SESSION_COOKIE);
	if (token) {
		// Revoke the session server-side (best-effort).
		try {
			const client = getEncoreClient(token);
			await client.auth.Logout({ token });
		} catch {
			/* ignore */
		}
	}
	clearSessionCookie(cookies);
	return json({ ok: true });
};
