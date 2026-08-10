<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BookOpenCheck } from 'lucide-svelte';
	import AdminTable from '$lib/components/AdminTable.svelte';

	let { data } = $props();

	let results = $state(data.results);
	let loading = $state(false);
	let confirmDelete = $state('');

	const columns = [
		{ key: 'title', label: 'Title', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'author', label: 'Author', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'published', label: 'Published' },
			{ value: 'pending_review', label: 'Pending' },
			{ value: 'rejected', label: 'Rejected' },
		]},
		{ key: 'view_count', label: 'Views', sortable: true },
		{ key: 'download_count', label: 'Downloads', sortable: true },
		{ key: 'created_at', label: 'Created', sortable: true },
	];

	const current = $derived($page.url);

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

	async function deleteComic(id: string) {
		await encore.comics.DeleteComic(id);
		results = results.filter(c => c.id !== id);
		confirmDelete = '';
	}

	function statusClass(s: string): string {
		return s === 'published' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : s === 'pending_review' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
	}
</script>

<svelte:head><title>Comics — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Comics</h1>

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
	>
		{#snippet children(row, col)}
			{#if col.key === 'title'}
				<a href={`http://localhost:5173/comics/${row.slug}`} target="_blank" class="hover:text-primary text-xs">{row.title as string}</a>
			{:else if col.key === 'author'}
				<span class="text-xs text-muted-foreground">{row.author as string || '—'}</span>
			{:else if col.key === 'status'}
				<span class="text-xs px-2 py-0.5 rounded-full flex items-center gap-1 w-fit {statusClass(row.status as string)}">
					{#if row.status === 'published'}<BookOpenCheck class="size-3" />{/if}
					{(row.status as string)?.replace('_', ' ')}
				</span>
			{:else if col.key === 'view_count'}
				<span class="text-xs">{row.view_count as number}</span>
			{:else if col.key === 'download_count'}
				<span class="text-xs">{row.download_count as number}</span>
			{:else if col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{new Date(row.created_at as string).toLocaleDateString()}</span>
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
