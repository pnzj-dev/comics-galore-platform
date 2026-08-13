<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { Label } from '$lib/components/ui/label/index.js';

	import { onDestroy, untrack } from 'svelte';

	let { data } = $props();

	const comics = $derived(data.comics);
	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));
	const langFilter = $derived(data.lang || '');
	// svelte-ignore state_referenced_locally
	let searchInput = $state(data.search);
	let searchTimer = $state<ReturnType<typeof setTimeout> | undefined>(undefined);
	const searchField = $derived(data.searchField);
	const activeTag = $derived(data.tag);

	// Keep the input in sync when the URL changes (back/forward, tag/lang nav).
	$effect(() => {
		if (untrack(() => searchInput) !== data.search) {
			searchInput = data.search;
		}
	});

	onDestroy(() => {
		if (searchTimer) clearTimeout(searchTimer);
	});

	const facets = $derived(data.facets || []);
	const popularTags = $derived(data.popularTags || []);

	const languages = $derived.by(() => {
		const codes = new Set(['']);
		for (const f of facets) codes.add(f.language);
		for (const c of ['en', 'ja', 'es', 'ko', 'fr', 'pt-BR', 'zh-CN', 'de', 'it', 'id']) codes.add(c);
		return [...codes];
	});
	const facetCount = $derived.by(() => {
		const map: Record<string, number> = {};
		for (const f of facets) map[f.language] = f.count;
		return map;
	});

	function langLabel(code: string): string {
		if (code === '') return 'All languages';
		return code + (facetCount[code] ? ` (${facetCount[code]})` : '');
	}

	function updateParams(updates: Record<string, string | undefined>) {
		const url = new URL(page.url);
		for (const [k, v] of Object.entries(updates)) {
			if (v) url.searchParams.set(k, v);
			else url.searchParams.delete(k);
		}
		url.searchParams.delete('page');
		goto(url.pathname + url.search, { keepFocus: true });
	}

	function changeLang(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		updateParams({ language: value || undefined });
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

	function toggleTag(tag: string) {
		updateParams({ tag: activeTag === tag ? undefined : tag });
	}

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true });
	}
</script>

<svelte:head>
	<title>Comics - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Browse Comics</h1>

	<!-- Full-width search + filters -->
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
					placeholder="Search by title, author, or description..."
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
				<option value="author">Author</option>
			</select>
			<div class="flex items-center gap-2 flex-shrink-0">
				<Label for="lang-filter" class="text-sm text-muted-foreground">Language:</Label>
				<select id="lang-filter" value={langFilter} onchange={changeLang} class="rounded-md border border-input bg-background px-3 py-1.5 text-sm">
					{#each languages as lang}
						<option value={lang}>{langLabel(lang)}</option>
					{/each}
				</select>
			</div>
		</div>

		{#if popularTags.length > 0}
			<div class="rounded-xl border border-border bg-background p-4">
				<h3 class="font-medium text-sm mb-3">Popular Tags</h3>
				<div class="flex flex-wrap gap-2">
					{#each popularTags as t}
						<button
							onclick={() => toggleTag(t.tag)}
							class="text-xs px-2.5 py-1 rounded-full border transition-colors {activeTag === t.tag
								? 'bg-primary text-primary-foreground border-primary'
								: 'bg-muted text-muted-foreground border-border hover:bg-primary/10 hover:text-primary'}"
						>
							{t.tag} <span class="opacity-70 ml-1">{t.count}</span>
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>

	{#if comics.length === 0}
		<div class="text-center py-20">
			<p class="text-lg text-muted-foreground">No comics found.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
		<Pagination page={data.page} {totalPages} onPage={goPage} />
	{/if}
</section>
