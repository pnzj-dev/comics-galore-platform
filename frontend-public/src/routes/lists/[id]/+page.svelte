<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';

	let { data } = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true });
	}
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
			<Pagination page={data.page} {totalPages} onPage={goPage} />
		{/if}
	{:else}
		<p class="text-destructive text-center py-20">List not found.</p>
	{/if}
</section>
