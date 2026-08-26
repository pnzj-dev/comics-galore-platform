<script lang="ts">
	import { page } from '$app/stores';
	import { encore } from '$lib/api/encore';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client as zodClient } from 'sveltekit-superforms/adapters';
	import { resetPasswordSchema } from '$lib/schemas/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import PasswordInput from '$lib/components/auth/PasswordInput.svelte';
	import PasswordStrengthMeter from '$lib/components/auth/PasswordStrengthMeter.svelte';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let success = $state(false);
	let error = $state('');
	let loading = $state(false);

	const token = $derived($page.url.searchParams.get('token') || '');

	const { form, errors, enhance } = superForm(
		{ password: '', confirm: '' },
		{
			SPA: true,
			validationMethod: 'submit-only',
			validators: zodClient(resetPasswordSchema),
			onUpdate: async ({ form: f }) => {
				if (!f.valid) return;
				error = '';
				loading = true;
				try {
					await encore.auth.ConfirmPasswordReset({ token, password: f.data.password });
					success = true;
				} catch (err) {
					error = (err as Error).message;
				} finally {
					loading = false;
				}
			},
		},
	);
</script>

<svelte:head><title>Set New Password — Comics Galore</title></svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Set New Password</CardTitle>
			<CardDescription>{success ? 'Password updated!' : 'Enter your new password'}</CardDescription>
		</CardHeader>
		<CardContent>
			{#if success}
				<p class="text-sm text-muted-foreground text-center py-4">Your password has been reset successfully.</p>
				<Button class="w-full" href="/login">Go to Login</Button>
			{:else if !token}
				<p class="text-sm text-destructive text-center py-4">No reset token provided. Please use the link from your email.</p>
				<Button class="w-full" variant="outline" href="/login">Back to Login</Button>
			{:else}
				<form method="POST" use:enhance class="space-y-4">
					<div class="space-y-2">
						<Label for="password">New Password</Label>
						<PasswordInput id="password" bind:value={$form.password} placeholder="Min 8 characters" autocomplete="new-password" />
						<PasswordStrengthMeter password={$form.password} />
						{#if $errors.password}<p class="text-xs text-destructive">{$errors.password}</p>{/if}
					</div>
					<div class="space-y-2">
						<Label for="confirm">Confirm Password</Label>
						<PasswordInput id="confirm" bind:value={$form.confirm} placeholder="Repeat password" autocomplete="new-password" />
						{#if $errors.confirm}<p class="text-xs text-destructive">{$errors.confirm}</p>{/if}
					</div>
					{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
					<Button type="submit" class="w-full" disabled={loading}>{loading ? 'Resetting...' : 'Reset Password'}</Button>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
