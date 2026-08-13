import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const [comics, flags] = await Promise.all([
			client.comics.PendingComics({
				Page: parseInt(url.searchParams.get('page') || '1'),
				Limit: 20,
				Search: url.searchParams.get('search') || '',
				Sort: url.searchParams.get('sort') || '',
				SortDir: url.searchParams.get('sort_dir') || '',
			}),
			client.comics.ListFlaggedComments({
				Page: 1,
				Limit: 50,
			}),
		]);
		return {
			results: comics.comics || [],
			total: comics.total || 0,
			page: parseInt(url.searchParams.get('page') || '1'),
			limit: 20,
			search: url.searchParams.get('search') || '',
			sort: url.searchParams.get('sort') || '',
			sortDir: url.searchParams.get('sort_dir') || '',
			flags: flags.flags || [],
		};
	} catch {
		return {
			results: [],
			total: 0,
			page: 1,
			limit: 20,
			search: '',
			sort: '',
			sortDir: '',
			flags: [],
		};
	}
};
