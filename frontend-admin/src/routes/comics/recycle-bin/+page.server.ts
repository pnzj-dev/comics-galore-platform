import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const res = await client.comics.RecycleBin();
		return { comics: res.comics || [] };
	} catch {
		return { comics: [] };
	}
};
