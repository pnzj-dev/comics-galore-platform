<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let pending = $state(data.pending);
	let selected = $state<Set<string>>(new Set());

	async function approve(id: string) {
		await encore.comics.ApproveComic(id);
		pending = pending.filter(c => c.id !== id);
	}

	async function rejectWithReason(id: string) {
		const reason = prompt('Rejection reason:');
		if (!reason) return;
		await encore.comics.RejectComic(id, { reason });
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
		await encore.comics.BulkModerate({ ids: [...selected], action: 'approve' as any });
		pending = pending.filter(c => !selected.has(c.id));
		selected = new Set();
	}

	async function bulkReject() {
		await encore.comics.BulkModerate({ ids: [...selected], action: 'reject' as any });
		pending = pending.filter(c => !selected.has(c.id));
		selected = new Set();
	}
</script>

<svelte:head><title>Moderation - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Moderation Queue</h1>

	{#if pending.length === 0}
		<div class="text-center py-12"><p class="text-lg text-muted-foreground">No comics pending review.</p></div>
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
				<label class="flex items-center gap-2 text-sm cursor-pointer"><input type="checkbox" checked={selected.size === pending.length} onchange={selectAll} class="rounded" /> Select all</label>
			</div>
			{#each pending as comic}
				<Card class={selected.has(comic.id) ? 'ring-2 ring-primary' : ''}>
					<CardHeader><div class="flex items-start gap-3"><input type="checkbox" checked={selected.has(comic.id)} onchange={() => toggleSelect(comic.id)} class="mt-1.5 rounded" /><div class="flex-1"><CardTitle>{comic.title}</CardTitle><p class="text-sm text-muted-foreground">Submitted: {new Date(comic.created_at).toLocaleDateString()}</p></div></div></CardHeader>
					<CardContent><div class="flex gap-2"><Button size="sm" onclick={() => approve(comic.id)}>Approve</Button><Button size="sm" variant="destructive" onclick={() => rejectWithReason(comic.id)}>Reject</Button></div></CardContent>
				</Card>
			{/each}
		</div>
	{/if}
</section>
