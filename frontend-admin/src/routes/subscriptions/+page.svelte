<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AdminTable from '$lib/components/table/AdminTable.svelte';
	import TierBadge from '$lib/components/TierBadge.svelte';
	import DetailRow from '$lib/components/DetailRow.svelte';
	import { PAID_TIER_OPTIONS } from '$lib/constants/tiers';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();

	const columns = [
		{ key: 'user_id', label: 'User ID', filterable: true, filterType: 'text' as const, filterPlaceholder: 'User ID...' },
		{ key: 'plan_id', label: 'Plan ID' },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'active', label: 'Active' },
			{ value: 'inactive', label: 'Inactive' },
			{ value: 'expired', label: 'Expired' },
		]},
		{ key: 'tier', label: 'Tier', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: PAID_TIER_OPTIONS },
		{ key: 'expires_at', label: 'Expires', sortable: true },
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

<svelte:head><title>Subscriptions — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Subscriptions</h1>

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
			{#if col.key === 'user_id' || col.key === 'plan_id'}
				<span class="font-mono text-xs truncate block" title={String(row[col.key])}>{String(row[col.key])}</span>
			{:else if col.key === 'status'}
				<span class="text-xs px-2 py-0.5 rounded-full {row.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">{row.status as string}</span>
			{:else if col.key === 'tier'}
				<TierBadge tier={row.tier as string} />
			{:else if col.key === 'expires_at' || col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{formatDate(row[col.key] as string)}</span>
			{/if}
		{/snippet}
		{#snippet details(row)}
			<DetailRow label="ID" value={row.id as string} />
			<DetailRow label="User ID" value={row.user_id as string} />
			<DetailRow label="Plan ID" value={row.plan_id as string} />
			<DetailRow label="Status" value={row.status as string} />
			<DetailRow label="Tier" value={row.tier as string} />
			<DetailRow label="Active" value={row.active ? 'Yes' : 'No'} />
			<DetailRow label="Activated" value={formatDate(row.activated_at as string)} />
			<DetailRow label="Expires" value={formatDate(row.expires_at as string)} />
			<DetailRow label="Created" value={formatDate(row.created_at as string)} />
		{/snippet}
	</AdminTable>
</section>
