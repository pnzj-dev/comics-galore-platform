<script lang="ts">
	import { currentUser, hydrated } from '$lib/stores/auth';
	import ComicCard from '$lib/components/comics/ComicCard.svelte';
	import TrendingPopularSeries from '$lib/components/home/TrendingPopularSeries.svelte';
	import ComicsHome from '$lib/components/home/ComicsHome.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BookOpen } from 'lucide-svelte';
	import { modal } from '$lib/stores/modal.svelte';

	let { data } = $props();

	const user = $derived($hydrated ? $currentUser : data.authed);
	const authed = $derived(!!user);

	// Home category/day state (client-side filtering of the already-loaded lists).
	let activeCategory = $state('');
	let activeDay = $state('mon');

	const homeCategories = $derived(data.home?.categories || []);
	const homePopular = $derived(data.home?.popular_by_category || []);
	const filteredPopular = $derived(
		activeCategory ? homePopular.filter((s) => s.category === activeCategory) : homePopular,
	);
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
			<Button size="lg" onclick={() => modal.open('register')}>Get Started</Button>
			<Button size="lg" variant="outline" onclick={() => modal.open('login')}>Sign In</Button>
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

<!-- Trending & Popular Series -->
{#if data.trendingSeries.length > 0 || data.popularSeries.length > 0}
	<TrendingPopularSeries
		trending={data.trendingSeries}
		popular={data.popularSeries}
		onViewAll={() => window.location.assign('/series')}
	/>
{/if}

<!-- Home sections (ad, categories, newly released, daily, indie) -->
{#if data.home}
	<ComicsHome
		ad={data.home.ad}
		categories={homeCategories}
		popular_by_category={filteredPopular}
		newly_released={data.home.newly_released}
		daily_series={data.home.daily_series}
		indie_series={data.home.indie_series}
		{activeCategory}
		{activeDay}
		onCategoryChange={(id) => (activeCategory = id)}
		onDayChange={(day) => (activeDay = day)}
		onViewAll={(section) => window.location.assign('/series')}
	/>
{/if}
