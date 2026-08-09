import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const token = cookies.get('token');
	if (!token) {
		throw redirect(302, '/login');
	}
	try {
		const payload = JSON.parse(atob(token.split('.')[1]));
		if (!payload.role) {
			throw redirect(302, '/login');
		}
	} catch {
		throw redirect(302, '/login');
	}
	return {};
};
