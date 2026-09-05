<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import SeriesCard from '$lib/components/home/SeriesCard.svelte';
	import Pagination from '$lib/components/common/Pagination.svelte';
	import { onDestroy, untrack } from 'svelte';

	let { data } = $props();

	let resultsRef = $state<HTMLDivElement | null>(null);
	// svelte-ignore state_referenced_locally
	let searchInput = $state(data.search);
	let searchTimer = $state<ReturnType<typeof setTimeout> | undefined>(undefined);

	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));
	const category = $derived(data.category);
	const searchField = $derived(data.searchField);

	// Keep the input in sync when the URL changes (back/forward, category nav).
	$effect(() => {
		if (untrack(() => searchInput) !== data.search) {
			searchInput = data.search;
		}
	});

	onDestroy(() => {
		if (searchTimer) clearTimeout(searchTimer);
	});

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
		searchInput = value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			updateParams({ search: value || undefined });
		}, 300);
	}

	function changeField(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		updateParams({ search_field: value || undefined });
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
		<div class="flex flex-col sm:flex-row gap-3">
			<div class="relative flex-1">
				<div class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none">
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="size-4"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
				</div>
				<input
					type="text"
					value={searchInput}
					oninput={onSearch}
					placeholder="Search by title or description…"
					class="w-full pl-10 pr-4 py-2.5 bg-background border border-input rounded-lg text-sm placeholder-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
				/>
			</div>
			<select
				value={searchField}
				onchange={changeField}
				class="px-3 py-2.5 bg-background border border-input rounded-lg text-sm text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary flex-shrink-0"
			>
				<option value="">All Fields</option>
				<option value="title">Title</option>
				<option value="description">Description</option>
			</select>
		</div>

		{#if data.categories.length > 0}
			<div class="flex flex-wrap gap-2">
				<button
					onclick={() => setCategory('')}
					class="text-xs px-2.5 py-1 rounded-full border transition-colors {category === ''
						? 'bg-primary text-primary-foreground border-primary'
						: 'bg-muted text-muted-foreground border-border hover:bg-primary/10 hover:text-primary'}"
				>
					All
				</button>
				{#each data.categories as cat}
					<button
						onclick={() => setCategory(cat)}
						class="text-xs px-2.5 py-1 rounded-full border transition-colors {category === cat
							? 'bg-primary text-primary-foreground border-primary'
							: 'bg-muted text-muted-foreground border-border hover:bg-primary/10 hover:text-primary'}"
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
					href="/series/{s.slug}"
					size="fluid"
				/>
			{/each}
		</div>
		<Pagination page={data.page} {totalPages} onPage={goPage} />
	{/if}
</section>
