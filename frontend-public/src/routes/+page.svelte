<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BookOpen } from 'lucide-svelte';

	const user = $derived($currentUser);
	const authed = $derived(!!user);
	let latestComics = $state<any[]>([]);
	let popularComics = $state<any[]>([]);
	let comicOfTheDay = $state<any>(null);
	let continueReading = $state<any[]>([]);
	let continueProgress = $state<Record<string, { current_page: number; total_pages: number }>>({});
	let loading = $state(true);

	onMount(async () => {
		try {
			const fetches: Promise<any>[] = [
				api.get<{ comics: any[] }>('/comics?limit=4&sort=newest'),
				api.get<{ comics: any[] }>('/comics?limit=4&sort=popular'),
				api.get<{ comics: any[] }>('/comics?limit=1&sort=random')
			];

			if (authed) {
				fetches.push(
					api.get<{ items: Array<{ comic_id: string; current_page: number; total_pages: number }> }>('/reading-continue')
				);
			}

			const results = await Promise.all(fetches);
			latestComics = results[0].comics;
			popularComics = results[1].comics;
			comicOfTheDay = results[2].comics[0] || null;

			if (authed && results[3]) {
				const readingItems = results[3].items || [];
				const prog: Record<string, { current_page: number; total_pages: number }> = {};
				readingItems.forEach((item: any) => {
					prog[item.comic_id] = { current_page: item.current_page, total_pages: item.total_pages };
				});
				continueProgress = prog;

				if (readingItems.length > 0) {
					const ids = readingItems.map((i: any) => i.comic_id);
					const batch = await api.post<{ comics: any[] }>('/comics-batch', { ids });
					continueReading = batch.comics || [];
				}
			}
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

<!-- Continue Reading -->
{#if authed && continueReading.length > 0}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold flex items-center gap-2"><BookOpen class="size-5" /> Continue Reading</h2>
		</div>
		<div class="flex gap-4 overflow-x-auto pb-2 -mx-1 px-1">
			{#each continueReading as comic}
				{@const progress = continueProgress[comic.id]}
				<div class="flex-shrink-0 w-48">
					<ComicCard {...comic} />
					{#if progress}
						<p class="text-xs text-muted-foreground text-center mt-1">Page {progress.current_page + 1} of {progress.total_pages}</p>
						<div class="mx-2 mt-1 h-1.5 bg-muted rounded-full overflow-hidden">
							<div class="h-full bg-primary rounded-full transition-all" style="width: {Math.round((progress.current_page + 1) / progress.total_pages * 100)}%"></div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	</section>
{/if}

<!-- Comic of the Day -->
{#if comicOfTheDay}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">Comic of the Day</h2>
		</div>
		<div class="max-w-sm mx-auto">
			<ComicCard {...comicOfTheDay} />
		</div>
	</section>
{/if}

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
