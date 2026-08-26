import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url, cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	const prefs = await getUserPreferences(cookies);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	const res = await client.comics.ListComics({ Page: page, Limit: limit, Language: '', Search: '', SearchField: '', Tag: params.tag, Sort: '', ExcludeMature: '' });
	return { comics: res.comics || [], total: res.total || 0, page, limit, tag: params.tag };
};
