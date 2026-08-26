<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import ComicCard from '$lib/components/comics/ComicCard.svelte';
	import Pagination from '$lib/components/common/Pagination.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Pencil, Check, X, Trash2 } from 'lucide-svelte';

	let { data } = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(data.total / data.limit)));

	let editing = $state(false);
	// svelte-ignore state_referenced_locally
	let name = $state(data.list?.name || '');
	let confirmDelete = $state(false);
	let busy = $state(false);
	let error = $state('');
	let resultsRef = $state<HTMLDivElement | null>(null);

	function goPage(p: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', String(p));
		goto(url.pathname + url.search, { keepFocus: true, noScroll: true }).then(() => {
			resultsRef?.scrollIntoView({ block: 'start' });
		});
	}

	async function saveName() {
		if (!name.trim() || !data.list) return;
		busy = true;
		error = '';
		try {
			await encore.comics.UpdateReadingList(data.list.id, { name: name.trim(), is_public: data.list.is_public });
			editing = false;
			await invalidateAll();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}

	async function togglePublic() {
		if (!data.list) return;
		busy = true;
		error = '';
		try {
			await encore.comics.UpdateReadingList(data.list.id, { name: data.list.name, is_public: !data.list.is_public });
			await invalidateAll();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}

	async function deleteList() {
		if (!data.list) return;
		busy = true;
		try {
			await encore.comics.DeleteReadingList(data.list.id);
			await goto('/lists');
		} catch (e) {
			error = (e as Error).message;
			busy = false;
		}
	}

	async function removeComic(comicId: string) {
		if (!data.list) return;
		busy = true;
		error = '';
		try {
			await encore.comics.RemoveFromReadingList(data.list.id, comicId);
			await invalidateAll();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{data.list?.name || 'Reading List'} - Comics Galore</title></svelte:head>

<section class="py-8">
	{#if data.list}
		<div class="mb-6">
			<a href="/lists" class="text-sm text-muted-foreground hover:text-foreground">&larr; My lists</a>

			<div class="mt-2 flex items-center gap-3 flex-wrap">
				{#if editing}
					<div class="flex items-center gap-1">
						<Input bind:value={name} class="h-9 max-w-xs" />
						<Button size="sm" variant="ghost" onclick={saveName} disabled={busy}><Check class="size-4 text-green-600" /></Button>
						<Button size="sm" variant="ghost" onclick={() => { editing = false; name = data.list.name; }}><X class="size-4" /></Button>
					</div>
				{:else}
					<h1 class="text-3xl font-bold">{data.list.name}</h1>
					{#if data.isOwner}
						<button onclick={() => { editing = true; name = data.list.name; }} class="p-1.5 rounded hover:bg-muted" aria-label="Rename"><Pencil class="size-4" /></button>
					{/if}
				{/if}

				{#if data.list.is_public}
					<span class="text-[10px] uppercase bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 px-1.5 py-0.5 rounded-full">public</span>
				{/if}

				{#if data.isOwner}
					<label class="flex items-center gap-1.5 text-sm text-muted-foreground ml-auto cursor-pointer">
						<Checkbox checked={data.list.is_public} onchange={togglePublic} disabled={busy} /> Public
					</label>
					{#if confirmDelete}
						<Button size="sm" variant="destructive" onclick={deleteList} disabled={busy}>Confirm delete</Button>
						<Button size="sm" variant="ghost" onclick={() => (confirmDelete = false)}>Cancel</Button>
					{:else}
						<Button size="sm" variant="ghost" class="text-destructive hover:bg-destructive/10" onclick={() => (confirmDelete = true)}><Trash2 class="size-4" /></Button>
					{/if}
				{/if}
			</div>

			{#if error}<p class="text-sm text-destructive mt-2">{error}</p>{/if}
			<p class="text-sm text-muted-foreground mt-1">{data.total} comic{data.total === 1 ? '' : 's'}</p>
		</div>

		{#if data.comics.length === 0}
			<p class="text-muted-foreground">This list is empty.</p>
		{:else}
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6 scroll-mt-16" bind:this={resultsRef}>
				{#each data.comics as comic}
					<div class="relative">
						<ComicCard {...comic} />
						{#if data.isOwner}
							<button
								onclick={() => removeComic(comic.id)}
								class="absolute top-2 right-2 z-10 size-6 rounded-full bg-black/70 text-white flex items-center justify-center hover:bg-red-600 transition-colors"
								aria-label="Remove from list"
							>
								<X class="size-3.5" />
							</button>
						{/if}
					</div>
				{/each}
			</div>
			<Pagination page={data.page} {totalPages} onPage={goPage} />
		{/if}
	{:else}
		<p class="text-destructive text-center py-20">List not found.</p>
	{/if}
</section>
