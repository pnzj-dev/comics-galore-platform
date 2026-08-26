import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const client = getEncoreClient(undefined);
	const res = await client.tiers.PlansReady();
	return { plansReady: res.complete };
};
