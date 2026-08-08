<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let status = $state<'loading' | 'success' | 'error'>('loading');
	let message = $state('');

	const token = $derived($page.url.searchParams.get('token') || '');

	onMount(async () => {
		if (!token) { status = 'error'; message = 'No verification token provided.'; return; }
		try {
			await api.post('/auth/verify-email', { token });
			status = 'success';
			message = 'Email verified successfully! You can now log in.';
		} catch (err) {
			status = 'error';
			message = (err as Error).message || 'Invalid or expired verification token.';
		}
	});
</script>

<svelte:head><title>Verify Email — Comics Galore</title></svelte:head>

<div class="flex min-h-[60vh] items-center justify-center p-4">
	<div class="text-center max-w-sm">
		{#if status === 'loading'}
			<div class="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin mx-auto mb-4"></div>
			<p class="text-muted-foreground">Verifying your email...</p>
		{:else if status === 'success'}
			<div class="text-green-500 text-4xl mb-4">✓</div>
			<p class="text-lg font-semibold">{message}</p>
			<Button class="mt-4" href="/login">Go to Login</Button>
		{:else}
			<div class="text-destructive text-4xl mb-4">✕</div>
			<p class="text-destructive">{message}</p>
			<Button class="mt-4" variant="outline" href="/login">Back to Login</Button>
		{/if}
	</div>
</div>
