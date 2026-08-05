<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { currentUser } from '$lib/stores/auth';

	let series = $state<any>(null);
	let comics = $state<any[]>([]);
	let loading = $state(true);
	let following = $state(false);

	const id = $derived($page.params.id);
	const user = $derived($currentUser);

	onMount(async () => {
		try {
			series = await api.get(`/series/${id}`);
			const res = await api.get<{ comics: any[] }>(`/series/${id}/comics`);
			comics = res.comics;
		} catch {}
		loading = false;
	});

	async function toggleFollow() {
		if (following) {
			await api.delete(`/series/${id}/follow`);
			following = false;
		} else {
			await api.post(`/series/${id}/follow`);
			following = true;
		}
	}
</script>

<svelte:head><title>{series?.title || 'Series'} — Comics Galore</title></svelte:head>

<section class="py-8">
	{#if loading}
		<div class="animate-pulse space-y-4">
			<div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
			<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
		</div>
	{:else if series}
		<div class="mb-8">
			<a href="/series" class="text-sm text-muted-foreground hover:text-foreground">&larr; All series</a>
			<h1 class="text-3xl font-bold mt-2">{series.title}</h1>
			<p class="text-muted-foreground mt-1">{series.description}</p>
			{#if user}
				<Button size="sm" variant="outline" class="mt-3" onclick={toggleFollow}>
					{following ? 'Unfollow' : 'Follow'} Series
				</Button>
			{/if}
		</div>

		<h2 class="text-xl font-semibold mb-4">{comics.length} Issues</h2>
		{#if comics.length > 0}
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
				{#each comics as comic}
					<div class="relative">
						<ComicCard {...comic} />
						{#if comic.series_order}
							<span class="absolute top-2 left-2 bg-black/70 text-white text-xs px-2 py-0.5 rounded">#{comic.series_order}</span>
						{/if}
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-muted-foreground text-center py-8">No comics in this series yet.</p>
		{/if}
	{:else}
		<p class="text-destructive text-center py-20">Series not found.</p>
	{/if}
</section>
