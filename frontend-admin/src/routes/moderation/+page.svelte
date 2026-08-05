<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let pending = $state<any[]>([]);
	let loading = $state(true);
	let selected = $state<Set<string>>(new Set());

	onMount(async () => {
		if (!$currentUser || ($currentUser.role !== 'moderator' && $currentUser.role !== 'admin')) {
			await goto('/');
			return;
		}
		await loadPending();
	});

	async function loadPending() {
		try {
			const res = await api.get<{ comics: any[] }>('/moderation/comics');
			pending = res.comics;
		} catch {}
		loading = false;
	}

	async function approve(id: string) {
		await api.post(`/moderation/comics/${id}/approve`);
		pending = pending.filter(c => c.id !== id);
	}

	async function rejectWithReason(id: string) {
		const reason = prompt('Rejection reason:');
		if (!reason) return;
		await api.post(`/moderation/comics/${id}/reject`, { reason });
		pending = pending.filter(c => c.id !== id);
	}

	function toggleSelect(id: string) {
		const next = new Set(selected);
		if (next.has(id)) next.delete(id); else next.add(id);
		selected = next;
	}

	function selectAll() {
		if (selected.size === pending.length) { selected = new Set(); return; }
		selected = new Set(pending.map(c => c.id));
	}

	async function bulkApprove() {
		await api.post('/moderation/bulk', { ids: [...selected], action: 'approve' });
		pending = pending.filter(c => !selected.has(c.id));
		selected = new Set();
	}

	async function bulkReject() {
		await api.post('/moderation/bulk', { ids: [...selected], action: 'reject' });
		pending = pending.filter(c => !selected.has(c.id));
		selected = new Set();
	}
</script>

<svelte:head><title>Moderation - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Moderation Queue</h1>

	{#if loading}
		<p class="text-muted-foreground">Loading...</p>
	{:else if pending.length === 0}
		<div class="text-center py-12">
			<p class="text-lg text-muted-foreground">No comics pending review.</p>
		</div>
	{:else}
		{#if selected.size > 0}
			<div class="flex items-center gap-3 mb-4 p-3 bg-muted rounded-lg">
				<span class="text-sm">{selected.size} selected</span>
				<Button size="sm" onclick={bulkApprove}>Approve All</Button>
				<Button size="sm" variant="destructive" onclick={bulkReject}>Reject All</Button>
			</div>
		{/if}

		<div class="space-y-3">
			<div class="flex items-center gap-2 mb-2">
				<label class="flex items-center gap-2 text-sm cursor-pointer">
					<input type="checkbox" checked={selected.size === pending.length} onchange={selectAll} class="rounded" />
					Select all
				</label>
			</div>

			{#each pending as comic}
				<Card class={selected.has(comic.id) ? 'ring-2 ring-primary' : ''}>
					<CardHeader>
						<div class="flex items-start gap-3">
							<input type="checkbox" checked={selected.has(comic.id)} onchange={() => toggleSelect(comic.id)} class="mt-1.5 rounded" />
							<div class="flex-1">
								<CardTitle>{comic.title}</CardTitle>
								<p class="text-sm text-muted-foreground">Submitted: {new Date(comic.created_at).toLocaleDateString()}</p>
							</div>
						</div>
					</CardHeader>
					<CardContent>
						<div class="flex gap-2">
							<Button size="sm" onclick={() => approve(comic.id)}>Approve</Button>
							<Button size="sm" variant="destructive" onclick={() => rejectWithReason(comic.id)}>Reject</Button>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
</section>
