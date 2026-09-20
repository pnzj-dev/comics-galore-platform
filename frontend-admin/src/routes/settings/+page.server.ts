import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	try {
		const client = getEncoreClient(await getSessionToken(locals));
		const res = await client.auth.GetAdminSettings();
		return { settings: res };
	} catch {
		return { settings: null };
	}
};
