import { redirect } from '@sveltejs/kit';
import { resolveUser } from '$lib/server/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	const user = await resolveUser(cookies);

	if (!user) {
		if (url.pathname === '/login') return {};
		throw redirect(302, '/login');
	}

	if (user.role !== 'admin' && user.role !== 'moderator') {
		cookies.delete('token', { path: '/' });
		throw redirect(302, '/login');
	}

	if (
		user.role === 'moderator' &&
		url.pathname !== '/moderation' &&
		!url.pathname.startsWith('/moderation/')
	) {
		throw redirect(302, '/moderation');
	}

	return { user };
};
