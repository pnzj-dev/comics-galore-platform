<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import CheckoutModal from '$lib/components/CheckoutModal.svelte';
	import { currentUser } from '$lib/stores/auth';

	let { class: className = '' }: { class?: string } = $props();
	let checkoutOpen = $state(false);
	let plansReady = $state(false);
	const user = $derived($currentUser);

	onMount(async () => {
		try {
			const res = await api.get<{ complete: boolean }>('/plans/ready');
			plansReady = res.complete;
		} catch {}
	});
</script>

{#if user && plansReady}
	<Button class={className} onclick={() => checkoutOpen = true}>Subscribe</Button>
	<CheckoutModal open={checkoutOpen} onClose={() => checkoutOpen = false} />
{/if}
