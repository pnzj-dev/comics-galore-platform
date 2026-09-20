import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	try {
		const [keysRes, usersRes] = await Promise.all([
			client.mcp.AdminListMcpKeys(),
			client.auth.AdminListUsers({
				Page: 1,
				Limit: 100,
				Search: '',
				Sort: '',
				SortDir: '',
				FilterRole: '',
				FilterTier: '',
				FilterEmail: '',
			}),
		]);
		return {
			keys: keysRes.keys || [],
			users: usersRes.users || [],
		};
	} catch {
		return { keys: [], users: [] };
	}
};
