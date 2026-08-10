<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import HeroComicCard from '$lib/components/HeroComicCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BookOpen } from 'lucide-svelte';

	let { data } = $props();

	const user = $derived($currentUser || data.authed);
	const authed = $derived(!!user);
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
		{#if !authed}
			<Button size="lg" href="/register">Get Started</Button>
			<Button size="lg" variant="outline" href="/login">Sign In</Button>
		{:else}
			<Button size="lg" href="/comics">Browse Comics</Button>
		{/if}
	</div>
</section>

<!-- Continue Reading -->
{#if authed && data.continueReading.length > 0}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold flex items-center gap-2"><BookOpen class="size-5" /> Continue Reading</h2>
		</div>
		<div class="flex gap-4 overflow-x-auto pb-2 -mx-1 px-1">
			{#each data.continueReading as comic}
				{@const progress = data.continueProgress[comic.id]}
				<div class="shrink-0 w-48">
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
{#if data.comicOfTheDay}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">Comic of the Day</h2>
		</div>
		<div class="max-w-3xl">
			<HeroComicCard {...data.comicOfTheDay} />
		</div>
	</section>
{/if}

<!-- Latest Comics -->
<section class="py-8">
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-xl font-semibold">Latest Comics</h2>
		<a href="/comics" class="text-sm text-primary hover:underline">View all</a>
	</div>
	{#if data.latestComics.length === 0}
		<p class="text-muted-foreground text-center py-8">No comics published yet. Be the first!</p>
	{:else}
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
			{#each data.latestComics as comic}<ComicCard {...comic} />{/each}
		</div>
	{/if}
</section>

<!-- Popular This Month -->
{#if data.popularComics.length > 0}
	<section class="py-8">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-xl font-semibold">Popular This Month</h2>
			<a href="/comics?sort=popular" class="text-sm text-primary hover:underline">View all</a>
		</div>
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
			{#each data.popularComics as comic}<ComicCard {...comic} />{/each}
		</div>
	</section>
{/if}
