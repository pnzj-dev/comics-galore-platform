import { getEncoreClient } from '$lib/server/encore';
import { getUserPreferences } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies, url }) => {
	const token = cookies.get('token');
	const client = getEncoreClient(token);
	const prefs = await getUserPreferences(cookies);
	const page = parseInt(url.searchParams.get('page') || '1');
	const limit = prefs?.items_per_page || 20;
	const s = await client.comics.GetSeries(params.slug);
	const c = await client.comics.SeriesComics(s.id, { Page: page, Limit: limit });
	const comics = c.comics || [];

	// Reading progress (only when authenticated).
	let progress: Record<string, { completed: boolean; current_page: number }> = {};
	if (token) {
		try {
			const res = await client.reading.SeriesProgress({ comic_ids: comics.map((x) => x.id) });
			for (const item of res.items || []) {
				progress[item.comic_id] = { completed: item.completed, current_page: item.current_page };
			}
		} catch {
			progress = {};
		}
	}

	const readCount = comics.filter((x) => progress[x.id]?.completed).length;

	return { series: s, comics, total: c.total || 0, page, limit, progress, readCount };
};
