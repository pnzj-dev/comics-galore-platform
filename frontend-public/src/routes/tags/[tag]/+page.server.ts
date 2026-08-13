import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url }) => {
	const client = getEncoreClient(undefined);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = 20;
	const res = await client.comics.ListComics({ Page: page, Limit: limit, Language: '', Search: '', Tag: params.tag, Sort: '', ExcludeMature: '' });
	return { comics: res.comics || [], total: res.total || 0, page, limit, tag: params.tag };
};
