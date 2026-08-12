<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/table/AdminTable.svelte';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();
	let confirmDelete = $state('');

	const columns = [
		{ key: 'title', label: 'Title', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Title...' },
		{ key: 'author', label: 'Author', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Author...' },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'published', label: 'Published' },
			{ value: 'pending_review', label: 'Pending' },
			{ value: 'rejected', label: 'Rejected' },
		]},
		{ key: 'view_count', label: 'Views', sortable: true },
		{ key: 'download_count', label: 'Downloads', sortable: true },
		{ key: 'created_at', label: 'Created', sortable: true },
	];

	function statusClass(s: string): string {
		return s === 'published' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : s === 'pending_review' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
	}

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
				<span class="text-xs px-2 py-0.5 rounded-full flex items-center gap-1 w-fit {statusClass(row.status as string)}">
					{#if row.status === 'published'}
						<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 21V7"/><path d="M16 21V7"/><path d="M8 21V7"/><path d="M3 13h2l1 2 3-5 1 1.5 2-3"/></svg>
					{/if}
					{(row.status as string)?.replace('_', ' ')}
				</span>
			{:else if col.key === 'view_count'}
				<span class="text-xs">{row.view_count as number}</span>
			{:else if col.key === 'download_count'}
				<span class="text-xs">{row.download_count as number}</span>
			{:else if col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{formatDate(row.created_at as string)}</span>
			{/if}
		{/snippet}
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
