import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie, SESSION_COOKIE } from '$lib/server/session';

// Impersonation must run server-side: the resulting session token is stored in
// the HttpOnly cookie. The admin's own session is forwarded to the backend.
export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json();
	const adminToken = cookies.get(SESSION_COOKIE);
	const client = getEncoreClient(adminToken);

	try {
		const res = await client.auth.AdminImpersonateUser(body.user_id);
		setSessionCookie(cookies, res.token);
		return json({ ok: true });
	} catch (e) {
		return json({ message: (e as Error).message || 'impersonation failed' }, { status: 403 });
	}
};
