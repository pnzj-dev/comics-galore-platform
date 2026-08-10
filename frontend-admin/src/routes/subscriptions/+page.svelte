<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AdminTable from '$lib/components/AdminTable.svelte';

	let { data } = $props();

	const columns = [
		{ key: 'user_id', label: 'User ID', sortable: false, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'plan_id', label: 'Plan ID', sortable: false },
		{ key: 'status', label: 'Status', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'active', label: 'Active' },
			{ value: 'inactive', label: 'Inactive' },
			{ value: 'expired', label: 'Expired' },
		]},
		{ key: 'tier', label: 'Tier', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'free', label: 'Free' },
			{ value: 'bronze', label: 'Bronze' },
			{ value: 'silver', label: 'Silver' },
			{ value: 'gold', label: 'Gold' },
			{ value: 'platinum', label: 'Platinum' },
		]},
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
		await goto(url.pathname + url.search);
	}
</script>

<svelte:head><title>Subscriptions — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Subscriptions</h1>

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
		loading={false}
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
	>
		{#snippet children(row, col)}
			{#if col.key === 'user_id'}
				<span class="font-mono text-xs">{String(row.user_id).slice(0, 8)}...</span>
			{:else if col.key === 'plan_id'}
				<span class="font-mono text-xs">{String(row.plan_id).slice(0, 8)}...</span>
			{:else if col.key === 'status'}
				<span class="text-xs px-2 py-0.5 rounded-full {row.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">{row.status as string}</span>
			{:else if col.key === 'tier'}
				<span class="text-xs capitalize">{row.tier as string}</span>
			{:else if col.key === 'expires_at' || col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{row.expires_at ? new Date(row.expires_at as string).toLocaleDateString() : row.created_at ? new Date(row.created_at as string).toLocaleDateString() : '-'}</span>
			{/if}
		{/snippet}
		{#snippet actions(row)}
			<button class="text-xs text-muted-foreground" disabled>Cancel</button>
		{/snippet}
	</AdminTable>
</section>
