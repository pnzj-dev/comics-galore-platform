import { redirect } from '@sveltejs/kit';
import { resolveUser } from '$lib/server/session';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	const user = await resolveUser(locals);
	if (!user) {
		throw redirect(302, '/');
	}
	return { user };
};
