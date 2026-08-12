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
	if (!user || user.role !== 'admin') {
		cookies.delete('token', { path: '/' });
		throw redirect(302, '/login');
	}

	return { user };
};
