import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url }) => {
	const client = getEncoreClient(undefined);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = 20;
	try {
		const res = await client.comics.GetReadingList(params.id, { Page: page, Limit: limit });
		return { list: res.list, comics: res.comics || [], total: res.total || 0, page, limit };
	} catch {
		return { list: null, comics: [], total: 0, page: 1, limit };
	}
};
