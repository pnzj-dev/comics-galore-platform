<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	const user = $derived($currentUser);
	let latestComics = $state<any[]>([]);
	let popularComics = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const [latest, popular] = await Promise.all([
				api.get<{ comics: any[] }>('/comics?limit=4&sort=newest'),
				api.get<{ comics: any[] }>('/comics?limit=4&sort=popular')
			]);
			latestComics = latest.comics;
			popularComics = popular.comics;
		} catch {}
		loading = false;
	});
</script>

<svelte:head>
	<title>Comics Galore</title>
	<meta name="description" content="Discover, read, and share digital comics." />
</svelte:head>

<section class="flex flex-col items-center justify-center py-16 text-center">
	<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">Welcome to Comics Galore</h1>
	<p class="mt-4 text-lg text-muted-foreground max-w-2xl">
		Discover and read amazing comics from creators around the world.
	</p>
	<div class="mt-8 flex gap-4">
		{#if !user}
			<Button size="lg" href="/register">Get Started</Button>
			<Button size="lg" variant="outline" href="/login">Sign In</Button>
		{:else}
			<Button size="lg" href="/comics">Browse Comics</Button>
		{/if}
	</div>
</section>

<!-- Latest Comics -->
<section class="py-8">
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-xl font-semibold">Latest Comics</h2>
		<a href="/comics" class="text-sm text-primary hover:underline">View all</a>
	</div>
	{#if loading}
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
			{#each Array(4) as _}
				<div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden animate-pulse">
					<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
					<div class="p-3 space-y-2"><div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div><div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div></div>
				</div>
			{/each}
		</div>
	{:else if latestComics.length === 0}
		<p class="text-muted-foreground text-center py-8">No comics published yet. Be the first!</p>
	{:else}
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
			{#each latestComics as comic}<ComicCard {...comic} />{/each}
		</div>
	{/if}
</section>

<!-- Popular This Month -->
{#if popularComics.length > 0}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">Popular This Month</h2>
			<a href="/comics?sort=popular" class="text-sm text-primary hover:underline">View all</a>
		</div>
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
			{#each popularComics as comic}<ComicCard {...comic} />{/each}
		</div>
	</section>
{/if}
