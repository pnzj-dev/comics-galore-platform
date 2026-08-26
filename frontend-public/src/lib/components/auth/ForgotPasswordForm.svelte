<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client as zodClient } from 'sveltekit-superforms/adapters';
	import { forgotPasswordSchema } from '$lib/schemas/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AuthCard from './AuthCard.svelte';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	interface Props {
		onSwitchToLogin?: () => void;
	}

	let { onSwitchToLogin }: Props = $props();

	let sent = $state(false);
	let error = $state('');
	let loading = $state(false);
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;

	const { form, errors, enhance } = superForm(
		{ email: '' },
		{
			SPA: true,
			validationMethod: 'submit-only',
			validators: zodClient(forgotPasswordSchema),
			onUpdate: async ({ form: f }) => {
				if (!f.valid) return;
				error = '';
				loading = true;
				try {
					await encore.auth.RequestPasswordReset({ email: f.data.email, turnstile_token: turnstileToken || '' });
					sent = true;
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
</script>

<AuthCard>
	<CardHeader class="text-center">
		<CardTitle>Forgot password</CardTitle>
		<CardDescription>{sent ? 'Check your email' : 'Enter your email to receive a reset link'}</CardDescription>
	</CardHeader>
	<CardContent>
		{#if sent}
			<p class="text-sm text-muted-foreground text-center py-4">If an account with that email exists, we've sent a password reset link. Check your inbox.</p>
			<Button class="w-full" variant="outline" onclick={() => onSwitchToLogin?.()}>Back to login</Button>
		{:else}
			<form method="POST" use:enhance class="space-y-4">
				<div class="space-y-2">
					<Label for="fp-email">Email</Label>
					<Input id="fp-email" type="email" bind:value={$form.email} placeholder="you@example.com" />
					{#if $errors.email}<p class="text-xs text-destructive">{$errors.email}</p>{/if}
				</div>
				{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
				<Turnstile action="password_reset" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />
				<Button type="submit" class="w-full" disabled={loading || (turnstileRequired && !turnstileToken)}>{loading ? 'Sending...' : 'Send reset link'}</Button>
			</form>
		{/if}
	</CardContent>
	<CardFooter class="flex justify-center">
		<button type="button" class="text-sm text-primary underline underline-offset-4 hover:no-underline" onclick={() => onSwitchToLogin?.()}>Back to login</button>
	</CardFooter>
</AuthCard>
