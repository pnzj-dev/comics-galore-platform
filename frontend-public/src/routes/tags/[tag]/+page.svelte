<script lang="ts">
	import ComicCard from '$lib/components/ComicCard.svelte';

	let { data } = $props();

	const tag = $derived(data.tag);
	const comics = $derived(data.comics);
</script>

<svelte:head>
	<title>{tag} — Comics Galore</title>
</svelte:head>

<section class="py-8">
	<div class="mb-6">
		<a href="/comics" class="text-sm text-muted-foreground hover:text-foreground">&larr; All comics</a>
		<h1 class="text-3xl font-bold mt-2 capitalize">{tag}</h1>
		<p class="text-sm text-muted-foreground mt-1">{comics.length} comics tagged with &quot;{tag}&quot;</p>
	</div>

	{#if comics.length === 0}
		<p class="text-muted-foreground text-center py-20">No comics found with this tag.</p>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
	{/if}
</section>
