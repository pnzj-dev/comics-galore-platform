import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const dashboard = await client.dashboard.GetDashboard();
		return { dashboard };
	} catch {
		return { dashboard: null };
	}
};
