import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const res = await client.auth.AdminListUsers({
			Page: parseInt(url.searchParams.get('page') || '1'),
			Limit: 20,
			Search: url.searchParams.get('search') || '',
			Sort: url.searchParams.get('sort') || '',
			SortDir: url.searchParams.get('sort_dir') || '',
			FilterRole: url.searchParams.get('filter_role') || '',
			FilterTier: url.searchParams.get('filter_tier') || '',
		});
		return {
			results: res.users || [],
			total: res.total || 0,
			page: parseInt(url.searchParams.get('page') || '1'),
			limit: 20,
			search: url.searchParams.get('search') || '',
			sort: url.searchParams.get('sort') || '',
			sortDir: url.searchParams.get('sort_dir') || '',
			filters: {
				role: url.searchParams.get('filter_role') || '',
				tier: url.searchParams.get('filter_tier') || '',
			},
		};
	} catch {
		return { results: [], total: 0, page: 1, limit: 20, search: '', sort: '', sortDir: '', filters: {} };
	}
};
