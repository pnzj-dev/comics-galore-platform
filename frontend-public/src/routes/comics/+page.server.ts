import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const client = getEncoreClient();
	const lang = url.searchParams.get('language');
	const params: any = { Page: 1, Limit: 20 };
	if (lang) params.Language = lang;
	const res = await client.comics.ListComics(params);
	return { comics: res.comics || [] };
};
