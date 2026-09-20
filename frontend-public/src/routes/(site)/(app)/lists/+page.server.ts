import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, cookies }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	try {
		const res = await client.comics.ListReadingLists({ ComicID: '' });
		return { lists: res.lists || [] };
	} catch {
		return { lists: [] };
	}
};
