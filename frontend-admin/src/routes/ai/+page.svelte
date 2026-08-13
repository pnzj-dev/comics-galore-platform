<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let queue = $state(data.queue);
	// svelte-ignore state_referenced_locally
	let decisions = $state(data.decisions);

	async function resolveItem(id: string, action: 'approve' | 'reject') {
		await encore.comics.ResolveAIReview(id, { action });
		queue = queue.filter((i) => i.id !== id);
	}
</script>

<svelte:head><title>AI Moderation - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">AI Moderation</h1>

	<h2 class="text-xl font-semibold mb-3">Review Queue ({queue.length})</h2>
	{#if queue.length === 0}
		<p class="text-muted-foreground text-sm mb-8">No uncertain items awaiting review.</p>
	{:else}
		<div class="space-y-3 mb-8">
			{#each queue as item}
				<Card>
					<CardHeader>
						<div class="flex items-center justify-between gap-3">
							<div class="flex-1 min-w-0">
								<CardTitle class="truncate">{item.preview}</CardTitle>
								<p class="text-xs text-muted-foreground mt-1">{item.target_type} · {item.target_id}</p>
							</div>
							<div class="flex gap-2 shrink-0">
								<Button size="sm" onclick={() => resolveItem(item.id, 'approve')}>Approve</Button>
								<Button size="sm" variant="destructive" onclick={() => resolveItem(item.id, 'reject')}>Reject</Button>
							</div>
						</div>
					</CardHeader>
				</Card>
			{/each}
		</div>
	{/if}

	<h2 class="text-xl font-semibold mb-3">Decision Log</h2>
	{#if decisions.length === 0}
		<p class="text-muted-foreground text-sm">No AI decisions recorded yet.</p>
	{:else}
		<div class="space-y-2">
			{#each decisions as d}
				<div class="flex items-center gap-3 p-3 rounded-lg border border-border">
					<span class="text-xs uppercase font-medium px-2 py-0.5 rounded-full {d.decision === 'approved' ? 'bg-green-100 text-green-800' : d.decision === 'rejected' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'}">{d.decision}</span>
					<span class="text-sm truncate flex-1">{d.reason || d.target_id}</span>
					<span class="text-xs text-muted-foreground">{Math.round(d.confidence * 100)}% · {d.model}</span>
				</div>
			{/each}
		</div>
	{/if}
</section>
