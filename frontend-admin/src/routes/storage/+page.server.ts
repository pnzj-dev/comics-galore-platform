import { getEncoreClient } from '$lib/server/encore';
import { getSessionToken } from '$lib/server/session';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const client = getEncoreClient(await getSessionToken(locals));
	try {
		const usage = await client.upload.GetStorageUsage();
		return { usage };
	} catch {
		return { usage: null };
	}
};
