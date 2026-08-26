import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { getEncoreClient } from '$lib/server/encore';
import { setSessionCookie } from '$lib/server/session';

// The Encore backend redirects the browser here after a successful OAuth
// provider login, carrying a one-time exchange code (or an error reason).
// We exchange the code for a session server-side, set the HttpOnly cookie,
// and send the user to the app.
export const GET: RequestHandler = async ({ url, cookies }) => {
	const code = url.searchParams.get('code');
	const error = url.searchParams.get('error');

	if (!code) {
		throw redirect(302, `/login?error=${error ?? 'oauth_failed'}`);
	}

	const client = getEncoreClient(undefined);
	try {
		const res = await client.auth.OAuthExchange({ code });
		setSessionCookie(cookies, res.token);
	} catch {
		throw redirect(302, '/login?error=oauth_failed');
	}

	throw redirect(302, '/');
};
