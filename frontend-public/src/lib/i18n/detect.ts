// Pure locale utilities — safe to import from server load functions (no runes).

import { DEFAULT_LOCALE, isEnabledLocale, type Locale } from './locales';

/** Normalize a BCP 47 / Accept-Language tag to a supported locale code. */
function normalizeTag(tag: string): string {
	const t = tag.trim().toLowerCase();
	if (t.startsWith('pt')) return 'pt-BR';
	if (t.startsWith('zh')) return 'zh-CN';
	if (t.startsWith('ja')) return 'ja';
	if (t.startsWith('ko')) return 'ko';
	if (t.startsWith('es')) return 'es';
	if (t.startsWith('fr')) return 'fr';
	if (t.startsWith('de')) return 'de';
	if (t.startsWith('it')) return 'it';
	if (t.startsWith('id')) return 'id';
	if (t.startsWith('en')) return 'en';
	return t;
}

/**
 * Resolve the UI locale. Priority: explicit user override → Accept-Language
 * header → default (`en`). Always returns an enabled locale.
 */
export function detectLocale(acceptLanguage: string | undefined, userLocale?: string | null): Locale {
	if (userLocale && isEnabledLocale(userLocale)) return userLocale;

	if (acceptLanguage) {
		for (const part of acceptLanguage.split(',')) {
			const code = normalizeTag(part.split(';')[0]);
			if (isEnabledLocale(code)) return code;
		}
	}

	return DEFAULT_LOCALE;
}
