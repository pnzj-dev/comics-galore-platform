import { resolveUser, getUserPreferences } from '$lib/server/session';
import { detectLocale } from '$lib/i18n/detect';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals, cookies, request }) => {
	// Priority: user's preferred language → locale cookie → Accept-Language.
	const prefs = await getUserPreferences(locals);
	const userLocale = prefs?.language || cookies.get('locale') || undefined;
	const locale = detectLocale(
		request.headers.get('accept-language') ?? undefined,
		userLocale,
	);
	const user = await resolveUser(locals);
	return { user, locale };
};
