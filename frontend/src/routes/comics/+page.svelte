<script lang="ts">
	import { api } from '$lib/api/client';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { onMount } from 'svelte';

	let comics = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await api.get<{ comics: any[] }>('/comics');
			comics = res.comics;
		} catch { /* empty */ }
		loading = false;
	});
</script>

<svelte:head>
	<title>Comics - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Browse Comics</h1>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each Array(8) as _}
				<div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden animate-pulse">
					<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
					<div class="p-3 space-y-2">
						<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3"></div>
					</div>
				</div>
			{/each}
		</div>
	{:else if comics.length === 0}
		<div class="text-center py-20">
			<p class="text-lg text-muted-foreground">No comics published yet.</p>
			<p class="text-sm text-muted-foreground mt-2">Be the first to upload!</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
	{/if}
</section>
