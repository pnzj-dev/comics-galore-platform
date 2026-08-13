import { decodeJWT } from '$lib/server/jwt';
import { detectLocale } from '$lib/i18n/detect';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, request }) => {
	const token = cookies.get('token');
	const locale = detectLocale(
		request.headers.get('accept-language') ?? undefined,
		cookies.get('locale'),
	);
	return { user: token ? decodeJWT(token) : null, locale };
};
