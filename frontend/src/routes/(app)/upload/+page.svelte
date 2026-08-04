<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import CompactCard from '$lib/components/CompactCard.svelte';
	import ComicForm from '$lib/components/ComicForm.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let tab = $state<'list' | 'create'>('list');
	let myComics = $state<any[]>([]);
	let loading = $state(true);

	const user = $derived($currentUser);

	onMount(async () => {
		if (!$currentUser || ($currentUser.role !== 'uploader' && $currentUser.role !== 'admin')) {
			await goto('/');
			return;
		}
		await loadComics();
	});

	async function loadComics() {
		try {
			const res = await api.get<{ comics: any[] }>('/uploader/comics');
			myComics = res.comics;
		} catch { /* */ }
		loading = false;
	}
</script>

<svelte:head>
	<title>Upload - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<div class="flex items-center gap-4 mb-6">
		<Button variant={tab === 'list' ? 'default' : 'outline'} onclick={() => { tab = 'list'; loading = true; loadComics(); }}>
			My Comics
		</Button>
		<Button variant={tab === 'create' ? 'default' : 'outline'} onclick={() => tab = 'create'}>
			New Comic
		</Button>
	</div>

	{#if tab === 'list'}
		<h2 class="text-2xl font-semibold mb-4">My Comics</h2>
		{#if loading}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
				{#each Array(10) as _}
					<div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden animate-pulse">
						<div class="aspect-[2/3] bg-gray-200 dark:bg-gray-700"></div>
						<div class="p-2 space-y-1.5">
							<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
							<div class="h-2 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
						</div>
					</div>
				{/each}
			</div>
		{:else if myComics.length === 0}
			<div class="text-center py-12">
				<p class="text-muted-foreground">You haven't created any comics yet.</p>
				<Button class="mt-4" onclick={() => tab = 'create'}>Create Your First Comic</Button>
			</div>
		{:else}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
				{#each myComics as comic}
					<CompactCard {...comic} />
				{/each}
			</div>
		{/if}
	{:else}
		<ComicForm />
	{/if}
</section>
