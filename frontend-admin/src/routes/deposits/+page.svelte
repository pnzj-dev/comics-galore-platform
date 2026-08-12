<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/table/AdminTable.svelte';
	import DepositDetailsModal from '$lib/components/DepositDetailsModal.svelte';
	import { formatDate, formatUSD } from '$lib/utils/format';

	let { data } = $props();
	let selectedDeposit = $state<Record<string, unknown> | null>(null);

	const columns = [
		{ key: 'user_id', label: 'User ID', filterable: true, filterType: 'text' as const, filterPlaceholder: 'User ID...' },
		{ key: 'currency_crypto', label: 'Currency', sortable: true },
		{ key: 'amount_crypto', label: 'Amount (crypto)', sortable: true },
		{ key: 'amount_usd_cents', label: 'USD', sortable: true },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'pending', label: 'Pending' },
			{ value: 'completed', label: 'Completed' },
			{ value: 'expired', label: 'Expired' },
			{ value: 'failed', label: 'Failed' },
		]},
		{ key: 'created_at', label: 'Created', sortable: true },
		{ key: 'completed_at', label: 'Completed', sortable: true },
	];

	function statusClass(s: string): string {
		return s === 'completed' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
			: s === 'failed' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
			: s === 'expired' ? 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
			: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400';
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
</script>

<svelte:head><title>Deposits — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Deposits</h1>

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
			{#if col.key === 'user_id'}
				<span class="font-mono text-xs truncate block" title={String(row.user_id)}>{String(row.user_id)}</span>
			{:else if col.key === 'currency_crypto'}
				<span class="text-xs uppercase">{row.currency_crypto as string}</span>
			{:else if col.key === 'amount_crypto'}
				<span class="font-mono text-xs">{row.amount_crypto as string || '—'}</span>
			{:else if col.key === 'amount_usd_cents'}
				<span class="text-xs font-medium">{formatUSD(row.amount_usd_cents as number)}</span>
			{:else if col.key === 'status'}
				<span class="text-xs px-2 py-0.5 rounded-full w-fit {statusClass(row.status as string)}">{(row.status as string)?.replace('_', ' ')}</span>
			{:else if col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{formatDate(row.created_at as string)}</span>
			{:else if col.key === 'completed_at'}
				<span class="text-xs text-muted-foreground">{formatDate(row.completed_at as string)}</span>
			{/if}
		{/snippet}
		{#snippet actions(row)}
			<Button size="sm" variant="outline" onclick={() => selectedDeposit = row}>Details</Button>
		{/snippet}
	</AdminTable>

	<DepositDetailsModal open={!!selectedDeposit} deposit={selectedDeposit} onClose={() => selectedDeposit = null} />
</section>
