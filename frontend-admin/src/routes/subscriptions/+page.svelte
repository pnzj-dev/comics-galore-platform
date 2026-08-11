<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import AdminTable from '$lib/components/AdminTable.svelte';
	import { createColumnHelper, renderComponent } from '@tanstack/svelte-table';
	import SortHeader from '$lib/components/SortHeader.svelte';

	let { data } = $props();

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const columns = columnHelper.columns([
		columnHelper.accessor('user_id', {
			header: 'User ID',
			enableColumnFilter: true,
			meta: { filterType: 'text', filterPlaceholder: 'Filter...' },
			cell: ({ getValue }) => String(getValue() || '').slice(0, 8) + '...',
		}),
		columnHelper.accessor('plan_id', {
			header: 'Plan ID',
			cell: ({ getValue }) => String(getValue() || '').slice(0, 8) + '...',
		}),
		columnHelper.accessor('status', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Status', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: {
				filterType: 'select',
				filterOptions: [
					{ value: 'active', label: 'Active' },
					{ value: 'inactive', label: 'Inactive' },
					{ value: 'expired', label: 'Expired' },
				],
			},
			cell: ({ getValue }) => {
				const s = getValue() as string;
				const cls = s === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400';
				return `<span class="text-xs px-2 py-0.5 rounded-full ${cls}">${s}</span>`;
			},
		}),
		columnHelper.accessor('tier', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Tier', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: {
				filterType: 'select',
				filterOptions: [
					{ value: 'free', label: 'Free' },
					{ value: 'bronze', label: 'Bronze' },
					{ value: 'silver', label: 'Silver' },
					{ value: 'gold', label: 'Gold' },
					{ value: 'platinum', label: 'Platinum' },
				],
			},
			cell: ({ getValue }) => getValue(),
		}),
		columnHelper.accessor('expires_at', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Expires', header }),
			enableSorting: true,
			cell: ({ getValue }) => {
				const d = getValue() as string;
				return d ? new Date(d).toLocaleDateString() : '—';
			},
		}),
		columnHelper.accessor('created_at', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Created', header }),
			enableSorting: true,
			cell: ({ getValue }) => new Date(getValue() as string).toLocaleDateString(),
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
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
	/>
</section>
