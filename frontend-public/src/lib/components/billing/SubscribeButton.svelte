<script lang="ts">
	import { onMount } from 'svelte';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { modal } from '$lib/stores/modal.svelte';
	import { currentUser } from '$lib/stores/auth';

	let { class: className = '' }: { class?: string } = $props();
	let plansReady = $state(false);
	const user = $derived($currentUser);

	onMount(async () => {
		try {
			const res = await encore.tiers.PlansReady();
			plansReady = res.complete;
		} catch {}
	});
</script>

{#if user && plansReady}
	<Button class={className} onclick={() => modal.open('checkout')}>Subscribe</Button>
{/if}
