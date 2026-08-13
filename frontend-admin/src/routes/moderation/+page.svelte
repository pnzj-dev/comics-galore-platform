<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let pending = $state(data.results);
	let selected = $state<Set<string>>(new Set());

	// svelte-ignore state_referenced_locally
	let flagged = $state(data.flags);

	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));

	// svelte-ignore state_referenced_locally
	let searchTerm = $state(data.search);
	let searchTimer = $state<ReturnType<typeof setTimeout>>();

	function handleSearchInput(value: string) {
		searchTerm = value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			const url = new URL(page.url);
			url.searchParams.set('page', '1');
			if (value) url.searchParams.set('search', value);
			else url.searchParams.delete('search');
			goto(url.pathname + url.search, { keepFocus: true });
		}, 300);
	}

	function handlePage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true });
	}

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

	async function resolveFlag(flagId: string) {
		await encore.comics.ResolveFlag(flagId);
		flagged = flagged.filter(f => f.flag_id !== flagId);
	}
</script>

<svelte:head><title>Moderation - Comics Galore</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Moderation Queue</h1>

	{#if flagged.length > 0}
		<div class="mb-8">
			<h2 class="text-xl font-semibold mb-3">Flagged Comments ({flagged.length})</h2>
			<div class="space-y-3">
				{#each flagged as flag (flag.flag_id)}
					<Card>
						<CardHeader>
							<div class="flex items-start justify-between gap-3">
								<div class="flex-1">
									<CardTitle>{flag.comic_title}</CardTitle>
									<p class="text-sm mt-1">{flag.body_text}</p>
									<div class="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
										<span>{flag.flag_count} flag{flag.flag_count !== 1 ? 's' : ''}</span>
										{#if flag.reason}<span>Reason: {flag.reason}</span>{/if}
										<span>Flagged {formatDate(flag.created_at)}</span>
									</div>
								</div>
								<div class="flex gap-2 shrink-0">
									<Button size="sm" variant="outline" onclick={() => resolveFlag(flag.flag_id)}>Resolve</Button>
									<Button size="sm" variant="destructive" onclick={async () => { await encore.comics.DeleteComment(flag.comment_id); await resolveFlag(flag.flag_id); }}>Delete</Button>
								</div>
							</div>
						</CardHeader>
					</Card>
				{/each}
			</div>
		</div>
	{/if}

	<div class="flex items-center gap-3 mb-4">
		<input
			type="text"
			placeholder="Search pending comics..."
			value={searchTerm}
			oninput={(e) => handleSearchInput((e.target as HTMLInputElement).value)}
			class="max-w-xs rounded-md border border-input bg-background px-3 py-1.5 text-sm"
		/>
	</div>

	{#if selected.size > 0}
		<div class="flex items-center gap-3 mb-4 p-3 bg-muted rounded-lg">
			<span class="text-sm">{selected.size} selected</span>
			<Button size="sm" onclick={bulkApprove}>Approve All</Button>
			<Button size="sm" variant="destructive" onclick={bulkReject}>Reject All</Button>
		</div>
	{/if}

	{#if pending.length === 0}
		<div class="text-center py-12"><p class="text-lg text-muted-foreground">No comics pending review.</p></div>
	{:else}
		<div class="space-y-3">
			<div class="flex items-center gap-2 mb-2">
				<label class="flex items-center gap-2 text-sm cursor-pointer">
					<input type="checkbox" checked={selected.size === pending.length} onchange={selectAll} class="rounded" /> Select all
				</label>
			</div>
			{#each pending as comic}
				<Card class={selected.has(comic.id) ? 'ring-2 ring-primary' : ''}>
					<CardHeader>
						<div class="flex items-start gap-3">
							<input type="checkbox" checked={selected.has(comic.id)} onchange={() => toggleSelect(comic.id)} class="mt-1.5 rounded" />
							<div class="flex-1">
								<CardTitle>{comic.title}</CardTitle>
								<p class="text-sm text-muted-foreground">Submitted: {formatDate(comic.created_at)}</p>
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

		<div class="flex items-center justify-between text-sm text-muted-foreground mt-4">
			<span>{data.total} result{data.total !== 1 ? 's' : ''}</span>
			<div class="flex items-center gap-2">
				<Button variant="outline" size="sm" disabled={data.page <= 1} onclick={() => handlePage(data.page - 1)}>
					← Prev
				</Button>
				<span class="text-xs">Page {data.page} of {totalPages}</span>
				<Button variant="outline" size="sm" disabled={data.page >= totalPages} onclick={() => handlePage(data.page + 1)}>
					Next →
				</Button>
			</div>
		</div>
	{/if}
</section>
