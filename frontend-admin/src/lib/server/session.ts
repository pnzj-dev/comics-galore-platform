import { getEncoreClient } from './encore';

// Logto owns authentication and manages its own encrypted HttpOnly session
// cookie. The SvelteKit server resolves the user by asking Encore to validate
// the Logto ID token and map it to the internal user.

// getSessionToken returns the Logto ID token for the current session, or
// undefined when signed out. Server load functions forward it as a Bearer
// token to the Encore backend.
export async function getSessionToken(locals: App.Locals): Promise<string | undefined> {
	const ctx = await locals.logtoClient.getContext();
	if (!ctx.isAuthenticated) return undefined;
	const token = await locals.logtoClient.getIdToken();
	return token ?? undefined;
}

export async function resolveUser(locals: App.Locals) {
	const token = await getSessionToken(locals);
	if (!token) return null;
	const client = getEncoreClient(token);
	try {
		return await client.auth.Me();
	} catch {
		return null;
	}
}
