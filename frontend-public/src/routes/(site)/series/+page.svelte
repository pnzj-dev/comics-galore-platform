<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import SeriesCard from '$lib/components/home/SeriesCard.svelte';
	import Pagination from '$lib/components/common/Pagination.svelte';

	let { data } = $props();

	let resultsRef = $state<HTMLDivElement | null>(null);

	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));
	const category = $derived(data.category);

	function updateParams(updates: Record<string, string | undefined>) {
		const url = new URL(page.url);
		for (const [k, v] of Object.entries(updates)) {
			if (v) url.searchParams.set(k, v);
			else url.searchParams.delete(k);
		}
		url.searchParams.delete('page');
		goto(url.pathname + url.search, { keepFocus: true });
	}

	function onSearch(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		updateParams({ search: value || undefined });
	}

	function setCategory(cat: string) {
		updateParams({ category: cat || undefined });
	}

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true, noScroll: true }).then(() => {
			resultsRef?.scrollIntoView({ block: 'start' });
		});
	}
</script>

<svelte:head><title>Series — Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Browse Series</h1>

	<div class="rounded-xl border border-border bg-muted/30 p-4 mb-6 space-y-4">
		<div class="relative">
			<div class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="size-4"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
			</div>
			<input
				type="text"
				value={data.search}
				onchange={onSearch}
				placeholder="Search by title…"
				class="w-full pl-10 pr-4 py-2.5 bg-background border border-input rounded-lg text-sm placeholder-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
			/>
		</div>

		{#if data.categories.length > 0}
			<div class="flex flex-wrap gap-2">
				<button
					onclick={() => setCategory('')}
					class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition {category === ''
						? 'bg-primary text-primary-foreground'
						: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
				>
					All
				</button>
				{#each data.categories as cat}
					<button
						onclick={() => setCategory(cat)}
						class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition {category === cat
							? 'bg-primary text-primary-foreground'
							: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
					>
						{cat}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#if data.series.length === 0}
		<p class="text-muted-foreground text-center py-12">No series found.</p>
	{:else}
		<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 scroll-mt-16" bind:this={resultsRef}>
			{#each data.series as s}
				<SeriesCard
					id={s.id}
					title={s.title}
					cover_url={s.cover_url}
					genre={s.genre}
					views_count={s.views_count}
					hearts_count={s.hearts_count}
					href="/series/{s.id}"
					size="fluid"
				/>
			{/each}
		</div>
		<Pagination page={data.page} {totalPages} onPage={goPage} />
	{/if}
</section>
