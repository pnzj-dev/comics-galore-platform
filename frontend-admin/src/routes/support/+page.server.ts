import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const client = getEncoreClient(await getSessionToken(locals));
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
