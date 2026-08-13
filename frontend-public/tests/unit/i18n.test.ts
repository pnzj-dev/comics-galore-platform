import { detectLocale } from '$lib/i18n/detect';
import { t, state, setLocale, initializeLocale } from '$lib/i18n/index.svelte';

describe('detectLocale', () => {
	it('defaults to en when no signal', () => {
		expect(detectLocale(undefined, null)).toBe('en');
		expect(detectLocale('', undefined)).toBe('en');
	});

	it('honors explicit user locale first', () => {
		expect(detectLocale('en-US,en;q=0.9', 'en')).toBe('en');
	});

	it('ignores disabled user locale and falls through', () => {
		// 'ja' is not enabled in v1.1 → falls back to default
		expect(detectLocale(undefined, 'ja')).toBe('en');
	});

	it('parses Accept-Language header for an enabled locale', () => {
		expect(detectLocale('en-US,en;q=0.9', null)).toBe('en');
	});

	it('falls back to en for not-yet-enabled locales', () => {
		// fr / pt-BR / zh-CN are not enabled until their packs land (v1.1 ships en only)
		expect(detectLocale('pt-BR,pt;q=0.9', null)).toBe('en');
		expect(detectLocale('zh-CN', null)).toBe('en');
		expect(detectLocale('fr-FR,fr;q=0.9', null)).toBe('en');
	});

	it('falls back to en for unknown languages', () => {
		expect(detectLocale('xx-YY', null)).toBe('en');
	});
});

describe('t', () => {
	beforeEach(() => {
		state.locale = 'en';
	});

	it('returns the en message', () => {
		expect(t('nav.browse')).toBe('Browse');
	});

	it('returns the key when message is missing', () => {
		expect(t('does.not.exist' as any)).toBe('does.not.exist');
	});

	it('interpolates params', () => {
		expect(t('common.welcome', { name: 'Ada' })).toBe('Welcome, Ada');
	});

	it('setLocale only applies enabled locales', () => {
		setLocale('ja' as any);
		expect(state.locale).toBe('en');
	});

	it('initializeLocale applies enabled locale', () => {
		initializeLocale('en');
		expect(state.locale).toBe('en');
	});
});
