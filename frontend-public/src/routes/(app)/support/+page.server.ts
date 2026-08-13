import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	try {
		const res = await client.social.ListMyTickets();
		return { tickets: res.tickets || [] };
	} catch {
		return { tickets: [] };
	}
};
