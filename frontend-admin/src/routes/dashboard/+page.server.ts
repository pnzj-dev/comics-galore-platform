import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	try {
		const client = getEncoreClient(await getSessionToken(locals));
		const dashboard = await client.dashboard.GetDashboard();
		return { dashboard };
	} catch {
		return { dashboard: null };
	}
};
