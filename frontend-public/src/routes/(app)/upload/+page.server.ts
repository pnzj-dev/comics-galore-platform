import { redirect } from '@sveltejs/kit';
import { decodeJWT } from '$lib/server/jwt';
import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const token = cookies.get('token');
	if (!token) throw redirect(302, '/login');

	const user = decodeJWT(token);
	if (!user || (user.role !== 'uploader' && user.role !== 'admin')) throw redirect(302, '/');

	const rawTab = url.searchParams.get('tab') ?? 'list';
	const validTabs = ['list', 'manual', 'archive'] as const;
	const tab = validTabs.includes(rawTab as typeof validTabs[number]) ? rawTab : 'list';

	const client = getEncoreClient(token);

	const [comicsRes, sessionsRes] = await Promise.allSettled([
		client.comics.MyComics(),
		client.upload.ListActiveSessions(),
	]);

	return {
		tab,
		comics: comicsRes.status === 'fulfilled' ? (comicsRes.value.comics || []) : [],
		activeSessions: sessionsRes.status === 'fulfilled' ? (sessionsRes.value.sessions || []) : [],
	};
};
