import { decodeJWT } from '$lib/server/jwt';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const token = cookies.get('token');
	return { user: token ? decodeJWT(token) : null };
};
