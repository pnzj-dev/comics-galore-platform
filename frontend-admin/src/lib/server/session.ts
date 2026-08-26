import { getEncoreClient } from './encore';

// The session cookie name is intentionally kept `token` for compatibility
// with existing server route guards; it now holds an opaque session id
// instead of a JWT. HttpOnly means browser JS can no longer read it.
export const SESSION_COOKIE = 'token';

const COOKIE_MAX_AGE = 60 * 60 * 24 * 30;

function sessionCookieOptions() {
	return {
		path: '/',
		httpOnly: true,
		sameSite: 'lax' as const,
		secure: process.env.NODE_ENV === 'production',
		maxAge: COOKIE_MAX_AGE,
	};
}

export function setSessionCookie(cookies: import('@sveltejs/kit').Cookies, token: string) {
	cookies.set(SESSION_COOKIE, token, sessionCookieOptions());
}

export function clearSessionCookie(cookies: import('@sveltejs/kit').Cookies) {
	cookies.delete(SESSION_COOKIE, { path: '/' });
}

// resolveUser returns the authenticated user for the session cookie, or null.
export async function resolveUser(cookies: import('@sveltejs/kit').Cookies) {
	const token = cookies.get(SESSION_COOKIE);
	if (!token) return null;
	const client = getEncoreClient(token);
	try {
		return await client.auth.Me();
	} catch {
		return null;
	}
}
