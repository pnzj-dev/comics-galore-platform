<script lang="ts">
	import { goto } from '$app/navigation';
	import { currentUser } from '$lib/stores/auth';
	import CompactCard from '$lib/components/CompactCard.svelte';
	import ComicForm from '$lib/components/ComicForm.svelte';
	import ArchiveForm from '$lib/components/ArchiveForm.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { data } = $props();

	const user = $derived($currentUser);
	const tab = $derived(data.tab as 'list' | 'manual' | 'archive');
	const comics = $derived(data.comics);
	const activeSessions = $derived(data.activeSessions);

	async function switchTab(t: 'list' | 'manual' | 'archive') {
		await goto(`/upload?tab=${t}`);
	}
</script>

<svelte:head>
	<title>Upload - Comics Galore</title>
</svelte:head>

<section class="py-8">
	<div class="flex items-center gap-4 mb-6">
		<Button variant={tab === 'list' ? 'default' : 'outline'} onclick={() => switchTab('list')}>
			My Comics
		</Button>
		<Button variant={tab === 'manual' ? 'default' : 'outline'} onclick={() => switchTab('manual')}>
			Manual
		</Button>
		<Button variant={tab === 'archive' ? 'default' : 'outline'} onclick={() => switchTab('archive')}>
			Archive
		</Button>
	</div>

	{#if activeSessions.length > 0}
		<div class="mb-6 p-4 rounded-lg bg-yellow-50 dark:bg-yellow-900/10 border border-yellow-200 dark:border-yellow-800 flex items-center justify-between">
			<div>
				<p class="text-sm font-medium text-yellow-700 dark:text-yellow-400">
					You have {activeSessions.length} active upload session{activeSessions.length > 1 ? 's' : ''}.
					{activeSessions[0]?.parts?.length > 0 ? ` ${activeSessions[0].parts.length} file(s) already uploaded.` : ''}
				</p>
				<p class="text-xs text-yellow-600 dark:text-yellow-500 mt-1">Switch to Manual or Archive tab to continue where you left off.</p>
			</div>
			<div class="flex gap-2 shrink-0">
				{#if activeSessions[0]?.parts?.length > 0}
					<Button size="sm" variant="outline" onclick={() => switchTab('manual')}>Resume Manual</Button>
				{/if}
				{#if activeSessions[0]?.mode === 'archive'}
					<Button size="sm" variant="outline" onclick={() => switchTab('archive')}>Resume Archive</Button>
				{/if}
			</div>
		</div>
	{/if}

	{#if tab === 'list'}
		<h2 class="text-2xl font-semibold mb-4">My Comics</h2>
		{#if comics.length === 0}
			<div class="text-center py-12">
				<p class="text-muted-foreground">You haven't created any comics yet.</p>
				<Button class="mt-4" onclick={() => switchTab('manual')}>Create Your First Comic</Button>
			</div>
		{:else}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
				{#each comics as comic}
					<CompactCard {...comic} />
				{/each}
			</div>
		{/if}
	{:else if tab === 'manual'}
		<h2 class="text-2xl font-semibold mb-4">New Comic</h2>
		<ComicForm />
	{:else if tab === 'archive'}
		<h2 class="text-2xl font-semibold mb-4">Upload Archive</h2>
		<ArchiveForm />
	{/if}
</section>
