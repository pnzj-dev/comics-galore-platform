<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let pending = $state<any[]>([]);
	let loading = $state(true);

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
		} catch { /* */ }
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
</script>

<svelte:head>
	<title>Moderation - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Moderation Queue</h1>

	{#if loading}
		<p class="text-muted-foreground">Loading...</p>
	{:else if pending.length === 0}
		<div class="text-center py-12">
			<p class="text-lg text-muted-foreground">No comics pending review.</p>
		</div>
	{:else}
		<div class="space-y-4">
			{#each pending as comic}
				<Card>
					<CardHeader>
						<CardTitle>{comic.title}</CardTitle>
						<p class="text-sm text-muted-foreground">Submitted: {new Date(comic.created_at).toLocaleDateString()}</p>
					</CardHeader>
					<CardContent>
						<div class="flex gap-2">
							<Button onclick={() => approve(comic.id)}>Approve</Button>
							<Button variant="destructive" onclick={() => rejectWithReason(comic.id)}>Reject</Button>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
</section>
