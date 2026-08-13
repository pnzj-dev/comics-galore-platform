<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';

	let { data } = $props();

	const comics = $derived(data.comics);
	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true });
	}
</script>

<svelte:head><title>Favorites - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Favorites</h1>

	{#if comics.length === 0}
		<div class="text-center py-16">
			<p class="text-lg text-muted-foreground">No favorites yet.</p>
			<p class="text-sm text-muted-foreground mt-1">Tap the ★ on any comic to save it here.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
		<Pagination page={data.page} {totalPages} onPage={goPage} />
	{/if}
</section>
