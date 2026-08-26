<script lang="ts">
	import { modal } from '$lib/stores/modal.svelte';

	interface Props {
		images: string[];
		startIndex?: number;
		onClose?: () => void;
	}

	let { images, startIndex = 0, onClose }: Props = $props();

	const open = $derived(modal.isOpen('lightbox'));
	let currentIndex = $state(0);

	$effect(() => {
		if (open) currentIndex = startIndex;
	});

	function close() {
		modal.close('lightbox');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!open) return;
		if (e.key === 'Escape') { close(); return; }
		if (e.key === 'ArrowLeft') { prev(); return; }
		if (e.key === 'ArrowRight') { next(); return; }
	}

	function prev() { if (currentIndex > 0) currentIndex--; }
	function next() { if (currentIndex < images.length - 1) currentIndex++; }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/95 flex flex-col" onclick={close} onkeydown={(e) => { if (e.key === 'Escape') close(); }} role="dialog" tabindex="-1" aria-label="Image lightbox">
		<div class="flex items-center justify-between p-4 text-white/80 text-sm flex-shrink-0" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="presentation">
			<span>{currentIndex + 1} / {images.length}</span>
			<div class="flex gap-1.5 mx-4">
				{#each images as _, i}
					<button onclick={() => currentIndex = i} class="w-2 h-2 rounded-full transition-colors {i === currentIndex ? 'bg-white' : 'bg-white/30 hover:bg-white/50'}" aria-label={`Image ${i + 1}`}></button>
				{/each}
			</div>
			<button onclick={close} class="hover:text-white" aria-label="Close">✕</button>
		</div>
		<div class="flex-1 flex items-center justify-center overflow-hidden relative" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="presentation">
			<button onclick={prev} disabled={currentIndex === 0} class="absolute left-4 top-1/2 -translate-y-1/2 z-10 p-2 rounded-full bg-white/10 hover:bg-white/20 text-white disabled:opacity-30" aria-label="Previous">&larr;</button>
			<button onclick={() => currentIndex < images.length - 1 && currentIndex++} class="p-0 border-0 bg-transparent" aria-label="Next image">
				<img src={images[currentIndex]} alt={`Image ${currentIndex + 1}`} class="max-h-[85vh] max-w-[90vw] object-contain select-none cursor-pointer" draggable="false" />
			</button>
			<button onclick={next} disabled={currentIndex >= images.length - 1} class="absolute right-4 top-1/2 -translate-y-1/2 z-10 p-2 rounded-full bg-white/10 hover:bg-white/20 text-white disabled:opacity-30" aria-label="Next">&rarr;</button>
		</div>
	</div>
{/if}
