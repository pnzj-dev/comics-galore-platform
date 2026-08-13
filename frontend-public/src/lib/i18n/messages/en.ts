// English message catalog. All keys are dot-namespaced and typed via
// `MessageKey` in `../index.ts` (derived from this file). This is the source of
// truth for which keys exist; other locale catalogs mirror the same keys.

const en = {
	// Navigation / chrome
	'nav.browse': 'Browse',
	'nav.pricing': 'Pricing',
	'nav.upload': 'Upload',
	'nav.settings': 'Settings',
	'nav.logout': 'Logout',
	'nav.login': 'Login',
	'nav.register': 'Register',
	'nav.home': 'Home',

	// Footer
	'footer.terms': 'Terms',
	'footer.privacy': 'Privacy',
	'footer.dmca': 'DMCA',

	// Common actions
	'common.loading': 'Loading…',
	'common.retry': 'Retry',
	'common.cancel': 'Cancel',
	'common.save': 'Save',
	'common.close': 'Close',
	'common.empty': 'Nothing here yet.',
	'common.welcome': 'Welcome, {name}',

	// Errors
	'error.notFound': 'Page not found.',
} as const;

export type Messages = typeof en;

export default en;
