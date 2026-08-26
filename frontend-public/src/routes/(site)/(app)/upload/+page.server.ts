import { redirect } from '@sveltejs/kit';
import { resolveUser } from '$lib/server/session';
import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const token = cookies.get('token');
	if (!token) throw redirect(302, '/login');

	const user = await resolveUser(cookies);
	if (!user || (user.role !== 'uploader' && user.role !== 'admin')) throw redirect(302, '/');

	const rawTab = url.searchParams.get('tab') ?? 'list';
	const validTabs = ['list', 'manual', 'archive'] as const;
	const tab = validTabs.includes(rawTab as typeof validTabs[number]) ? rawTab : 'list';

	const client = getEncoreClient(token);

	const [comicsRes, sessionsRes, configRes] = await Promise.allSettled([
		client.comics.MyComics(),
		client.upload.ListActiveSessions(),
		client.auth.GetSiteConfig(),
	]);

	return {
		tab,
		comics: comicsRes.status === 'fulfilled' ? (comicsRes.value.comics || []) : [],
		activeSessions: sessionsRes.status === 'fulfilled' ? (sessionsRes.value.sessions || []) : [],
		uploadMode: configRes.status === 'fulfilled' && configRes.value.upload_mode === 'direct' ? 'direct' : 'backend',
		pagePreviewThreshold:
			configRes.status === 'fulfilled' && configRes.value.page_preview_threshold > 0
				? configRes.value.page_preview_threshold
				: 20,
		uploadPartSizeMB:
			configRes.status === 'fulfilled' && configRes.value.upload_part_size_mb > 0
				? configRes.value.upload_part_size_mb
				: 100,
		uploadConcurrency:
			configRes.status === 'fulfilled' && configRes.value.upload_concurrency > 0
				? configRes.value.upload_concurrency
				: 4,
	};
};
