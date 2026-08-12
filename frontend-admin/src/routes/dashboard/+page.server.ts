import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const [users, comics, billing, reading, storage] = await Promise.allSettled([
			client.auth.AdminDashboardStats(),
			client.comics.GetComicsStats(),
			client.billing.GetBillingStats(),
			client.reading.GetReadingStats(),
			client.upload.GetStorageStats(),
		]);

		return {
			users: users.status === 'fulfilled' ? users.value : null,
			comics: comics.status === 'fulfilled' ? comics.value : null,
			billing: billing.status === 'fulfilled' ? billing.value : null,
			reading: reading.status === 'fulfilled' ? reading.value : null,
			storage: storage.status === 'fulfilled' ? storage.value : null,
		};
	} catch {
		return { users: null, comics: null, billing: null, reading: null, storage: null };
	}
};
