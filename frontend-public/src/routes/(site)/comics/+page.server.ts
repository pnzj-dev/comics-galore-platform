import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { comics } from '$lib/server/client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url, cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	const prefs = await getUserPreferences(cookies);

	// Default the language filter to the user's preferred content language.
	const lang = url.searchParams.get('language') || prefs?.content_language || '';
	const search = url.searchParams.get('search') || '';
	const searchField = url.searchParams.get('search_field') || '';
	const tag = url.searchParams.get('tag') || '';
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	const params: comics.ListComicsParams = {
		Page: page,
		Limit: limit,
		Language: lang || '',
		Search: search,
		SearchField: searchField,
		Tag: tag,
		Sort: '',
		ExcludeMature: '',
	};

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
