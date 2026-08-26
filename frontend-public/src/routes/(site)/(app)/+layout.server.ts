import { redirect } from '@sveltejs/kit';
import { resolveUser } from '$lib/server/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const user = await resolveUser(cookies);
	if (!user) {
		throw redirect(302, '/');
	}
	return { user };
};
