<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/AdminTable.svelte';
	import { createColumnHelper, renderComponent } from '@tanstack/svelte-table';
	import SortHeader from '$lib/components/SortHeader.svelte';

	let { data } = $props();

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const columns = columnHelper.columns([
		columnHelper.accessor('email', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Email', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: { filterType: 'text', filterPlaceholder: 'Filter...' },
			cell: ({ getValue }) => getValue(),
		}),
		columnHelper.accessor('role', {
			header: ({ header }) => renderComponent(SortHeader, { label: 'Role', header }),
			enableSorting: true,
			enableColumnFilter: true,
			meta: {
				filterType: 'select',
				filterOptions: [
					{ value: 'user', label: 'User' },
					{ value: 'uploader', label: 'Uploader' },
					{ value: 'moderator', label: 'Moderator' },
					{ value: 'admin', label: 'Admin' },
				],
			},
			cell: ({ getValue }) => getValue(),
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
		columnHelper.accessor('status', {
			header: 'Status',
			cell: ({ row }) => {
				const u = row.original;
				if (u.banned_at) return '<span class="text-xs px-2 py-0.5 rounded-full bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">Banned</span>';
				if (u.suspended_at) return '<span class="text-xs px-2 py-0.5 rounded-full bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400">Suspended</span>';
				return '<span class="text-xs px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Active</span>';
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
	async function changeRole(userId: string, newRole: string) {
		await encore.auth.AdminUpdateUserRole(userId, { role: newRole });
		await goto(page.url.pathname + page.url.search);
	}
	async function banUser(userId: string) {
		await encore.auth.AdminBanUser(userId, { reason: '' });
		await goto(page.url.pathname + page.url.search);
	}
	async function unbanUser(userId: string) {
		await encore.auth.AdminUnbanUser(userId);
		await goto(page.url.pathname + page.url.search);
	}
	async function suspendUser(userId: string) {
		await encore.auth.AdminSuspendUser(userId, { reason: '' });
		await goto(page.url.pathname + page.url.search);
	}
	async function unsuspendUser(userId: string) {
		await encore.auth.AdminUnsuspendUser(userId);
		await goto(page.url.pathname + page.url.search);
	}
	function statusBadge(u: Record<string, unknown>): string {
		if (u.banned_at) return 'banned';
		if (u.suspended_at) return 'suspended';
		return 'active';
	}
</script>

<svelte:head><title>Users — Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Users</h1>

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
		{onSort}
		{onSearch}
		{onFilter}
		{onPage}
	>
		{#snippet actions(row)}
			{@const st = statusBadge(row)}
			{#if st === 'banned'}
				<Button size="sm" variant="outline" onclick={() => unbanUser(row.id as string)}>Unban</Button>
			{:else if st === 'suspended'}
				<Button size="sm" variant="outline" onclick={() => unsuspendUser(row.id as string)}>Unsuspend</Button>
			{:else}
				<Button size="sm" variant="outline" onclick={() => suspendUser(row.id as string)}>Suspend</Button>
				<Button size="sm" variant="destructive" onclick={() => banUser(row.id as string)}>Ban</Button>
			{/if}
		{/snippet}
	</AdminTable>
</section>
