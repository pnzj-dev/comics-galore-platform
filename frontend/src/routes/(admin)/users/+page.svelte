<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let users = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/'); return; }
		try {
			const res = await api.get<{ users: any[] }>('/admin/users');
			users = res.users;
		} catch {}
		loading = false;
	});

	async function changeRole(userId: string, newRole: string) {
		await api.post(`/admin/users/${userId}/role`, { role: newRole });
		users = users.map(u => u.id === userId ? { ...u, role: newRole } : u);
	}
</script>

<svelte:head><title>Users - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Users</h1>

	{#if loading}
		<p class="text-muted-foreground">Loading...</p>
	{:else if users.length === 0}
		<p class="text-muted-foreground">No users found.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead><tr class="border-b text-left">
					<th class="py-2 pr-4">Email</th>
					<th class="py-2 pr-4">Role</th>
					<th class="py-2 pr-4">Tier</th>
					<th class="py-2">Created</th>
				</tr></thead>
				<tbody>
					{#each users as u}
						<tr class="border-b">
							<td class="py-2 pr-4">{u.email}</td>
							<td class="py-2 pr-4">
								<select value={u.role} onchange={(e) => changeRole(u.id, e.target.value)} class="rounded border border-input bg-background px-2 py-0.5 text-xs">
									<option value="user">user</option>
									<option value="uploader">uploader</option>
									<option value="moderator">moderator</option>
									<option value="admin">admin</option>
								</select>
							</td>
							<td class="py-2 pr-4 capitalize">{u.tier}</td>
							<td class="py-2 text-xs text-muted-foreground">{new Date(u.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
