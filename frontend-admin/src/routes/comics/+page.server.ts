import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	try {
		let search = url.searchParams.get('search') || '';
		const fallback: string[] = [];
		url.searchParams.forEach((v, k) => {
			if (k.startsWith('filter_') && v && !['filter_status', 'filter_author'].includes(k)) {
				fallback.push(v);
			}
		});
		if (fallback.length > 0) search = [search, ...fallback].filter(Boolean).join(' ');

		const client = getEncoreClient(cookies.get('token'));
		const res = await client.comics.AdminListComics({
			Page: parseInt(url.searchParams.get('page') || '1'),
			Limit: 20,
			Search: search,
			Sort: url.searchParams.get('sort') || '',
			SortDir: url.searchParams.get('sort_dir') || '',
			FilterStatus: url.searchParams.get('filter_status') || '',
			FilterAuthor: url.searchParams.get('filter_author') || '',
		});
		return {
			results: res.comics || [],
			total: res.total || 0,
			page: parseInt(url.searchParams.get('page') || '1'),
			limit: 20,
			search,
			sort: url.searchParams.get('sort') || '',
			sortDir: url.searchParams.get('sort_dir') || '',
			filters: {
				status: url.searchParams.get('filter_status') || '',
				author: url.searchParams.get('filter_author') || '',
				title: url.searchParams.get('filter_title') || '',
			},
		};
	} catch {
		return { results: [], total: 0, page: 1, limit: 20, search: '', sort: '', sortDir: '', filters: {} };
	}
};
