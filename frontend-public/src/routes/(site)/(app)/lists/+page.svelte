<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Pencil, Trash2, Check, X, List } from 'lucide-svelte';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let lists = $state(data.lists);
	let name = $state('');
	let isPublic = $state(false);
	let error = $state('');

	let editingId = $state<string | null>(null);
	let editingName = $state('');
	let confirmDelete = $state<string | null>(null);
	let busy = $state(false);

	async function refresh() {
		const res = await encore.comics.ListReadingLists({ ComicID: '' });
		lists = res.lists || [];
	}

	async function createList() {
		if (!name.trim()) {
			error = 'Name is required';
			return;
		}
		error = '';
		busy = true;
		try {
			await encore.comics.CreateReadingList({ name: name.trim(), is_public: isPublic });
			name = '';
			isPublic = false;
			await refresh();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}

	function startEdit(list: { id: string; name: string }) {
		editingId = list.id;
		editingName = list.name;
	}

	async function saveEdit(list: { id: string; is_public: boolean }) {
		if (!editingName.trim()) return;
		busy = true;
		try {
			await encore.comics.UpdateReadingList(list.id, { name: editingName.trim(), is_public: list.is_public });
			editingId = null;
			await refresh();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}

	async function togglePublic(list: { id: string; name: string; is_public: boolean }) {
		busy = true;
		try {
			await encore.comics.UpdateReadingList(list.id, { name: list.name, is_public: !list.is_public });
			await refresh();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}

	async function deleteList(id: string) {
		busy = true;
		try {
			await encore.comics.DeleteReadingList(id);
			confirmDelete = null;
			await refresh();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>Reading Lists - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Reading Lists</h1>

	<div class="flex items-end gap-3 mb-6">
		<div class="space-y-1.5">
			<label for="list-name" class="text-sm text-muted-foreground">New list</label>
			<Input id="list-name" bind:value={name} placeholder="Favorites" />
		</div>
		<label class="flex items-center gap-2 text-sm pb-1.5">
			<Checkbox bind:checked={isPublic} /> Public
		</label>
		<Button onclick={createList} disabled={busy}>Create</Button>
		{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
	</div>

	{#if lists.length === 0}
		<div class="text-center py-16 text-muted-foreground">
			<List class="size-10 mx-auto mb-3 opacity-50" />
			<p class="text-sm">No lists yet. Create one to start curating a shelf of comics.</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each lists as list (list.id)}
				<div class="p-4 rounded-lg border border-border space-y-3">
					<div class="flex items-start justify-between gap-2">
						{#if editingId === list.id}
							<div class="flex items-center gap-1 flex-1">
								<Input bind:value={editingName} class="h-8 text-sm" />
								<button onclick={() => saveEdit(list)} class="p-1.5 rounded hover:bg-muted" aria-label="Save"><Check class="size-4 text-green-600" /></button>
								<button onclick={() => (editingId = null)} class="p-1.5 rounded hover:bg-muted" aria-label="Cancel"><X class="size-4" /></button>
							</div>
						{:else}
							<a href="/lists/{list.id}" class="font-medium hover:text-primary truncate">{list.name}</a>
							<div class="flex items-center gap-1 shrink-0">
								<button onclick={() => startEdit(list)} class="p-1.5 rounded hover:bg-muted" aria-label="Rename"><Pencil class="size-3.5" /></button>
								{#if confirmDelete === list.id}
									<button onclick={() => deleteList(list.id)} class="p-1.5 rounded hover:bg-destructive/10 text-destructive" aria-label="Confirm delete"><Check class="size-3.5" /></button>
									<button onclick={() => (confirmDelete = null)} class="p-1.5 rounded hover:bg-muted" aria-label="Cancel"><X class="size-3.5" /></button>
								{:else}
									<button onclick={() => (confirmDelete = list.id)} class="p-1.5 rounded hover:bg-destructive/10 text-muted-foreground" aria-label="Delete"><Trash2 class="size-3.5" /></button>
								{/if}
							</div>
						{/if}
					</div>

					<div class="flex items-center justify-between text-xs text-muted-foreground">
						<span>{list.comic_count} comic{list.comic_count === 1 ? '' : 's'}</span>
						<label class="flex items-center gap-1.5 cursor-pointer">
							<Checkbox checked={list.is_public} onchange={() => togglePublic(list)} disabled={busy} />
							Public
						</label>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</section>
