<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();
</script>

<svelte:head><title>Series — Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Series</h1>

	{#if data.series.length === 0}
		<p class="text-muted-foreground text-center py-12">No series created yet.</p>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.series as s}
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
