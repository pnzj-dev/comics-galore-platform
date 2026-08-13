// Locale registry for Comics Galore UI (chrome) localization.
//
// UI locale is independent of comic `content_language` (ADR 0015). English is
// the default; additional locale packs land progressively (see v1-scope.md).

export type Locale = 'en' | 'ja' | 'es' | 'ko' | 'fr' | 'pt-BR' | 'zh-CN' | 'de' | 'it' | 'id';

export const DEFAULT_LOCALE: Locale = 'en';

// v1.1 ships English only; new packs enable their code here (and provide a
// matching catalog under `./messages/<locale>.ts`).
export const ENABLED_LOCALES: Locale[] = ['en'];

// Priority locales for engagement (ADR 0015). Ordering is informational.
export const PRIORITY_LOCALES: Locale[] = [
	'en',
	'ja',
	'es',
	'ko',
	'fr',
	'pt-BR',
	'zh-CN',
	'de',
	'it',
	'id',
];

export interface LocaleMeta {
	code: Locale;
	label: string; // native name
}

export const LOCALE_META: Record<Locale, LocaleMeta> = {
	en: { code: 'en', label: 'English' },
	ja: { code: 'ja', label: '日本語' },
	es: { code: 'es', label: 'Español' },
	ko: { code: 'ko', label: '한국어' },
	fr: { code: 'fr', label: 'Français' },
	'pt-BR': { code: 'pt-BR', label: 'Português (Brasil)' },
	'zh-CN': { code: 'zh-CN', label: '简体中文' },
	de: { code: 'de', label: 'Deutsch' },
	it: { code: 'it', label: 'Italiano' },
	id: { code: 'id', label: 'Bahasa Indonesia' },
};

export function isEnabledLocale(l: string): l is Locale {
	return (ENABLED_LOCALES as string[]).includes(l);
}
