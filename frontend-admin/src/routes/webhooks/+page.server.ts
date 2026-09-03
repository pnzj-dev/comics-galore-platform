import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	try {
		const client = getEncoreClient(cookies.get('token'));
		const [subs, deps] = await Promise.all([
			client.billing.AdminListSubscriptions({
				Page: 1,
				Limit: 50,
				Search: '',
				Sort: 'created_at',
				SortDir: 'desc',
				FilterStatus: '',
				FilterTier: '',
				FilterUserID: '',
			}),
			client.billing.AdminListDeposits({
				Page: 1,
				Limit: 50,
				Search: '',
				Sort: 'created_at',
				SortDir: 'desc',
				FilterStatus: '',
			}),
		]);
		return {
			subscriptions: subs.subscriptions || [],
			deposits: deps.deposits || [],
		};
	} catch {
		return { subscriptions: [], deposits: [] };
	}
};
