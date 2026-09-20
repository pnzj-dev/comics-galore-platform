import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	const prefs = await getUserPreferences(locals);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 24;

	const [res, catsRes] = await Promise.all([
		client.comics.SearchSeries({
			Search: url.searchParams.get('search') || '',
			SearchField: url.searchParams.get('search_field') || '',
			Category: url.searchParams.get('category') || '',
			Page: page,
			Limit: limit,
		}),
		client.comics.ListSeriesCategories(),
	]);

	return {
		series: res.series || [],
		total: res.total || 0,
		categories: catsRes.categories || [],
		page,
		limit,
		search: url.searchParams.get('search') || '',
		searchField: url.searchParams.get('search_field') || '',
		category: url.searchParams.get('category') || '',
	};
};
