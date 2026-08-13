import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	try {
		const res = await client.social.ListConversations();
		return { conversations: res.conversations || [] };
	} catch {
		return { conversations: [] };
	}
};
