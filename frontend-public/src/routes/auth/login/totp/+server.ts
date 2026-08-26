import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie } from '$lib/server/session';

// Completes the TOTP step of login: verifies the authenticator code against
// the short-lived MFA challenge, then sets the session cookie.
export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json();
	const client = getEncoreClient(undefined);

	try {
		const res = await client.auth.VerifyTOTPLogin({ mfa_token: body.mfa_token, code: body.code });
		setSessionCookie(cookies, res.token);
		return json(res.user);
	} catch (e) {
		return json({ message: (e as Error).message || 'invalid code' }, { status: 401 });
	}
};
