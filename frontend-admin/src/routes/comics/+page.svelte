<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BookOpenCheck } from 'lucide-svelte';

	let comics = $state<any[]>([]);
	let loading = $state(true);
	let confirmDelete = $state('');

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		await loadComics();
	});

	async function loadComics() {
		try { const res = await api.get<{ comics: any[] }>('/admin/comics'); comics = res.comics; } catch {}
		loading = false;
	}

	async function deleteComic(id: string) {
		await api.delete(`/comics/${id}`);
		comics = comics.filter(c => c.id !== id);
		confirmDelete = '';
	}

	function statusClass(s: string): string {
		return s === 'published' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : s === 'pending_review' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
	}
</script>

<svelte:head><title>Comics — Admin</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Comics</h1>

	{#if loading}
		<div class="rounded-xl border border-border overflow-hidden animate-pulse">
			<div class="divide-y divide-border">
				{#each Array(8) as _}
					<div class="flex items-center gap-4 px-6 py-3">
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 flex-1"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-16"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-12"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-12"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
					</div>
				{/each}
			</div>
		</div>
	{:else if comics.length === 0}
		<p class="text-muted-foreground text-center py-12">No comics found.</p>
	{:else}
		<div class="rounded-xl border border-border overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-muted/50 text-left">
						<th class="px-6 py-3 font-medium">Title</th>
						<th class="px-6 py-3 font-medium">Status</th>
						<th class="px-6 py-3 font-medium">Views</th>
						<th class="px-6 py-3 font-medium">Downloads</th>
						<th class="px-6 py-3 font-medium">Created</th>
						<th class="px-6 py-3 font-medium w-40">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each comics as c}
						<tr class="hover:bg-muted/30">
							<td class="px-6 py-3"><a href={`http://localhost:5173/comics/${c.slug}`} target="_blank" class="hover:text-primary">{c.title}</a></td>
							<td class="px-6 py-3"><span class="text-xs px-2 py-0.5 rounded-full flex items-center gap-1 w-fit {statusClass(c.status)}">{#if c.status === 'published'}<BookOpenCheck class="size-3" />{/if}{c.status?.replace('_', ' ')}</span></td>
							<td class="px-6 py-3 text-xs">{c.view_count}</td>
							<td class="px-6 py-3 text-xs">{c.download_count}</td>
							<td class="px-6 py-3 text-xs text-muted-foreground">{new Date(c.created_at).toLocaleDateString()}</td>
							<td class="px-6 py-3">
								<div class="flex gap-1">
									<Button size="sm" variant="outline" disabled>Edit</Button>
									{#if confirmDelete === c.id}
										<Button size="sm" variant="destructive" onclick={() => deleteComic(c.id)}>Confirm</Button>
										<Button size="sm" variant="ghost" onclick={() => confirmDelete = ''}>Cancel</Button>
									{:else}
										<Button size="sm" variant="ghost" class="text-destructive hover:bg-destructive/10" onclick={() => confirmDelete = c.id}>Delete</Button>
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
