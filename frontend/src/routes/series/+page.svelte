<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let series = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await api.get<{ series: any[] }>('/series');
			series = res.series;
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Series — Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Series</h1>

	{#if loading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _}
				<div class="rounded-xl border border-border p-4 animate-pulse space-y-2">
					<div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-2/3"></div>
					<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
				</div>
			{/each}
		</div>
	{:else if series.length === 0}
		<p class="text-muted-foreground text-center py-12">No series created yet.</p>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each series as s}
				<a href="/series/{s.id}">
					<Card class="hover:border-primary/50 transition-colors h-full">
						<CardHeader>
							<CardTitle>{s.title}</CardTitle>
						</CardHeader>
						<CardContent>
							<p class="text-sm text-muted-foreground line-clamp-2">{s.description || 'No description'}</p>
						</CardContent>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</section>
