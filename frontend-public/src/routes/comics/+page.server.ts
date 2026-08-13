import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const client = getEncoreClient(undefined);
	const lang = url.searchParams.get('language');
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = 20;
	const params: any = { Page: page, Limit: limit, Language: '', Search: '', Tag: '', Sort: '', ExcludeMature: '' };
	if (lang) params.Language = lang;

	const [res, facets] = await Promise.all([
		client.comics.ListComics(params),
		client.comics.LanguageFacets().catch(() => ({ facets: [] })),
	]);

	return { comics: res.comics || [], total: res.total || 0, page, limit, facets: facets.facets || [], lang: lang || '' };
};
