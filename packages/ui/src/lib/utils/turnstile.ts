// Cloudflare Turnstile helper: sitekey + one-time script loader.

declare global {
	interface Window {
		turnstile?: {
			render(
				container: HTMLElement,
				options: {
					sitekey: string;
					action?: string;
					theme?: 'light' | 'dark' | 'auto';
					callback?: (token: string) => void;
					'expired-callback'?: () => void;
					'error-callback'?: () => void;
				}
			): string;
			reset(widgetId: string): void;
			getResponse(widgetId: string): string | undefined;
		};
	}
}

export const TURNSTILE_SITEKEY = ((import.meta.env.VITE_TURNSTILE_SITEKEY as string) || '').trim();

let scriptPromise: Promise<boolean> | null = null;

// loadTurnstile injects the Turnstile api.js script exactly once and resolves
// once window.turnstile is available. Returns false when the sitekey is unset
// or the script fails to load.
export function loadTurnstile(): Promise<boolean> {
	if (!TURNSTILE_SITEKEY) return Promise.resolve(false);
	if (!scriptPromise) {
		scriptPromise = new Promise((resolve) => {
			if (window.turnstile) {
				resolve(true);
				return;
			}
			const script = document.createElement('script');
			script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
			script.async = true;
			script.defer = true;
			script.onload = () => resolve(true);
			script.onerror = () => resolve(false);
			document.head.appendChild(script);
		});
	}
	return scriptPromise;
}
