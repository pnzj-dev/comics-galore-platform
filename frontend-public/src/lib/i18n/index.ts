// Barrel re-export. Runes state lives in `index.svelte.ts`; this file makes
// `$lib/i18n` resolvable as a plain module path.

export {
	state,
	setLocale,
	initializeLocale,
	t,
	registerCatalog,
	type MessageKey,
	type Locale,
	DEFAULT_LOCALE,
	ENABLED_LOCALES,
	PRIORITY_LOCALES,
	LOCALE_META,
	isEnabledLocale,
	detectLocale,
} from './index.svelte';
