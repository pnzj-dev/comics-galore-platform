<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { currentUser, isAuthenticated, type User } from '$lib/stores/auth';
	import { onMount, onDestroy } from 'svelte';
	import {
		webauthnSupported,
		webauthnAutofillSupported,
		prepareRequestOptions,
		convertAssertion,
		mapWebauthnError,
	} from '$lib/webauthn';

	interface Props {
		onError?: (message: string) => void;
		onSuccess?: () => void;
	}

	let { onError, onSuccess }: Props = $props();

	let loading = $state(false);

	const supported = webauthnSupported();

	// Conditional (autofill) ceremony, tracked as a controller + promise pair so
	// the explicit "Continue with Passkey" click can abort it *and* await its
	// teardown before starting its own — the two `navigator.credentials.get()`
	// calls must never overlap (overlap throws "A request is already pending").
	let autofill = $state<{ controller: AbortController; promise: Promise<void> } | null>(null);

	// Attempt conditional (autofill) passkey login exactly once on mount.
	// Errors here are silent: a cancelled or absent autofill is not a
	// user-facing failure (only the explicit button surfaces errors).
	onMount(() => {
		if (!supported) return;
		webauthnAutofillSupported().then((v) => {
			if (!v || loading) return;
			startAutofill();
		});
	});

	// Release any pending conditional (autofill) request when the login form
	// unmounts, otherwise the browser keeps it "in flight" and later WebAuthn
	// ceremonies (e.g. registering a passkey) fail with "A request is already
	// pending".
	onDestroy(() => {
		autofill?.controller.abort();
		autofill = null;
	});

	function startAutofill() {
		const controller = new AbortController();
		const promise = (async () => {
			try {
				const optsRes = await encore.auth.PasskeyLoginOptions();
				const options = prepareRequestOptions(optsRes.options);
				const assertion = (await navigator.credentials.get({
					...options,
					mediation: 'conditional',
					signal: controller.signal,
				} as CredentialRequestOptions)) as PublicKeyCredential | null;

				if (!assertion) return;
				await finishLogin(assertion);
			} catch {
				// Autofill cancelled/absent — intentionally silent.
			} finally {
				if (autofill?.controller === controller) autofill = null;
			}
		})();
		autofill = { controller, promise };
	}

	async function doLogin() {
		if (loading) return;
		loading = true;
		try {
			// Abort any pending autofill ceremony and wait for it to settle so
			// the browser releases the pending request before we start a new one.
			if (autofill) {
				autofill.controller.abort();
				try {
					await autofill.promise;
				} catch {
					/* already swallowed by startAutofill */
				}
				autofill = null;
			}

			const optsRes = await encore.auth.PasskeyLoginOptions();
			const options = prepareRequestOptions(optsRes.options);
			const assertion = (await navigator.credentials.get({
				...options,
				mediation: 'optional',
			} as CredentialRequestOptions)) as PublicKeyCredential | null;

			if (!assertion) {
				onError?.('No passkey selected.');
				return;
			}
			await finishLogin(assertion);
		} catch (err) {
			onError?.(mapWebauthnError(err));
		} finally {
			loading = false;
		}
	}

	async function finishLogin(assertion: PublicKeyCredential) {
		const payload = convertAssertion(assertion);
		const res = await fetch('/auth/passkey/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload),
		});
		if (!res.ok) {
			const data = await res.json().catch(() => ({}));
			onError?.(data.message || 'Passkey login failed.');
			return;
		}
		const user = (await res.json()) as User;
		currentUser.set(user);
		isAuthenticated.set(true);
		onSuccess?.();
	}
</script>

{#if supported}
	<button
		type="button"
		onclick={doLogin}
		disabled={loading}
		class="w-full flex items-center justify-center gap-2 rounded-md border border-border bg-background px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
	>
		<svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
		{loading ? 'Signing in…' : 'Continue with Passkey'}
	</button>
{/if}
