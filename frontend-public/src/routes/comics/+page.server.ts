import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	const client = getEncoreClient(undefined);
	const lang = url.searchParams.get('language');
	const search = url.searchParams.get('search') || '';
	const searchField = url.searchParams.get('search_field') || '';
	const tag = url.searchParams.get('tag') || '';
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = 20;
	const params: any = {
		Page: page,
		Limit: limit,
		Language: '',
		Search: search,
		SearchField: searchField,
		Tag: tag,
		Sort: '',
		ExcludeMature: '',
	};
	if (lang) params.Language = lang;

	const [res, facets, tags] = await Promise.all([
		client.comics.ListComics(params),
		client.comics.LanguageFacets().catch(() => ({ facets: [] })),
		client.comics.PopularTags().catch(() => ({ tags: [] })),
	]);

	return {
		comics: res.comics || [],
		total: res.total || 0,
		page,
		limit,
		facets: facets.facets || [],
		popularTags: tags.tags || [],
		lang: lang || '',
		search,
		searchField,
		tag,
	};
};
