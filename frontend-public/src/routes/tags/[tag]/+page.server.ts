import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const client = getEncoreClient();
	const res = await client.comics.ListComics({ Page: 1, Limit: 20, Tag: params.tag });
	return { comics: res.comics || [], tag: params.tag };
};
