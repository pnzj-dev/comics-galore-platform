<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { addToListTarget, clearAddToList } from '$lib/stores/add-to-list.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { LoaderCircle, Plus, Check } from 'lucide-svelte';

	const open = $derived(modal.isOpen('add-to-list'));
	const target = $derived(addToListTarget);

	type ListItem = { id: string; name: string; is_public: boolean; has_comic: boolean; comic_count: number };

	let lists = $state<ListItem[]>([]);
	let loading = $state(true);
	let savingId = $state<string | null>(null);
	let newName = $state('');
	let creating = $state(false);
	let error = $state('');

	async function load() {
		const comicId = target.comicId;
		if (!comicId) return;
		loading = true;
		try {
			const res = await encore.comics.ListReadingLists({ ComicID: comicId });
			lists = res.lists || [];
		} catch {
			lists = [];
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (open && target.comicId) {
			newName = '';
			error = '';
			load();
		}
	});

	async function toggleList(list: ListItem) {
		const comicId = target.comicId;
		if (savingId || !comicId) return;
		savingId = list.id;
		try {
			if (list.has_comic) {
				await encore.comics.RemoveFromReadingList(list.id, comicId);
			} else {
				await encore.comics.AddToReadingList(list.id, { comic_id: comicId });
			}
			await load();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			savingId = null;
		}
	}

	async function createList() {
		if (!newName.trim() || !target.comicId) return;
		creating = true;
		error = '';
		try {
			await encore.comics.CreateReadingList({ name: newName.trim(), is_public: false });
			newName = '';
			await load();
		} catch (e) {
			error = (e as Error).message;
		} finally {
			creating = false;
		}
	}

	function close() {
		modal.close('add-to-list');
		clearAddToList();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open && target.comicId}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4 overflow-y-auto" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-md" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-base font-semibold">Add to reading list</h2>
				<button onclick={close} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">&times;</button>
			</div>

			<div class="p-4 space-y-3">
				<p class="text-sm text-muted-foreground truncate">{target.title}</p>

				{#if loading}
					<div class="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
						<LoaderCircle class="size-4 animate-spin" /> Loading your lists…
					</div>
				{:else}
					{#if lists.length === 0}
						<p class="text-sm text-muted-foreground py-2">You have no lists yet.</p>
					{:else}
						<div class="space-y-1 max-h-56 overflow-y-auto">
							{#each lists as list (list.id)}
								<button
									type="button"
									onclick={() => toggleList(list)}
									disabled={savingId !== null}
									class="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm hover:bg-muted transition-colors"
								>
									<span class="min-w-0 text-left">
										<span class="block truncate">{list.name}</span>
										<span class="block text-xs text-muted-foreground">{list.comic_count} comics{list.is_public ? ' · public' : ''}</span>
									</span>
									{#if savingId === list.id}
										<LoaderCircle class="size-4 animate-spin text-muted-foreground" />
									{:else if list.has_comic}
										<Check class="size-4 text-primary" />
									{:else}
										<span class="size-4 rounded-full border border-border"></span>
									{/if}
								</button>
							{/each}
						</div>
					{/if}

					<div class="flex gap-2 pt-1">
						<Input bind:value={newName} placeholder="New list name" class="h-9 text-sm" />
						<Button size="sm" variant="outline" onclick={createList} disabled={creating || !newName.trim()}>
							<Plus class="size-4 mr-1" /> Create
						</Button>
					</div>
				{/if}

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
			</div>

			<div class="flex justify-end p-4 pt-0">
				<Button size="sm" onclick={close}>Done</Button>
			</div>
		</div>
	</div>
{/if}
