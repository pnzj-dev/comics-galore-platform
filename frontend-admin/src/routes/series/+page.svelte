<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AdminTable from '$lib/components/table/AdminTable.svelte';
	import DetailRow from '$lib/components/DetailRow.svelte';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();

	const columns = [
		{ key: 'title', label: 'Title', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Title...' },
		{ key: 'genre', label: 'Genre', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Genre...' },
		{ key: 'category', label: 'Category', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Category...' },
		{ key: 'schedule_day', label: 'Schedule', sortable: true },
		{ key: 'views_count', label: 'Views', sortable: true },
		{ key: 'hearts_count', label: 'Hearts', sortable: true },
		{ key: 'created_at', label: 'Created', sortable: true },
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
</script>

<svelte:head><title>Series — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Series</h1>

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
		detailsTitle="Series details"
	>
		{#snippet children(row, col)}
			{#if col.key === 'title'}
				<a href={`http://localhost:5173/series/${row.id}`} target="_blank" class="hover:text-primary text-xs">{row.title as string}</a>
			{:else if col.key === 'genre'}
				<span class="text-xs text-muted-foreground">{row.genre as string || '—'}</span>
			{:else if col.key === 'category'}
				<span class="text-xs text-muted-foreground">{row.category as string || '—'}</span>
			{:else if col.key === 'schedule_day'}
				<span class="text-xs">{row.schedule_day as string || '—'}</span>
			{:else if col.key === 'views_count'}
				<span class="text-xs">{row.views_count as number}</span>
			{:else if col.key === 'hearts_count'}
				<span class="text-xs">{row.hearts_count as number}</span>
			{:else if col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{formatDate(row.created_at as string)}</span>
			{/if}
		{/snippet}
		{#snippet details(row)}
			<DetailRow label="ID" value={row.id as string} />
			<DetailRow label="Slug" value={row.slug as string} />
			<DetailRow label="Title" value={row.title as string} />
			<DetailRow label="Description" value={row.description as string} />
			<DetailRow label="Genre" value={row.genre as string} />
			<DetailRow label="Category" value={row.category as string} />
			<DetailRow label="Schedule" value={row.schedule_day as string} />
			<DetailRow label="Views" value={row.views_count as number} />
			<DetailRow label="Hearts" value={row.hearts_count as number} />
			<DetailRow label="Uploader" value={row.uploader_id as string} />
			<DetailRow label="Created" value={formatDate(row.created_at as string)} />
			<div class="pt-2">
				<a href={`http://localhost:5173/series/${row.id}`} target="_blank" class="text-xs text-primary hover:underline">View on site ↗</a>
			</div>
		{/snippet}
	</AdminTable>
</section>
