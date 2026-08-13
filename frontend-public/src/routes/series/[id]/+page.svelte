<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { currentUser } from '$lib/stores/auth';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let comics = $state(data.comics);
	let following = $state(false);

	const user = $derived($currentUser);
	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));

	// Missing-issue gaps: expected contiguous series_order 1..max, flag absent ones.
	const maxOrder = $derived(comics.reduce((m, c) => Math.max(m, c.series_order || 1), 0));
	const missingOrders = $derived.by(() => {
		const present = new Set(comics.map((c) => c.series_order || 1));
		const gaps: number[] = [];
		for (let i = 1; i < maxOrder; i++) {
			if (!present.has(i)) gaps.push(i);
		}
		return gaps;
	});
	const progressPct = $derived(comics.length > 0 ? Math.round((data.readCount / comics.length) * 100) : 0);

	async function toggleFollow() {
		if (following) {
			await encore.comics.UnfollowSeries(data.series.id);
			following = false;
		} else {
			await encore.comics.FollowSeries(data.series.id);
			following = true;
		}
	}

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true });
	}
</script>

<svelte:head><title>{data.series?.title || 'Series'} — Comics Galore</title></svelte:head>

<section class="py-8">
	{#if data.series}
		<div class="mb-8">
			<a href="/series" class="text-sm text-muted-foreground hover:text-foreground">&larr; All series</a>
			<h1 class="text-3xl font-bold mt-2">{data.series.title}</h1>
			<p class="text-muted-foreground mt-1">{data.series.description}</p>

			{#if user}
				<div class="flex items-center gap-3 mt-3">
					<Button size="sm" variant="outline" onclick={toggleFollow}>
						{following ? 'Unfollow' : 'Follow'} Series
					</Button>
					{#if comics.length > 0}
						<div class="flex items-center gap-2">
							<div class="w-32 h-2 rounded-full bg-muted overflow-hidden">
								<div class="h-full bg-primary transition-all" style="width:{progressPct}%"></div>
							</div>
							<span class="text-xs text-muted-foreground">{data.readCount} of {comics.length} read</span>
						</div>
					{/if}
				</div>
			{/if}

			{#if missingOrders.length > 0}
				<p class="text-xs text-amber-600 dark:text-amber-400 mt-2">
					Missing issues: {missingOrders.map((o) => `#${o}`).join(', ')}
				</p>
			{/if}
		</div>

		<h2 class="text-xl font-semibold mb-4">{data.total} Issues</h2>
		{#if comics.length > 0}
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
				{#each comics as comic}
					<div class="relative">
						<ComicCard {...comic} />
						{#if comic.series_order}
							<span class="absolute top-2 left-2 bg-black/70 text-white text-xs px-2 py-0.5 rounded">#{comic.series_order}</span>
						{/if}
						{#if data.progress[comic.id]?.completed}
							<span class="absolute top-2 right-2 bg-green-600 text-white text-[10px] px-1.5 py-0.5 rounded">Read</span>
						{/if}
					</div>
				{/each}
			</div>
			<Pagination page={data.page} {totalPages} onPage={goPage} />
		{:else}
			<p class="text-muted-foreground text-center py-8">No comics in this series yet.</p>
		{/if}
	{:else}
		<p class="text-destructive text-center py-20">Series not found.</p>
	{/if}
</section>
