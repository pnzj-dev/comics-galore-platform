import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	try {
		let search = url.searchParams.get('search') || '';
		const fallback: string[] = [];
		url.searchParams.forEach((v, k) => {
			if (k.startsWith('filter_') && v && !['filter_status', 'filter_tier'].includes(k)) {
				fallback.push(v);
			}
		});
		if (fallback.length > 0) search = [search, ...fallback].filter(Boolean).join(' ');

		const client = getEncoreClient(cookies.get('token'));
		const res = await client.billing.AdminListSubscriptions({
			Page: parseInt(url.searchParams.get('page') || '1'),
			Limit: 20,
			Search: search,
			Sort: url.searchParams.get('sort') || '',
			SortDir: url.searchParams.get('sort_dir') || '',
			FilterStatus: url.searchParams.get('filter_status') || '',
			FilterTier: url.searchParams.get('filter_tier') || '',
		});
		return {
			results: res.subscriptions || [],
			total: res.total || 0,
			page: parseInt(url.searchParams.get('page') || '1'),
			limit: 20,
			search,
			sort: url.searchParams.get('sort') || '',
			sortDir: url.searchParams.get('sort_dir') || '',
			filters: {
				status: url.searchParams.get('filter_status') || '',
				tier: url.searchParams.get('filter_tier') || '',
				user_id: url.searchParams.get('filter_user_id') || '',
				plan_id: url.searchParams.get('filter_plan_id') || '',
			},
		};
	} catch {
		return { results: [], total: 0, page: 1, limit: 20, search: '', sort: '', sortDir: '', filters: {} };
	}
};
