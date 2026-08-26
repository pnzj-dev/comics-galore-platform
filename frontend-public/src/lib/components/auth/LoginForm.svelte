<script lang="ts">
	import { login, verifyTotpLogin } from '$lib/stores/auth';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client as zodClient } from 'sveltekit-superforms/adapters';
	import { loginSchema } from '$lib/schemas/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import OAuthButtons from '$lib/components/auth/OAuthButtons.svelte';
	import PasskeyLogin from '$lib/components/auth/PasskeyLogin.svelte';
	import AuthCard from './AuthCard.svelte';
	import PasswordInput from './PasswordInput.svelte';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { CardContent, CardFooter } from '$lib/components/ui/card/index.js';

	interface Props {
		onSwitchToRegister?: () => void;
		onForgotPassword?: () => void;
		onSuccess?: () => void;
	}

	let { onSwitchToRegister, onForgotPassword, onSuccess }: Props = $props();

	let error = $state('');
	let loading = $state(false);
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;

	// TOTP (2FA) step, shown only when the account has it enabled.
	let totpStep = $state(false);
	let mfaToken = $state('');
	let totpCode = $state('');
	let totpLoading = $state(false);

	const { form, errors, enhance } = superForm(
		{ email: '', password: '' },
		{
			SPA: true,
			validationMethod: 'submit-only',
			validators: zodClient(loginSchema),
			onUpdate: async ({ form: f }) => {
				if (!f.valid) return;
				error = '';
				loading = true;
				try {
					const result = await login(f.data.email, f.data.password, turnstileToken || '');
					if ('requires_totp' in result) {
						mfaToken = result.mfa_token;
						totpStep = true;
					} else {
						onSuccess?.();
					}
				} catch (err) {
					error = (err as Error).message;
				} finally {
					loading = false;
					turnstileToken = null;
					turnstileReset++;
				}
			},
		},
	);

	async function handleTotpSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		totpLoading = true;
		try {
			await verifyTotpLogin(mfaToken, totpCode);
			onSuccess?.();
		} catch (err) {
			error = (err as Error).message;
		} finally {
			totpLoading = false;
		}
	}

	function resetToPassword() {
		totpStep = false;
		totpCode = '';
		error = '';
	}

	$effect(() => {
		// Surface OAuth error query param (e.g. /login?error=oauth_failed).
		const params = new URLSearchParams(window.location.search);
		if (params.get('error')) {
			error = 'Sign-in was cancelled or failed. Please try again.';
		}
	});
</script>

<AuthCard>
	<CardContent class="space-y-4">
		{#if totpStep}
			<form onsubmit={handleTotpSubmit} class="space-y-4">
				<div class="space-y-1.5">
					<Label for="totp-code">Verification code</Label>
					<Input
						id="totp-code"
						bind:value={totpCode}
						inputmode="numeric"
						maxlength={6}
						placeholder="123456"
						autocomplete="one-time-code"
						required
					/>
					<p class="text-xs text-muted-foreground">Enter the 6-digit code from your authenticator app.</p>
				</div>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<Button type="submit" class="w-full" disabled={totpLoading}>
					{totpLoading ? 'Verifying…' : 'Verify'}
				</Button>

				<button type="button" class="w-full text-center text-xs text-muted-foreground hover:text-foreground transition-colors" onclick={resetToPassword}>
					Back to sign in
				</button>
			</form>
		{:else}
			<p class="text-center text-sm text-muted-foreground">Sign in to continue</p>

			<form method="POST" use:enhance class="space-y-4">
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input id="email" type="email" bind:value={$form.email} placeholder="you@example.com" />
					{#if $errors.email}<p class="text-xs text-destructive">{$errors.email}</p>{/if}
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<PasswordInput id="password" bind:value={$form.password} placeholder="••••••••" autocomplete="current-password" />
					{#if $errors.password}<p class="text-xs text-destructive">{$errors.password}</p>{/if}
				</div>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<Turnstile action="login" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />

				<Button type="submit" class="w-full" disabled={loading || (turnstileRequired && !turnstileToken)}>
					{loading ? 'Signing in...' : 'Sign in'}
				</Button>

				<button type="button" class="w-full text-center text-xs text-muted-foreground hover:text-primary underline underline-offset-4 transition-colors" onclick={() => onForgotPassword?.()}>
					Forgot password?
				</button>
			</form>

			<div class="flex items-center gap-3 text-xs text-muted-foreground">
				<Separator class="flex-1" />
				or continue with
				<Separator class="flex-1" />
			</div>

			<PasskeyLogin onSuccess={onSuccess} onError={(m) => (error = m)} />
			<OAuthButtons onError={(m) => (error = m)} />
		{/if}
	</CardContent>
	<CardFooter class="flex justify-center">
		<p class="text-sm text-muted-foreground">
			Don't have an account?
			<button type="button" class="text-primary underline underline-offset-4 hover:no-underline" onclick={() => onSwitchToRegister?.()}>Register</button>
		</p>
	</CardFooter>
</AuthCard>
