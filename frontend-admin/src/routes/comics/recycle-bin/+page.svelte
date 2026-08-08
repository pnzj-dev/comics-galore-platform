<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { RefreshCw, Trash2 } from 'lucide-svelte';

	let comics = $state<any[]>([]);
	let loading = $state(true);
	let actionLoading = $state('');

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		await loadRecycleBin();
	});

	async function loadRecycleBin() {
		loading = true;
		try { const res = await api.get<{ comics: any[] }>('/admin/recycle-bin'); comics = res.comics; } catch {}
		loading = false;
	}

	async function restoreComic(id: string) {
		actionLoading = id;
		await api.post(`/admin/comics/${id}/restore`, {});
		comics = comics.filter(c => c.id !== id);
		actionLoading = '';
	}

	async function permanentDelete(id: string) {
		actionLoading = id;
		await api.delete(`/comics/${id}`);
		comics = comics.filter(c => c.id !== id);
		actionLoading = '';
	}
</script>

<svelte:head><title>Recycle Bin — Admin</title></svelte:head>

<section class="py-8">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold">Recycle Bin</h1>
		<div class="flex gap-2">
			<Button variant="outline" size="sm" onclick={loadRecycleBin}>
				<RefreshCw class="size-4 mr-1" /> Refresh
			</Button>
			<Button variant="ghost" size="sm" href="/admin/comics">Back to Comics</Button>
		</div>
	</div>

	{#if loading}
		<div class="rounded-xl border border-border overflow-hidden animate-pulse">
			<div class="divide-y divide-border">
				{#each Array(5) as _}
					<div class="flex items-center gap-4 px-6 py-3">
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 flex-1"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
						<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-20"></div>
					</div>
				{/each}
			</div>
		</div>
	{:else if comics.length === 0}
		<div class="text-center py-12">
			<div class="text-4xl mb-4"><Trash2 class="size-10 mx-auto text-muted-foreground/40" /></div>
			<p class="text-muted-foreground">Recycle bin is empty. Deleted comics will appear here before permanent deletion.</p>
		</div>
	{:else}
		<div class="rounded-xl border border-border overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b bg-muted/50 text-left">
						<th class="px-6 py-3 font-medium">Title</th>
						<th class="px-6 py-3 font-medium">Author</th>
						<th class="px-6 py-3 font-medium">Deleted</th>
						<th class="px-6 py-3 font-medium w-48">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each comics as c}
						<tr class="hover:bg-muted/30">
							<td class="px-6 py-3">
								<a href={`http://localhost:5173/comics/${c.slug}`} target="_blank" class="hover:text-primary">{c.title}</a>
							</td>
							<td class="px-6 py-3 text-xs text-muted-foreground">{c.author || '—'}</td>
							<td class="px-6 py-3 text-xs text-muted-foreground">{c.deleted_at ? new Date(c.deleted_at).toLocaleDateString() : '—'}</td>
							<td class="px-6 py-3">
								<div class="flex gap-1">
									<Button size="sm" variant="outline" onclick={() => restoreComic(c.id)} disabled={actionLoading === c.id}>Restore</Button>
									<Button size="sm" variant="destructive" onclick={() => permanentDelete(c.id)} disabled={actionLoading === c.id}>Delete</Button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
