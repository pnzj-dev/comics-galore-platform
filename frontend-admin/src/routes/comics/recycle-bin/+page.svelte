<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/AdminTable.svelte';

	let { data } = $props();

	let results = $state(data.results);
	let actionLoading = $state('');

	const columns = [
		{ key: 'title', label: 'Title', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'author', label: 'Author', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'published', label: 'Published' },
			{ value: 'pending_review', label: 'Pending' },
			{ value: 'rejected', label: 'Rejected' },
		]},
	];

	function buildUrl(updates: Record<string, string | undefined>) {
		const url = new URL($page.url);
		url.searchParams.set('page', '1');
		for (const [k, v] of Object.entries(updates)) {
			if (v) url.searchParams.set(k, v);
			else url.searchParams.delete(k);
		}
		return url.pathname + url.search;
	}

	async function onSort(key: string, dir: 'asc' | 'desc') {
		await goto(buildUrl({ sort: key, sort_dir: dir }));
	}

	async function onSearch(value: string) {
		await goto(buildUrl({ search: value || undefined }));
	}

	async function onFilter(key: string, value: string) {
		await goto(buildUrl({ [`filter_${key}`]: value || undefined }));
	}

	async function onPage(p: number) {
		const url = new URL($page.url);
		url.searchParams.set('page', String(p));
		await goto(url.pathname + url.search);
	}

	async function restoreComic(id: string) {
		actionLoading = id;
		await encore.comics.RestoreComic(id);
		results = results.filter(c => c.id !== id);
		actionLoading = '';
	}

	async function permanentDelete(id: string) {
		actionLoading = id;
		await encore.comics.DeleteComic(id);
		results = results.filter(c => c.id !== id);
		actionLoading = '';
	}
</script>

<svelte:head><title>Recycle Bin — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Recycle Bin</h1>

	<AdminTable
		{columns}
		data={results as Record<string, unknown>[]}
		total={data.total}
		page={data.page}
		limit={data.limit}
		sortKey={data.sort}
		sortDir={data.sortDir as 'asc' | 'desc'}
		search={data.search}
		filters={data.filters as Record<string, string>}
		loading={false}
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
		emptyMessage="Recycle bin is empty."
	>
		{#snippet children(row, col)}
			{#if col.key === 'title'}
				<a href={`http://localhost:5173/comics/${row.slug}`} target="_blank" class="hover:text-primary text-xs">{row.title as string}</a>
			{:else if col.key === 'author'}
				<span class="text-xs text-muted-foreground">{row.author as string || '—'}</span>
			{:else if col.key === 'status'}
				<span class="text-xs px-2 py-0.5 rounded-full w-fit {(row.status as string) === 'published' ? 'bg-green-100 text-green-700' : (row.status as string) === 'pending_review' ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700'}">{(row.status as string)?.replace('_', ' ')}</span>
			{/if}
		{/snippet}
		{#snippet actions(row)}
			<Button size="sm" variant="outline" onclick={() => restoreComic(row.id as string)} disabled={actionLoading === row.id}>Restore</Button>
			<Button size="sm" variant="destructive" onclick={() => permanentDelete(row.id as string)} disabled={actionLoading === row.id}>Delete</Button>
		{/snippet}
	</AdminTable>
</section>
