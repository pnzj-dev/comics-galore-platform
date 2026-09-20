import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	const prefs = await getUserPreferences(locals);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	try {
		const res = await client.comics.ListFavorites({ Page: page, Limit: limit });
		return { comics: res.comics || [], total: res.total || 0, page, limit };
	} catch {
		return { comics: [], total: 0, page: 1, limit };
	}
};
