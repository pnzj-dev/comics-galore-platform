import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// Starts a Logto authentication flow. The `?mode` query param deep-links to a
// specific screen of Logto's hosted sign-in experience:
//   (default) sign-in (which itself offers sign-up and forgot-password)
//   ?mode=signup → register
//   ?mode=forgot → reset_password
export const GET: RequestHandler = async ({ locals, url }) => {
	const mode = url.searchParams.get('mode');
	const firstScreen = mode === 'signup' ? 'register' : mode === 'forgot' ? 'reset_password' : 'signIn';
	await locals.logtoClient.signIn({ redirectUri: `${url.origin}/callback`, firstScreen });
	throw redirect(302, url.origin);
};
