import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	try {
		const [queue, decisions] = await Promise.all([
			client.comics.AIReviewQueue(),
			client.comics.AIDecisions(),
		]);
		return { queue: queue.items || [], decisions: decisions.decisions || [] };
	} catch {
		return { queue: [], decisions: [] };
	}
};
