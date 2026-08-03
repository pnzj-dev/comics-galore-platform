<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let subscriptions = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') {
			await goto('/');
			return;
		}
		try {
			const res = await api.get<{ subscriptions: any[] }>('/admin/subscriptions');
			subscriptions = res.subscriptions;
		} catch { /* */ }
		loading = false;
	});
</script>

<svelte:head>
	<title>Subscriptions - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Subscriptions</h1>

	{#if loading}
		<p class="text-muted-foreground">Loading...</p>
	{:else if subscriptions.length === 0}
		<p class="text-muted-foreground">No subscriptions yet.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border text-left">
						<th class="py-2 pr-4">ID</th>
						<th class="py-2 pr-4">User</th>
						<th class="py-2 pr-4">Plan</th>
						<th class="py-2 pr-4">Status</th>
						<th class="py-2 pr-4">Activated</th>
						<th class="py-2 pr-4">Expires</th>
						<th class="py-2">Created</th>
					</tr>
				</thead>
				<tbody>
					{#each subscriptions as sub}
						<tr class="border-b border-border">
							<td class="py-2 pr-4 font-mono text-xs">{sub.id?.slice(0, 8)}...</td>
							<td class="py-2 pr-4 font-mono text-xs">{sub.user_id?.slice(0, 8)}...</td>
							<td class="py-2 pr-4 font-mono text-xs">{sub.plan_id?.slice(0, 8)}...</td>
							<td class="py-2 pr-4">
								<span class="text-xs rounded-full px-2 py-0.5 {sub.status === 'active' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'}">{sub.status}</span>
							</td>
							<td class="py-2 pr-4 text-xs text-muted-foreground">{sub.activated_at ? new Date(sub.activated_at).toLocaleDateString() : '-'}</td>
							<td class="py-2 pr-4 text-xs text-muted-foreground">{sub.expires_at ? new Date(sub.expires_at).toLocaleDateString() : '-'}</td>
							<td class="py-2 text-xs text-muted-foreground">{new Date(sub.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
