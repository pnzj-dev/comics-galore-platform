<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	let { data } = $props();

	function statusClass(s: string): string {
		return s === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400';
	}
</script>

<svelte:head><title>Subscriptions — Admin</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Subscriptions</h1>

	{#if data.subs.length === 0}
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
					{#each data.subs as s}
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
