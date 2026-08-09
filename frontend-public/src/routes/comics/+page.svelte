<script lang="ts">
	import { encore } from '$lib/api/encore';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { Label } from '$lib/components/ui/label/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let comics = $state(data.comics);
	let loading = $state(false);
	let langFilter = $state('');
	let hideMature = $state(false);

	const languages = ['', 'en', 'ja', 'es', 'ko', 'fr', 'pt-BR', 'zh-CN', 'de', 'it', 'id'];

	async function loadComics() {
		loading = true;
		try {
			const q: any = {};
			if (langFilter) q.language = langFilter;
			if (hideMature) q.exclude_mature = 'true';
			const res = await encore.comics.ListComics(q);
			comics = res.comics;
		} catch {}
		loading = false;
	}

	function changeLang(e: Event) {
		langFilter = (e.target as HTMLSelectElement).value;
		loadComics();
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
				<option value="">All languages</option>
				{#each languages.filter(l => l !== '') as lang}
					<option value={lang}>{lang}</option>
				{/each}
			</select>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each Array(8) as _}
				<div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden animate-pulse">
					<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
					<div class="p-3 space-y-2">
						<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
					</div>
				</div>
			{/each}
		</div>
	{:else if comics.length === 0}
		<div class="text-center py-20">
			<p class="text-lg text-muted-foreground">No comics found.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
			{#each comics as comic}
				<ComicCard {...comic} />
			{/each}
		</div>
	{/if}
</section>
