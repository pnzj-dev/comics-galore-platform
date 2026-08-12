<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/table/AdminTable.svelte';
	import DetailRow from '$lib/components/DetailRow.svelte';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();
	let actionLoading = $state('');

	const columns = [
		{ key: 'title', label: 'Title', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Title...' },
		{ key: 'author', label: 'Author', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Author...' },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'published', label: 'Published' },
			{ value: 'pending_review', label: 'Pending' },
			{ value: 'rejected', label: 'Rejected' },
		]},
	];

	function buildUrl(updates: Record<string, string | undefined>) {
		const url = new URL(page.url);
		url.searchParams.set('page', '1');
		for (const [k, v] of Object.entries(updates)) {
			if (v) url.searchParams.set(k, v);
			else url.searchParams.delete(k);
		}
		return url.pathname + url.search;
	}
	async function onSort(key: string, dir: 'asc' | 'desc') { await goto(buildUrl({ sort: key, sort_dir: dir }), { keepFocus: true }); }
	async function onSearch(value: string) { await goto(buildUrl({ search: value || undefined }), { keepFocus: true }); }
	async function onFilter(key: string, value: string) { await goto(buildUrl({ [`filter_${key}`]: value || undefined }), { keepFocus: true }); }
	async function onPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		await goto(url.pathname + url.search, { keepFocus: true });
	}
	async function toggleFilters() {
		const url = new URL(page.url);
		if (data.showFilters) url.searchParams.delete('show_filters');
		else url.searchParams.set('show_filters', '1');
		await goto(url.pathname + url.search, { keepFocus: true });
	}
	async function restoreComic(id: string) {
		actionLoading = id;
		await encore.comics.RestoreComic(id);
		actionLoading = '';
		await goto(page.url.pathname + page.url.search);
	}
	async function permanentDelete(id: string) {
		actionLoading = id;
		await encore.comics.DeleteComic(id);
		actionLoading = '';
		await goto(page.url.pathname + page.url.search);
	}
</script>

<svelte:head><title>Recycle Bin — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Recycle Bin</h1>

	<AdminTable
		{columns}
		data={data.results}
		total={data.total}
		page={data.page}
		limit={data.limit}
		sortKey={data.sort}
		sortDir={data.sortDir as 'asc' | 'desc'}
		search={data.search}
		filters={data.filters as Record<string, string>}
		showFilters={data.showFilters}
		onToggleFilters={toggleFilters}
		emptyMessage="Recycle bin is empty."
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
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
		{#snippet details(row)}
			<DetailRow label="ID" value={row.id as string} />
			<DetailRow label="Title" value={row.title as string} />
			<DetailRow label="Author" value={row.author as string} />
			<DetailRow label="Status" value={(row.status as string)?.replace('_', ' ')} />
			<DetailRow label="Deleted at" value={formatDate(row.deleted_at as string)} />
		{/snippet}
	</AdminTable>
</section>
