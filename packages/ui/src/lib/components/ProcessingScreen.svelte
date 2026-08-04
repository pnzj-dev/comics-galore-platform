<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { subscriptionId, onSuccess, onRetry }: {
		subscriptionId: string;
		onSuccess: () => void;
		onRetry: () => void;
	} = $props();

	let elapsed = $state(0);
	let timeout = $state(false);
	let interval: ReturnType<typeof setInterval>;
	const MAX_SECONDS = 300; // 5 minutes
	const POLL_MS = 3000;

	onMount(() => {
		interval = setInterval(async () => {
			elapsed += POLL_MS / 1000;
			if (elapsed >= MAX_SECONDS) {
				clearInterval(interval);
				timeout = true;
				return;
			}

			try {
				const res = await api.get<{ active: boolean }>(`/billing/subscription/${subscriptionId}/poll`);
				if (res.active) {
					clearInterval(interval);
					onSuccess();
				}
			} catch {
				// keep polling
			}
		}, POLL_MS);

		return () => clearInterval(interval);
	});

	let remaining = $derived(MAX_SECONDS - elapsed);
</script>

<div class="space-y-6 text-center">
	<div class="flex justify-center">
		<svg class="animate-spin size-12 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
	</div>

	<h3 class="text-lg font-semibold">Processing Subscription</h3>
	<p class="text-sm text-muted-foreground">
		Please wait while your subscription is being processed...
	</p>
	<p class="text-xs text-muted-foreground">Time remaining: {Math.ceil(remaining)}s</p>

	{#if timeout}
		<div class="space-y-3">
			<p class="text-sm text-destructive">Subscription processing timed out.</p>
			<Button onclick={onRetry}>Retry</Button>
		</div>
	{/if}
</div>
