<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import AdminTable from '$lib/components/AdminTable.svelte';

	let { data } = $props();

	let results = $state(data.results);

	const columns = [
		{ key: 'email', label: 'Email', sortable: true, filterable: true, filterType: 'text' as const, filterPlaceholder: 'Filter...' },
		{ key: 'role', label: 'Role', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'user', label: 'User' },
			{ value: 'uploader', label: 'Uploader' },
			{ value: 'moderator', label: 'Moderator' },
			{ value: 'admin', label: 'Admin' },
		]},
		{ key: 'tier', label: 'Tier', sortable: true, filterable: true, filterType: 'select' as const, filterOptions: [
			{ value: 'free', label: 'Free' },
			{ value: 'bronze', label: 'Bronze' },
			{ value: 'silver', label: 'Silver' },
			{ value: 'gold', label: 'Gold' },
			{ value: 'platinum', label: 'Platinum' },
		]},
		{ key: 'status', label: 'Status', sortable: false },
		{ key: 'created_at', label: 'Created', sortable: true },
	];

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

	async function changeRole(userId: string, newRole: string) {
		await encore.auth.AdminUpdateUserRole(userId, { role: newRole });
		results = results.map(u => u.id === userId ? { ...u, role: newRole } : u);
	}

	async function banUser(userId: string) {
		await encore.auth.AdminBanUser(userId, { reason: '' });
		results = results.map(u => u.id === userId ? { ...u, banned_at: new Date().toISOString() } : u);
	}

	async function unbanUser(userId: string) {
		await encore.auth.AdminUnbanUser(userId);
		results = results.map(u => u.id === userId ? { ...u, banned_at: null } : u);
	}

	async function suspendUser(userId: string) {
		await encore.auth.AdminSuspendUser(userId, { reason: '' });
		results = results.map(u => u.id === userId ? { ...u, suspended_at: new Date().toISOString() } : u);
	}

	async function unsuspendUser(userId: string) {
		await encore.auth.AdminUnsuspendUser(userId);
		results = results.map(u => u.id === userId ? { ...u, suspended_at: null } : u);
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
			{#if col.key === 'email'}
				<span class="text-xs">{row.email as string}</span>
			{:else if col.key === 'role'}
				<select value={row.role as string} onchange={(e) => changeRole(row.id as string, (e.target as HTMLSelectElement).value)} class="rounded border border-input bg-background px-2 py-1 text-xs">
					<option value="user">user</option><option value="uploader">uploader</option><option value="moderator">moderator</option><option value="admin">admin</option>
				</select>
			{:else if col.key === 'tier'}
				<span class="text-xs px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 capitalize">{row.tier as string}</span>
			{:else if col.key === 'status'}
				{@const st = statusBadge(row)}
				{#if st === 'banned'}<span class="px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">Banned</span>
				{:else if st === 'suspended'}<span class="px-2 py-0.5 rounded-full text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400">Suspended</span>
				{:else}<span class="px-2 py-0.5 rounded-full text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Active</span>
				{/if}
			{:else if col.key === 'created_at'}
				<span class="text-xs text-muted-foreground">{new Date(row.created_at as string).toLocaleDateString()}</span>
			{/if}
		{/snippet}
		{#snippet actions(row)}
			{@const st = statusBadge(row)}
			{#if st === 'banned'}<Button size="sm" variant="outline" onclick={() => unbanUser(row.id as string)}>Unban</Button>
			{:else if st === 'suspended'}<Button size="sm" variant="outline" onclick={() => unsuspendUser(row.id as string)}>Unsuspend</Button>
			{:else}
				<Button size="sm" variant="outline" onclick={() => suspendUser(row.id as string)}>Suspend</Button>
				<Button size="sm" variant="destructive" onclick={() => banUser(row.id as string)}>Ban</Button>
			{/if}
		{/snippet}
	</AdminTable>
</section>
