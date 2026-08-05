<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import ComicCard from '$lib/components/ComicCard.svelte';

	let comics = $state<any[]>([]);
	let loading = $state(true);

	const tag = $derived($page.params.tag);

	onMount(async () => {
		try {
			const res = await api.get<{ comics: any[] }>(`/comics?tag=${encodeURIComponent(tag)}&limit=20`);
			comics = res.comics;
		} catch {}
		loading = false;
	});
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

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each Array(8) as _}
				<div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden animate-pulse">
					<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
					<div class="p-3 space-y-2">
						<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
					</div>
				</div>
			{/each}
		</div>
	{:else if comics.length === 0}
		<p class="text-muted-foreground text-center py-20">No comics found with this tag.</p>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
	{/if}
</section>
