import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, params, url, cookies }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	const prefs = await getUserPreferences(locals);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	const res = await client.comics.ListComics({ Page: page, Limit: limit, Language: '', Search: '', SearchField: '', Tag: params.tag, Sort: '', ExcludeMature: '' });
	return { comics: res.comics || [], total: res.total || 0, page, limit, tag: params.tag };
};
