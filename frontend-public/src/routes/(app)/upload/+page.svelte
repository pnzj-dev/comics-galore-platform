<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import CompactCard from '$lib/components/CompactCard.svelte';
	import ComicForm from '$lib/components/ComicForm.svelte';
	import ArchiveForm from '$lib/components/ArchiveForm.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let tab = $state<'list' | 'manual' | 'archive'>('list');
	let myComics = $state<any[]>([]);
	let loading = $state(true);
	let activeSessions = $state<any[]>([]);

	const user = $derived($currentUser);

	onMount(async () => {
		if (!$currentUser || ($currentUser.role !== 'uploader' && $currentUser.role !== 'admin')) {
			await goto('/');
			return;
		}
		await loadComics();
		loadSessions();
	});

	async function loadComics() {
		try {
			const res = await api.get<{ comics: any[] }>('/uploader/comics');
			myComics = res.comics;
		} catch {}
		loading = false;
	}

	async function loadSessions() {
		try {
			const res = await api.get<{ sessions: any[] }>('/upload-sessions');
			activeSessions = res.sessions;
		} catch {}
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
		<Button variant={tab === 'manual' ? 'default' : 'outline'} onclick={() => tab = 'manual'}>
			Manual
		</Button>
		<Button variant={tab === 'archive' ? 'default' : 'outline'} onclick={() => tab = 'archive'}>
			Archive
		</Button>
	</div>

	{#if activeSessions.length > 0}
		<div class="mb-6 p-3 rounded-lg bg-yellow-50 dark:bg-yellow-900/10 border border-yellow-200 dark:border-yellow-800">
			<p class="text-sm font-medium text-yellow-700 dark:text-yellow-400">
				You have {activeSessions.length} active upload session{activeSessions.length > 1 ? 's' : ''}.
				{activeSessions[0].parts?.length > 0 ? ` ${activeSessions[0].parts.length} file(s) already uploaded.` : ''}
			</p>
			<p class="text-xs text-yellow-600 dark:text-yellow-500 mt-1">Return to the same tab to continue where you left off. Your uploaded files are preserved.</p>
		</div>
	{/if}

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
				<Button class="mt-4" onclick={() => tab = 'manual'}>Create Your First Comic</Button>
			</div>
		{:else}
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
				{#each myComics as comic}
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
