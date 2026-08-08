<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';

	let users = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		try { const res = await api.get<{ users: any[] }>('/admin/users'); users = res.users; } catch {}
		loading = false;
	});

	async function changeRole(userId: string, newRole: string) {
		await api.post(`/admin/users/${userId}/role`, { role: newRole });
		users = users.map(u => u.id === userId ? { ...u, role: newRole } : u);
	}

	async function banUser(userId: string) {
		await api.post(`/admin/users/${userId}/ban`, { reason: '' });
		users = users.map(u => u.id === userId ? { ...u, banned_at: new Date().toISOString() } : u);
	}

	async function unbanUser(userId: string) {
		await api.post(`/admin/users/${userId}/unban`);
		users = users.map(u => u.id === userId ? { ...u, banned_at: null } : u);
	}

	async function suspendUser(userId: string) {
		await api.post(`/admin/users/${userId}/suspend`, { reason: '' });
		users = users.map(u => u.id === userId ? { ...u, suspended_at: new Date().toISOString() } : u);
	}

	async function unsuspendUser(userId: string) {
		await api.post(`/admin/users/${userId}/unsuspend`);
		users = users.map(u => u.id === userId ? { ...u, suspended_at: null } : u);
	}

	function statusBadge(u: any): string {
		if (u.banned_at) return 'banned';
		if (u.suspended_at) return 'suspended';
		return 'active';
	}
</script>

<svelte:head><title>Users — Admin</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Users</h1>

	{#if loading}
		<div class="rounded-xl border border-border overflow-hidden animate-pulse">
			<div class="divide-y divide-border">
				{#each Array(8) as _}
					<div class="flex items-center gap-4 px-6 py-3">
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 flex-1"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24"></div>
					</div>
				{/each}
			</div>
		</div>
	{:else if users.length === 0}
		<p class="text-muted-foreground text-center py-12">No users found.</p>
	{:else}
		<div class="rounded-xl border border-border overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-muted/50 text-left">
						<th class="px-6 py-3 font-medium">Email</th>
						<th class="px-6 py-3 font-medium">Role</th>
						<th class="px-6 py-3 font-medium">Tier</th>
						<th class="px-6 py-3 font-medium">Status</th>
						<th class="px-6 py-3 font-medium">Created</th>
						<th class="px-6 py-3 font-medium w-60">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each users as u}
						{@const st = statusBadge(u)}
						<tr class="hover:bg-muted/30">
							<td class="px-6 py-3">{u.email}</td>
							<td class="px-6 py-3">
								<select value={u.role} onchange={(e) => changeRole(u.id, e.target.value)} class="rounded border border-input bg-background px-2 py-1 text-xs">
									<option value="user">user</option>
									<option value="uploader">uploader</option>
									<option value="moderator">moderator</option>
									<option value="admin">admin</option>
								</select>
							</td>
							<td class="px-6 py-3 capitalize text-xs"><span class="px-2 py-0.5 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">{u.tier}</span></td>
							<td class="px-6 py-3">
								{#if st === 'banned'}
									<span class="px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">Banned</span>
								{:else if st === 'suspended'}
									<span class="px-2 py-0.5 rounded-full text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400">Suspended</span>
								{:else}
									<span class="px-2 py-0.5 rounded-full text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Active</span>
								{/if}
							</td>
							<td class="px-6 py-3 text-xs text-muted-foreground">{new Date(u.created_at).toLocaleDateString()}</td>
							<td class="px-6 py-3">
								<div class="flex items-center gap-1">
									{#if st === 'banned'}
										<Button size="sm" variant="outline" onclick={() => unbanUser(u.id)}>Unban</Button>
									{:else if st === 'suspended'}
										<Button size="sm" variant="outline" onclick={() => unsuspendUser(u.id)}>Unsuspend</Button>
									{:else}
										<Button size="sm" variant="outline" onclick={() => suspendUser(u.id)}>Suspend</Button>
										<Button size="sm" variant="destructive" onclick={() => banUser(u.id)}>Ban</Button>
									{/if}
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
