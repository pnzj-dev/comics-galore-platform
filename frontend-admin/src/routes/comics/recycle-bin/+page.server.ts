import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const res = await client.comics.RecycleBin({
			Page: parseInt(url.searchParams.get('page') || '1'),
			Limit: 20,
			Search: url.searchParams.get('search') || '',
			Sort: url.searchParams.get('sort') || '',
			SortDir: url.searchParams.get('sort_dir') || '',
			FilterStatus: url.searchParams.get('filter_status') || '',
			FilterAuthor: url.searchParams.get('filter_author') || '',
			FilterTitle: url.searchParams.get('filter_title') || '',
		});
		return {
			results: res.comics || [],
			total: res.total || 0,
			page: parseInt(url.searchParams.get('page') || '1'),
			limit: 20,
			search: url.searchParams.get('search') || '',
			sort: url.searchParams.get('sort') || '',
			sortDir: url.searchParams.get('sort_dir') || '',
			showFilters: url.searchParams.get('show_filters') === '1',
			filters: {
				status: url.searchParams.get('filter_status') || '',
				author: url.searchParams.get('filter_author') || '',
				title: url.searchParams.get('filter_title') || '',
			},
		};
	} catch {
		return { results: [], total: 0, page: 1, limit: 20, search: '', sort: '', sortDir: '', showFilters: false, filters: {} };
	}
};
