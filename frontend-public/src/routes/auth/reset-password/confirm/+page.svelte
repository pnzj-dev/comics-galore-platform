<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let password = $state('');
	let confirm = $state('');
	let success = $state(false);
	let error = $state('');
	let loading = $state(false);

	const token = $derived($page.url.searchParams.get('token') || '');

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault(); error = '';
		if (password !== confirm) { error = 'Passwords do not match'; return; }
		if (password.length < 8) { error = 'Password must be at least 8 characters'; return; }
		loading = true;
		try {
			await api.post('/auth/password-reset/confirm', { token, password });
			success = true;
		} catch (err) {
			error = (err as Error).message;
		} finally {
			loading = false;
		}
	}
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
				<form onsubmit={handleSubmit} class="space-y-4">
					<div class="space-y-2">
						<Label for="password">New Password</Label>
						<Input id="password" type="password" bind:value={password} required placeholder="Min 8 characters" />
					</div>
					<div class="space-y-2">
						<Label for="confirm">Confirm Password</Label>
						<Input id="confirm" type="password" bind:value={confirm} required placeholder="Repeat password" />
					</div>
					{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
					<Button type="submit" class="w-full" disabled={loading}>{loading ? 'Resetting...' : 'Reset Password'}</Button>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
