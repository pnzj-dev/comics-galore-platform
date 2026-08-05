<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';

	let comics = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/'); return; }
		try {
			const res = await api.get<{ comics: any[] }>('/admin/comics');
			comics = res.comics;
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Comics - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Manage Comics</h1>

	{#if loading}
		<p class="text-muted-foreground">Loading...</p>
	{:else if comics.length === 0}
		<p class="text-muted-foreground">No comics found.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead><tr class="border-b text-left">
					<th class="py-2 pr-4">Title</th>
					<th class="py-2 pr-4">Status</th>
					<th class="py-2 pr-4">Views</th>
					<th class="py-2 pr-4">Downloads</th>
					<th class="py-2">Created</th>
				</tr></thead>
				<tbody>
					{#each comics as c}
						<tr class="border-b">
							<td class="py-2 pr-4"><a href="/comics/{c.id}" class="hover:text-primary">{c.title}</a></td>
							<td class="py-2 pr-4">
								<span class="text-xs rounded-full px-2 py-0.5 {c.status === 'published' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : c.status === 'pending_review' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'}">{c.status?.replace('_', ' ')}</span>
							</td>
							<td class="py-2 pr-4">{c.view_count}</td>
							<td class="py-2 pr-4">{c.download_count}</td>
							<td class="py-2 text-xs text-muted-foreground">{new Date(c.created_at).toLocaleDateString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
