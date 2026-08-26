import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const client = getEncoreClient(cookies.get('token'));
	try {
		const usage = await client.upload.GetStorageUsage();
		return { usage };
	} catch {
		return { usage: null };
	}
};
