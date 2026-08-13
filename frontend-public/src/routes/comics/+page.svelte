<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { Label } from '$lib/components/ui/label/index.js';

	let { data } = $props();

	const comics = $derived(data.comics);
	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));
	const langFilter = $derived(data.lang || '');

	const facets = $derived(data.facets || []);
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

	function changeLang(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		const url = new URL(page.url);
		if (value) url.searchParams.set('language', value);
		else url.searchParams.delete('language');
		url.searchParams.delete('page');
		goto(url.pathname + url.search, { keepFocus: true });
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
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold">Browse Comics</h1>
		<div class="flex items-center gap-2">
			<Label for="lang-filter" class="text-sm text-muted-foreground">Language:</Label>
			<select id="lang-filter" value={langFilter} onchange={changeLang} class="rounded-md border border-input bg-background px-3 py-1.5 text-sm">
				{#each languages as lang}
					<option value={lang}>{langLabel(lang)}</option>
				{/each}
			</select>
		</div>
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
