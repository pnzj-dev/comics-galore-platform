// Reactive i18n store (Svelte 5 runes) + translation helper.
//
// `state.locale` is a runes `$state`; `t()` reads it reactively so components
// that call `t('key')` re-render when the locale changes. Catalogs are keyed
// by locale with English as the always-present fallback.

import { isEnabledLocale, type Locale } from './locales';
import en from './messages/en';
import ja from './messages/ja';
import es from './messages/es';
import ko from './messages/ko';
import fr from './messages/fr';
import ptBR from './messages/pt-BR';
import zhCN from './messages/zh-CN';
import de from './messages/de';
import it from './messages/it';
import id from './messages/id';

export type { Locale } from './locales';
export { DEFAULT_LOCALE, ENABLED_LOCALES, PRIORITY_LOCALES, LOCALE_META, isEnabledLocale } from './locales';
export { detectLocale } from './detect';

export type MessageKey = keyof typeof en;
type Params = Record<string, string | number>;

const catalogs: Record<string, Partial<Record<MessageKey, string>>> = {
	en,
	ja,
	es,
	ko,
	fr,
	'pt-BR': ptBR,
	'zh-CN': zhCN,
	de,
	it,
	id,
};

export function registerCatalog(locale: Locale, messages: Partial<Record<MessageKey, string>>) {
	catalogs[locale] = messages;
}

export const state = $state<{ locale: Locale }>({ locale: 'en' });

export function setLocale(next: Locale) {
	if (!isEnabledLocale(next)) return;
	state.locale = next;
	if (typeof document !== 'undefined') {
		document.documentElement.lang = next;
		// Persist so the server resolves the same locale on subsequent SSR loads.
		document.cookie = `locale=${next}; path=/; SameSite=Lax; max-age=2592000`;
	}
}

export function initializeLocale(localeCode: Locale) {
	if (isEnabledLocale(localeCode)) state.locale = localeCode;
}

/**
 * Translate `key` for the active locale. Interpolates `{param}` placeholders.
 * Falls back to English, then to the raw key, if the message is missing.
 */
export function t(key: MessageKey, params?: Params): string {
	let msg = catalogs[state.locale]?.[key] ?? en[key] ?? key;
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			msg = msg.replaceAll(`{${k}}`, String(v));
		}
	}
	return msg;
}
