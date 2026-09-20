import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const client = getEncoreClient(await getSessionToken(locals));
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
