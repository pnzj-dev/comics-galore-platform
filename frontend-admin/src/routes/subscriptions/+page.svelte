<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';

	let subs = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		try { const res = await api.get<{ subscriptions: any[] }>('/admin/subscriptions'); subs = res.subscriptions; } catch {}
		loading = false;
	});

	function statusClass(s: string): string {
		return s === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400';
	}
</script>

<svelte:head><title>Subscriptions — Admin</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Subscriptions</h1>

	{#if loading}
		<div class="rounded-xl border border-border overflow-hidden animate-pulse">
			<div class="divide-y divide-border">
				{#each Array(6) as _}
					<div class="flex items-center gap-4 px-6 py-3">
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16 flex-1"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
					</div>
				{/each}
			</div>
		</div>
	{:else if subs.length === 0}
		<p class="text-muted-foreground text-center py-12">No subscriptions yet.</p>
	{:else}
		<div class="rounded-xl border border-border overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-muted/50 text-left">
						<th class="px-6 py-3 font-medium">User ID</th>
						<th class="px-6 py-3 font-medium">Plan</th>
						<th class="px-6 py-3 font-medium">Status</th>
						<th class="px-6 py-3 font-medium">Tier</th>
						<th class="px-6 py-3 font-medium">Expires</th>
						<th class="px-6 py-3 font-medium w-40">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each subs as s}
						<tr class="hover:bg-muted/30">
							<td class="px-6 py-3 font-mono text-xs">{s.user_id?.slice(0, 8)}...</td>
							<td class="px-6 py-3 font-mono text-xs">{s.plan_id?.slice(0, 8)}...</td>
							<td class="px-6 py-3"><span class="text-xs px-2 py-0.5 rounded-full {statusClass(s.status)}">{s.status}</span></td>
							<td class="px-6 py-3 capitalize text-xs">{s.tier}</td>
							<td class="px-6 py-3 text-xs text-muted-foreground">{s.expires_at ? new Date(s.expires_at).toLocaleDateString() : '-'}</td>
							<td class="px-6 py-3"><Button size="sm" variant="outline" disabled>Cancel</Button></td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
