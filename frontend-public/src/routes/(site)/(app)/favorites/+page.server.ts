import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const client = getEncoreClient(cookies.get('token'));
	const prefs = await getUserPreferences(cookies);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	try {
		const res = await client.comics.ListFavorites({ Page: page, Limit: limit });
		return { comics: res.comics || [], total: res.total || 0, page, limit };
	} catch {
		return { comics: [], total: 0, page: 1, limit };
	}
};
