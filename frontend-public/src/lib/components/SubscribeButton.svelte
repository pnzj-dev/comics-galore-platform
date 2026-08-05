<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import CheckoutModal from '$lib/components/CheckoutModal.svelte';
	import { currentUser } from '$lib/stores/auth';

	let { class: className = '' }: { class?: string } = $props();
	let checkoutOpen = $state(false);
	const user = $derived($currentUser);
</script>

{#if user}
	<Button class={className} onclick={() => checkoutOpen = true}>Subscribe</Button>
	<CheckoutModal open={checkoutOpen} onClose={() => checkoutOpen = false} />
{/if}
