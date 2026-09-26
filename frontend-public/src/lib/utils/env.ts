// Baked-in build environment (VITE_ENV set by the CI/Dockerfile; local dev sets it in .env).
export type EnvName = 'dev' | 'staging' | 'local' | 'prod';

const raw = ((import.meta.env.VITE_ENV as string) || '').trim().toLowerCase();

export const ENV: EnvName | null =
	raw === 'dev' || raw === 'development'
		? 'dev'
		: raw === 'staging'
			? 'staging'
			: raw === 'local'
				? 'local'
				: raw === 'prod' || raw === 'production'
					? 'prod'
					: null;

export const IS_NON_PROD = ENV !== null && ENV !== 'prod';

export const ENV_LABEL = ENV && ENV !== 'prod' ? ENV.toUpperCase() : '';

export const ENV_TOOLTIP =
	ENV === 'dev'
		? 'Development environment — not production'
		: ENV === 'staging'
			? 'Staging environment — not production'
			: ENV === 'local'
				? 'Local development environment'
				: '';

const TITLE_SUFFIX = ENV_LABEL ? ` · ${ENV_LABEL}` : '';

// Client-only side effects for non-prod: a console warning and a document.title
// suffix so the browser tab is distinguishable. Returns a cleanup function.
export function applyEnvironmentEffects(): () => void {
	if (!IS_NON_PROD) return () => {};

	console.warn(
		`%cComics Galore ${ENV_LABEL} environment`,
		'background:#18181b;color:#fafafa;padding:3px 8px;border-radius:4px;font-weight:700',
		'Not production — data here may be reset at any time.',
	);

	const applyTitle = () => {
		if (document.title.endsWith(TITLE_SUFFIX)) return;
		document.title = `${document.title}${TITLE_SUFFIX}`;
	};
	applyTitle();

	const observer = new MutationObserver(applyTitle);
	observer.observe(document.head, { subtree: true, childList: true, characterData: true });
	return () => observer.disconnect();
}
