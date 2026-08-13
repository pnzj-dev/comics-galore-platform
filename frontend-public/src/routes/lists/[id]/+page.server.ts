import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const client = getEncoreClient(undefined);
	try {
		const res = await client.comics.GetReadingList(params.id);
		return { list: res.list, comics: res.comics || [] };
	} catch {
		return { list: null, comics: [] };
	}
};
