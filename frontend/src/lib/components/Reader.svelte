<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';

	interface Props {
		comicId: string;
		pageKeys: string[];
		totalPages: number;
	}

	let { comicId, pageKeys, totalPages }: Props = $props();

	let currentPage = $state(0);
	let savedProgress = $state(false);
	let autosaveTimer: ReturnType<typeof setTimeout>;

	const pageCount = $derived(totalPages || pageKeys.length || 1);

	function getPageLabel(page: number, total: number): string {
		return `Page ${page + 1} of ${total}`;
	}

	function nextPage() {
		if (currentPage < pageCount - 1) currentPage++;
	}

	function prevPage() {
		if (currentPage > 0) currentPage--;
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
			} catch { /* ignore */ }
		}, 1000);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowRight' || e.key === ' ') { e.preventDefault(); nextPage(); saveProgress(); }
		if (e.key === 'ArrowLeft') { e.preventDefault(); prevPage(); saveProgress(); }
	}

	$effect(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});

	onMount(async () => {
		try {
			const progress = await api.get<{ current_page: number }>(`/reading/${comicId}`);
			currentPage = progress.current_page || 0;
		} catch { /* no progress yet */ }
	});
</script>

{#if currentPage >= 0}
	<div class="fixed inset-0 z-50 bg-black flex flex-col">
		<div class="flex items-center justify-between p-2 bg-black/80 text-white">
			<div class="flex items-center gap-2">
				<span class="text-sm">Page {currentPage + 1} / {pageCount}</span>
				{#if savedProgress}
					<span class="text-xs text-green-400">Saved</span>
				{/if}
			</div>
			<Button variant="ghost" size="sm" class="text-white" onclick={() => { window.location.href = `/comics/${comicId}`; }}>
				Close Reader
			</Button>
		</div>

		<div class="flex-1 flex items-center justify-center overflow-hidden bg-black" onclick={nextPage} role="button" tabindex="0" aria-label={getPageLabel(currentPage, pageCount)} onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); nextPage(); } }}>
			<div class="max-h-full max-w-full p-2">
				{#if pageKeys[currentPage]}
					<img
						src={`/media/${pageKeys[currentPage]}`}
						alt={`Page ${currentPage + 1} of ${pageCount}`}
						class="max-h-[85vh] max-w-full object-contain rounded shadow-lg"
						loading="eager"
					/>
				{:else}
					<div class="bg-gray-800 rounded-lg w-full max-w-3xl aspect-[3/4] flex items-center justify-center text-white/40 text-lg">
						Page {currentPage + 1}
					</div>
				{/if}
			</div>
		</div>

		<div class="flex justify-center gap-4 p-4 bg-black/80">
			<Button variant="outline" size="lg" onclick={() => { prevPage(); saveProgress(); }} disabled={currentPage === 0} class="text-white border-white/20">
				Previous
			</Button>
			<Button variant="outline" size="lg" onclick={() => { nextPage(); saveProgress(); }} disabled={currentPage >= pageCount - 1} class="text-white border-white/20">
				Next
			</Button>
		</div>
	</div>
{/if}
