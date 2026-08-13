<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let lists = $state(data.lists);
	let name = $state('');
	let isPublic = $state(false);
	let error = $state('');

	async function createList() {
		if (!name.trim()) {
			error = 'Name is required';
			return;
		}
		error = '';
		try {
			await encore.comics.CreateReadingList({ name, is_public: isPublic });
			name = '';
			isPublic = false;
			const res = await encore.comics.ListReadingLists();
			lists = res.lists || [];
		} catch (e) {
			error = (e as Error).message;
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
			<input type="checkbox" bind:checked={isPublic} class="rounded" /> Public
		</label>
		<Button onclick={createList}>Create</Button>
		{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
	</div>

	{#if lists.length === 0}
		<p class="text-muted-foreground text-sm">No lists yet. Create one to start a shelf.</p>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each lists as list}
				<div class="p-4 rounded-lg border border-border">
					<div class="flex items-center justify-between">
						<span class="font-medium">{list.name}</span>
						{#if list.is_public}
							<span class="text-[10px] uppercase bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 px-1.5 py-0.5 rounded-full">public</span>
						{/if}
					</div>
					{#if list.is_public}
						<a href="/lists/{list.id}" class="text-sm text-primary hover:underline mt-2 inline-block">View public page</a>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</section>
