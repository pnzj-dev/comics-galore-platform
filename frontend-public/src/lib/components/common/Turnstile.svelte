<script lang="ts">
	import { onMount } from 'svelte';
	import { TURNSTILE_SITEKEY, loadTurnstile } from '$lib/utils/turnstile';

	let {
		action = 'submit',
		onToken,
		resetSignal = 0
	}: {
		action?: string;
		onToken: (token: string | null) => void;
		resetSignal?: number;
	} = $props();

	let container: HTMLDivElement | undefined = $state();
	let widgetId: string | undefined;

	onMount(() => {
		if (!TURNSTILE_SITEKEY) return;
		loadTurnstile().then((ok) => {
			if (!ok || !container || !window.turnstile) return;
			widgetId = window.turnstile.render(container, {
				sitekey: TURNSTILE_SITEKEY,
				action,
				callback: (token) => onToken(token),
				'expired-callback': () => onToken(null),
				'error-callback': () => onToken(null)
			});
		});
	});

	$effect(() => {
		if (resetSignal > 0 && widgetId && window.turnstile) {
			window.turnstile.reset(widgetId);
			onToken(null);
		}
	});
</script>

{#if TURNSTILE_SITEKEY}
	<div bind:this={container} data-action={action}></div>
{/if}
