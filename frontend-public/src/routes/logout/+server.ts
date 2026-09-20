import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// Ends the Logto session. Redirects to Logto, then back to the app origin.
export const GET: RequestHandler = async ({ locals, url }) => {
	await locals.logtoClient.signOut(url.origin);
	throw redirect(302, url.origin);
};
