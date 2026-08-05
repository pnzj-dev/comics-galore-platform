<script lang="ts">
	import PlanGrid from '$lib/components/PlanGrid.svelte';
	import { currentUser } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import SubscribeButton from '$lib/components/SubscribeButton.svelte';

	const user = $derived($currentUser);
</script>

<svelte:head>
	<title>Pricing - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-2 text-center">Plans & Pricing</h1>
	<p class="text-muted-foreground mb-8 text-center">Choose a plan to unlock premium features. Pay with crypto.</p>

	<PlanGrid mode="page" />

	<div class="text-center mt-8">
		{#if user}
			<SubscribeButton class="px-8" />
		{:else}
			<Button size="lg" href="/register">Register to Subscribe</Button>
		{/if}
	</div>
</section>
