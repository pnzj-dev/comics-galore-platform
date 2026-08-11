<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/AdminTable.svelte';
	import { createColumnHelper, renderComponent } from '@tanstack/svelte-table';
	import SortHeader from '$lib/components/SortHeader.svelte';

	let { data } = $props();
	let actionLoading = $state('');

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const columns = columnHelper.columns([
		columnHelper.accessor('title', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Title', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: { filterType: 'text', filterPlaceholder: 'Filter...' },
			cell: ({ getValue }) => getValue(),
		}),
		columnHelper.accessor('author', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Author', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: { filterType: 'text', filterPlaceholder: 'Filter...' },
			cell: ({ getValue }) => getValue() || '—',
		}),
		columnHelper.accessor('status', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Status', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: {
				filterType: 'select',
				filterOptions: [
					{ value: 'published', label: 'Published' },
					{ value: 'pending_review', label: 'Pending' },
					{ value: 'rejected', label: 'Rejected' },
				],
			},
			cell: ({ getValue }) => {
				const s = getValue() as string;
				return s?.replace('_', ' ') || '';
			},
		}),
	]);

	function buildUrl(updates: Record<string, string | undefined>) {
		const url = new URL(page.url);
		url.searchParams.set('page', '1');
		for (const [k, v] of Object.entries(updates)) {
			if (v) url.searchParams.set(k, v);
			else url.searchParams.delete(k);
		}
		return url.pathname + url.search;
	}
	async function onSort(key: string, dir: 'asc' | 'desc') { await goto(buildUrl({ sort: key, sort_dir: dir })); }
	async function onSearch(value: string) { await goto(buildUrl({ search: value || undefined })); }
	async function onFilter(key: string, value: string) { await goto(buildUrl({ [`filter_${key}`]: value || undefined })); }
	async function onPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
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
		emptyMessage="Recycle bin is empty."
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
	>
		{#snippet actions(row)}
			<Button size="sm" variant="outline" onclick={() => restoreComic(row.id as string)} disabled={actionLoading === row.id}>Restore</Button>
			<Button size="sm" variant="destructive" onclick={() => permanentDelete(row.id as string)} disabled={actionLoading === row.id}>Delete</Button>
		{/snippet}
	</AdminTable>
</section>
