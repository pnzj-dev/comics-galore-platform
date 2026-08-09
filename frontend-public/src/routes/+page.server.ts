import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get('token');
	const client = getEncoreClient(token);

	const [latest, popular, random] = await Promise.all([
		client.comics.ListComics({ Page: 1, Limit: 4, Sort: 'newest' }),
		client.comics.ListComics({ Page: 1, Limit: 4, Sort: 'popular' }),
		client.comics.ListComics({ Page: 1, Limit: 1, Sort: 'random' }),
	]);

	let continueReading: any[] = [];
	let continueProgress: Record<string, { current_page: number; total_pages: number }> = {};

	if (token) {
		try {
			const cr = await client.reading.ContinueReading();
			const items = cr.items || [];
			for (const item of items) {
				continueProgress[item.comic_id] = { current_page: item.current_page, total_pages: item.total_pages };
			}
			if (items.length > 0) {
				const ids = items.map(i => i.comic_id);
				const batch = await client.comics.BatchGetComics({ ids });
				continueReading = batch.comics || [];
			}
		} catch {}
	}

	return {
		latestComics: latest.comics || [],
		popularComics: popular.comics || [],
		comicOfTheDay: random.comics?.[0] || null,
		continueReading,
		continueProgress,
		authed: !!token,
	};
};
