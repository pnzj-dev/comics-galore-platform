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

	it('honors an enabled user locale', () => {
		expect(detectLocale(undefined, 'ja')).toBe('ja');
	});

	it('parses Accept-Language header for an enabled locale', () => {
		expect(detectLocale('ja-JP,ja;q=0.9', null)).toBe('ja');
		expect(detectLocale('pt-BR,pt;q=0.9', null)).toBe('pt-BR');
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

	it('translates to a non-en locale', () => {
		state.locale = 'ja';
		expect(t('nav.browse')).toBe('閲覧');
	});

	it('falls back to en for a missing non-en key', () => {
		state.locale = 'ja';
		expect(t('footer.dmca')).toBe('DMCA');
	});

	it('setLocale applies enabled locales', () => {
		setLocale('ja');
		expect(state.locale).toBe('ja');
	});

	it('setLocale rejects disabled locales', () => {
		setLocale('en');
		setLocale('xx' as any);
		expect(state.locale).toBe('en');
	});

	it('initializeLocale applies enabled locale', () => {
		initializeLocale('fr');
		expect(state.locale).toBe('fr');
	});
});
