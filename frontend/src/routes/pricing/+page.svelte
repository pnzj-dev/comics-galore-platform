<script lang="ts">
	import CheckoutModal from '$lib/components/CheckoutModal.svelte';
	import { currentUser } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';

	let modalOpen = $state(false);
	const user = $derived($currentUser);
</script>

<svelte:head>
	<title>Pricing - Comics Galore</title>
</svelte:head>

<section class="py-8 text-center">
	<h1 class="text-3xl font-bold mb-2">Plans & Pricing</h1>
	<p class="text-muted-foreground mb-8">Choose a plan to unlock premium features. Pay with crypto.</p>

	{#if user}
		<Button size="lg" onclick={() => modalOpen = true}>View Plans</Button>
	{:else}
		<Button size="lg" href="/register">Register to Subscribe</Button>
	{/if}

	<CheckoutModal open={modalOpen} onClose={() => modalOpen = false} />
</section>
