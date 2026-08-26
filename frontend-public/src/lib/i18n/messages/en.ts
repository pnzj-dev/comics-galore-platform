// English message catalog. All keys are dot-namespaced and typed via
// `MessageKey` in `../index.ts` (derived from this file). This is the source of
// truth for which keys exist; other locale catalogs mirror the same keys.

const en = {
	// Navigation / chrome
	'nav.browse': 'Browse',
	'nav.series': 'Series',
	'nav.pricing': 'Pricing',
	'nav.upload': 'Upload',
	'nav.settings': 'Settings',
	'nav.logout': 'Logout',
	'nav.login': 'Login',
	'nav.register': 'Register',
	'nav.home': 'Home',

	// Account menu
	'menu.profile': 'Profile',
	'menu.preferences': 'Preferences',
	'menu.security': 'Security',

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

	// Series
	'series.trendingPopular': 'Trending & Popular Series',
	'series.trending': 'Trending',
	'series.popular': 'Popular',
	'series.viewAll': 'View all',
	'series.rankUp': 'Up {n} places',
	'series.rankDown': 'Down {n} places',

	// Comic detail
	'comic.signInToLike': 'Sign in to like',
	'comic.signInToDislike': 'Sign in to dislike',
	'comic.signInToFavorite': 'Sign in to favorite',
	'comic.upgradeForMorePreviews': 'Upgrade to unlock more previews',
	'comic.upgradeToRead': 'Upgrade to read',

	// Errors
	'error.notFound': 'Page not found.',
} as const;

export type Messages = typeof en;

export default en;
