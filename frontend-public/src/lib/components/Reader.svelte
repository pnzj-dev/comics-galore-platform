<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Maximize, Columns, ArrowDownToLine } from 'lucide-svelte';

	interface Props {
		comicId: string;
		pageKeys: string[];
		pageUrls?: string[];
		totalPages: number;
	}

	let { comicId, pageKeys, pageUrls, totalPages }: Props = $props();

	type FitMode = 'height' | 'width' | 'original';

	let currentPage = $state(0);
	let savedProgress = $state(false);
	let fitMode = $state<FitMode>('height');
	let showThumbnails = $state(true);
	let autosaveTimer: ReturnType<typeof setTimeout>;

	const pageCount = $derived(totalPages || pageKeys.length || 1);
	const pageImages = $derived(pageUrls || pageKeys.map(k => `http://localhost:4000/media/${k}`));

	const fitClass = $derived(fitMode === 'height' ? 'max-h-[85vh] w-auto object-contain' :
		fitMode === 'width' ? 'w-full max-h-none object-contain' :
		'max-h-[90vh] max-w-[95vw] object-contain');

	function nextPage() { if (currentPage < pageCount - 1) currentPage++; }
	function prevPage() { if (currentPage > 0) currentPage--; }
	function goToPage(p: number) { currentPage = p; }

	function cycleFit() {
		if (fitMode === 'height') fitMode = 'width';
		else if (fitMode === 'width') fitMode = 'original';
		else fitMode = 'height';
	}

	function saveProgress() {
		clearTimeout(autosaveTimer);
		autosaveTimer = setTimeout(async () => {
			try {
				await api.post(`/reading/${comicId}`, {
					current_page: currentPage,
					total_pages: pageCount,
					completed: currentPage >= pageCount - 1
				});
				savedProgress = true;
				setTimeout(() => savedProgress = false, 2000);
			} catch {}
		}, 1000);
	}

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case 'ArrowRight': case ' ': e.preventDefault(); nextPage(); saveProgress(); break;
			case 'ArrowLeft': e.preventDefault(); prevPage(); saveProgress(); break;
			case 'Escape': window.location.href = `/comics/${comicId}`; break;
			case 'f': cycleFit(); break;
			case 't': showThumbnails = !showThumbnails; break;
		}
	}

	$effect(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});

	onMount(async () => {
		try {
			const progress = await api.get<{ current_page: number }>(`/reading/${comicId}`);
			currentPage = progress.current_page || 0;
		} catch {}
	});
</script>

<div class="fixed inset-0 z-50 bg-black flex flex-col">
	<!-- Header -->
	<div class="flex items-center justify-between px-3 py-2 bg-black/90 text-white text-sm flex-shrink-0">
		<div class="flex items-center gap-3">
			<span class="tabular-nums">{currentPage + 1} / {pageCount}</span>
			{#if savedProgress}<span class="text-xs text-green-400">Saved</span>{/if}

			<div class="flex items-center gap-1 ml-2">
				<button onclick={cycleFit} class="px-2 py-1 rounded hover:bg-white/10 text-xs flex items-center gap-1" title="Fit mode (F)">
					{#if fitMode === 'height'}<ArrowDownToLine class="size-3 rotate-180" /> Height
					{:else if fitMode === 'width'}<Columns class="size-3" /> Width
					{:else}<Maximize class="size-3" /> Original
					{/if}
				</button>
				<button onclick={() => showThumbnails = !showThumbnails} class="px-2 py-1 rounded hover:bg-white/10 text-xs {showThumbnails ? 'text-white' : 'text-white/40'}" title="Toggle thumbnails (T)">
					Thumbnails
				</button>
			</div>
		</div>

		<Button variant="ghost" size="sm" class="text-white hover:text-white/80 text-xs" onclick={() => { window.location.href = `/comics/${comicId}`; }}>
			Close
		</Button>
	</div>

	<!-- Main image area -->
	<div class="flex-1 flex items-center justify-center overflow-hidden relative" onclick={() => { nextPage(); saveProgress(); }} role="button" tabindex="0" aria-label={`Page ${currentPage + 1} of ${pageCount}`} onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); nextPage(); } }}>
		{#if pageKeys[currentPage]}
			<img src={pageImages[currentPage]} alt={`Page ${currentPage + 1}`} class={fitClass + ' cursor-pointer select-none'} loading="eager" draggable="false" />
		{:else}
			<div class="bg-gray-800 rounded-lg max-w-3xl aspect-[3/4] flex items-center justify-center text-white/20 text-lg">Page {currentPage + 1}</div>
		{/if}

		<button onclick={(e) => { e.stopPropagation(); prevPage(); saveProgress(); }} disabled={currentPage === 0} class="absolute left-4 top-1/2 -translate-y-1/2 p-2 rounded-full bg-white/5 hover:bg-white/15 text-white/70 disabled:opacity-20 transition-opacity" aria-label="Previous">&larr;</button>
		<button onclick={(e) => { e.stopPropagation(); nextPage(); saveProgress(); }} disabled={currentPage >= pageCount - 1} class="absolute right-4 top-1/2 -translate-y-1/2 p-2 rounded-full bg-white/5 hover:bg-white/15 text-white/70 disabled:opacity-20 transition-opacity" aria-label="Next">&rarr;</button>
	</div>

	<!-- Thumbnail strip -->
	{#if showThumbnails && pageCount > 1}
		<div class="flex-shrink-0 bg-black/90 border-t border-white/10 px-2 py-2 overflow-x-auto">
			<div class="flex gap-1.5 min-w-max">
				{#each Array(pageCount) as _, i}
					<button
						onclick={() => goToPage(i)}
						class="flex-shrink-0 w-12 h-16 rounded overflow-hidden border-2 transition-all {i === currentPage ? 'border-white' : 'border-transparent hover:border-white/30 opacity-60 hover:opacity-100'}"
						aria-label={`Go to page ${i + 1}`}
					>
						<img src={pageImages[i]} alt={`Page ${i + 1}`} class="w-full h-full object-cover" loading="lazy" />
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
