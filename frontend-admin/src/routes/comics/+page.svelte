<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/AdminTable.svelte';
	import { createColumnHelper, renderComponent } from '@tanstack/svelte-table';
	import SortHeader from '$lib/components/SortHeader.svelte';

	let { data } = $props();
	let confirmDelete = $state('');

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const columns = columnHelper.columns([
		columnHelper.accessor('title', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Title', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: { filterType: 'text', filterPlaceholder: 'Filter...' },
			cell: ({ row }) => row.getValue(),
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
				const cls = s === 'published' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : s === 'pending_review' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
				return `<span class="text-xs px-2 py-0.5 rounded-full inline-flex items-center gap-1 ${cls}">${s === 'published' ? '<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-book-open-check"><path d="M12 21V7"/><path d="M16 21V7"/><path d="M8 21V7"/><path d="M3 13h2l1 2 3-5 1 1.5 2-3"/></svg>' : ''}${(s as string)?.replace('_', ' ')}</span>`;
			},
		}),
		columnHelper.accessor('view_count', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Views', header }),
			enableSorting: true,
			cell: ({ getValue }) => getValue(),
		}),
		columnHelper.accessor('download_count', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Downloads', header }),
			enableSorting: true,
			cell: ({ getValue }) => getValue(),
		}),
		columnHelper.accessor('created_at', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Created', header }),
			enableSorting: true,
			cell: ({ getValue }) => {
				const d = getValue() as string;
				return new Date(d).toLocaleDateString();
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
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		await goto(url.pathname + url.search, { keepFocus: true });
	}
	async function deleteComic(id: string) {
		await encore.comics.DeleteComic(id);
		confirmDelete = '';
		await goto(page.url.pathname + page.url.search);
	}
</script>

<svelte:head><title>Comics — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Comics</h1>

	<AdminTable
		{columns}
		data={data.results as Record<string, unknown>[]}
		total={data.total}
		page={data.page}
		limit={data.limit}
		sortKey={data.sort}
		sortDir={data.sortDir as 'asc' | 'desc'}
		search={data.search}
		filters={data.filters as Record<string, string>}
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
	>
		{#snippet actions(row)}
			{#if confirmDelete === row.id}
				<Button size="sm" variant="destructive" onclick={() => deleteComic(row.id as string)}>Confirm</Button>
				<Button size="sm" variant="ghost" onclick={() => confirmDelete = ''}>Cancel</Button>
			{:else}
				<Button size="sm" variant="ghost" class="text-destructive hover:bg-destructive/10" onclick={() => confirmDelete = row.id as string}>Delete</Button>
			{/if}
		{/snippet}
	</AdminTable>
</section>
