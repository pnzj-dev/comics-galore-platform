import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const token = cookies.get('token');
	const client = getEncoreClient(token);
	const [s, c] = await Promise.all([
		client.comics.GetSeries(params.id),
		client.comics.SeriesComics(params.id),
	]);
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

	return { series: s, comics, progress, readCount };
};
