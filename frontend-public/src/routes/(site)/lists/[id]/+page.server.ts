import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, cookies }) => {
	const token = cookies.get('token');
	const client = getEncoreClient(token);
	const prefs = await getUserPreferences(cookies);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;

	// Owner view first (auth): works for private lists.
	if (token) {
		try {
			const res = await client.comics.GetMyReadingList(params.id, { Page: page, Limit: limit });
			return { list: res.list, comics: res.comics || [], total: res.total || 0, page, limit, isOwner: true };
		} catch {
			// not the owner (or missing) — fall through to the public view
		}
	}

	try {
		const res = await client.comics.GetReadingList(params.id, { Page: page, Limit: limit });
		return { list: res.list, comics: res.comics || [], total: res.total || 0, page, limit, isOwner: false };
	} catch {
		return { list: null, comics: [], total: 0, page: 1, limit, isOwner: false };
	}
};
