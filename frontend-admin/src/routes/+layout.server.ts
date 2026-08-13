import { redirect } from '@sveltejs/kit';
import { decodeJWT } from '$lib/server/jwt';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	const token = cookies.get('token');
	if (!token) {
		if (url.pathname === '/login') return {};
		throw redirect(302, '/login');
	}

	const user = decodeJWT(token);
	if (!user || (user.role !== 'admin' && user.role !== 'moderator')) {
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
