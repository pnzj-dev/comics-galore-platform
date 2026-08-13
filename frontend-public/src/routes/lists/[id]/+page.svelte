<script lang="ts">
	import ComicCard from '$lib/components/ComicCard.svelte';

	let { data } = $props();
</script>

<svelte:head><title>{data.list?.name || 'Reading List'} - Comics Galore</title></svelte:head>

<section class="py-8">
	{#if data.list}
		<h1 class="text-3xl font-bold mb-6">{data.list.name}</h1>
		{#if data.comics.length === 0}
			<p class="text-muted-foreground">This list is empty.</p>
		{:else}
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
				{#each data.comics as comic}
					<ComicCard {...comic} />
				{/each}
			</div>
		{/if}
	{:else}
		<p class="text-destructive text-center py-20">List not found.</p>
	{/if}
</section>
