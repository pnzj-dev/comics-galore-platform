import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const client = getEncoreClient();
	const [s, c] = await Promise.all([
		client.comics.GetSeries(params.id),
		client.comics.SeriesComics(params.id),
	]);
	return { series: s, comics: c.comics || [] };
};
