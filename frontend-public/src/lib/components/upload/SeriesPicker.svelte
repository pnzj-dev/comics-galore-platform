<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Input } from '$lib/components/ui/input/index.js';
	import { LoaderCircle, ChevronsUpDown, Plus, Check } from 'lucide-svelte';

	interface SeriesOption {
		id: string;
		title: string;
		category?: string;
	}

	interface Props {
		value?: string | null;
		onSelect: (series: SeriesOption) => void;
		onCreateNew: () => void;
	}

	let { value = null, onSelect, onCreateNew }: Props = $props();

	let open = $state(false);
	let search = $state('');
	let category = $state('');
	let categories = $state<string[]>([]);
	let items = $state<SeriesOption[]>([]);
	let total = $state(0);
	let page = $state(0);
	let loading = $state(false);

	const hasMore = $derived(items.length < total);

	let searchTimer: ReturnType<typeof setTimeout> | undefined;

	function loadCategories() {
		encore.comics.ListSeriesCategories()
			.then((r) => (categories = r.categories || []))
			.catch(() => {});
	}

	async function loadPage(p: number) {
		loading = true;
		try {
			const res = await encore.comics.SearchSeries({
				Search: search || '',
				SearchField: '',
				Category: category || '',
				Page: p,
				Limit: 20,
			});
			const list = (res.series || []).map((s) => ({ id: s.id, title: s.title, category: s.category }));
			items = p === 1 ? list : [...items, ...list];
			total = res.total || 0;
			page = p;
		} catch {
			// ignore — empty state handles it
		} finally {
			loading = false;
		}
	}

	function resetAndLoad() {
		items = [];
		total = 0;
		page = 0;
		loadPage(1);
	}

	function onSearch(e: Event) {
		search = (e.target as HTMLInputElement).value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(resetAndLoad, 300);
	}

	function onCategory(e: Event) {
		category = (e.target as HTMLSelectElement).value;
		resetAndLoad();
	}

	function toggle() {
		open = !open;
		if (open) {
			if (items.length === 0 && page === 0) loadPage(1);
			if (categories.length === 0) loadCategories();
		}
	}

	function choose(s: SeriesOption) {
		onSelect(s);
		open = false;
	}
</script>

<div class="relative">
	<button
		type="button"
		onclick={(e) => {
			e.stopPropagation();
			toggle();
		}}
		class="flex w-full items-center justify-between gap-2 rounded-lg border border-input bg-background px-3 py-2 text-sm text-left"
	>
		<span class="truncate {value ? 'text-foreground' : 'text-muted-foreground'}">
			{value ? items.find((i) => i.id === value)?.title || 'Selected series' : 'Select a series…'}
		</span>
		<ChevronsUpDown class="size-4 shrink-0 text-muted-foreground" />
	</button>

	{#if open}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="absolute z-20 mt-1 w-full rounded-lg border border-border bg-popover shadow-lg"
			onclick={(e) => e.stopPropagation()}
			role="listbox"
			tabindex="-1"
		>
			<div class="border-b border-border p-2 space-y-2">
				<Input
					type="text"
					value={search}
					oninput={onSearch}
					placeholder="Search series…"
					class="h-8 text-sm"
				/>
				{#if categories.length > 0}
					<select value={category} onchange={onCategory} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm">
						<option value="">All categories</option>
						{#each categories as cat (cat)}
							<option value={cat}>{cat}</option>
						{/each}
					</select>
				{/if}
			</div>

			<div class="max-h-56 overflow-y-auto p-1">
				<button type="button" class="flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm font-medium text-primary hover:bg-muted" onclick={() => { onCreateNew(); open = false; }}>
					<Plus class="size-4" /> Create new series
				</button>

				{#if items.length === 0 && !loading}
					<p class="px-2 py-3 text-center text-sm text-muted-foreground">No series found.</p>
				{/if}

				{#each items as s (s.id)}
					<button type="button" class="flex w-full items-center justify-between gap-2 rounded-md px-2 py-2 text-sm hover:bg-muted" onclick={() => choose(s)}>
						<span class="min-w-0">
							<span class="block truncate">{s.title}</span>
							{#if s.category}<span class="block text-xs text-muted-foreground">{s.category}</span>{/if}
						</span>
						{#if value === s.id}<Check class="size-4 shrink-0 text-primary" />{/if}
					</button>
				{/each}

				{#if loading}
					<p class="flex items-center justify-center gap-2 px-2 py-3 text-sm text-muted-foreground">
						<LoaderCircle class="size-4 animate-spin" /> Loading…
					</p>
				{/if}

				{#if hasMore && !loading}
					<button type="button" class="w-full rounded-md px-2 py-2 text-center text-sm text-primary hover:bg-muted" onclick={() => loadPage(page + 1)}>
						Load more ({items.length} / {total})
					</button>
				{/if}
			</div>
		</div>
	{/if}
</div>

<svelte:window onclick={() => (open = false)} />
