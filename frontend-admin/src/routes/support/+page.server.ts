import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const client = getEncoreClient(cookies.get('token'));
	try {
		const res = await client.social.AdminListTickets({
			Status: url.searchParams.get('status') || '',
		});
		return {
			tickets: res.tickets || [],
			status: url.searchParams.get('status') || '',
		};
	} catch {
		return { tickets: [], status: '' };
	}
};
