import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const client = getEncoreClient(cookies.get('token'));
	const prefs = await getUserPreferences(cookies);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 24;

	const [res, catsRes] = await Promise.all([
		client.comics.SearchSeries({
			Search: url.searchParams.get('search') || '',
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
		category: url.searchParams.get('category') || '',
	};
};
