<!--
  SeriesCarousel.svelte
  Horizontal scrollable carousel of SeriesCard components.
  Left/right arrow affordances appear based on scroll position.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import SeriesCard from './SeriesCard.svelte';

	interface Series {
		id: string;
		title: string;
		cover_url: string;
		genre?: string;
		views_count?: number;
		hearts_count?: number;
		badge?: string | null;
		overlay_title?: string;
	}

	interface Props {
		series: Series[];
		size?: 'sm' | 'md' | 'lg';
		showArrow?: boolean;
		class?: string;
	}

	let { series = [], size = 'md', showArrow = true, class: className = '' }: Props = $props();

	let scrollEl: HTMLDivElement | undefined = $state();
	let canScrollLeft = $state(false);
	let canScrollRight = $state(false);

	const BATCH_SIZE = 5;
	const GAP = 16;

	function updateScrollState() {
		if (!scrollEl) return;
		canScrollLeft = scrollEl.scrollLeft > 1;
		canScrollRight = scrollEl.scrollLeft < scrollEl.scrollWidth - scrollEl.clientWidth - 1;
	}

	onMount(() => {
		updateScrollState();
		window.addEventListener('resize', updateScrollState);
		return () => window.removeEventListener('resize', updateScrollState);
	});

	function batchDistance(): number {
		if (!scrollEl) return 0;
		const cards = Array.from(scrollEl.children) as HTMLElement[];
		if (cards.length === 0) return 0;
		const count = Math.min(BATCH_SIZE, cards.length);
		let distance = 0;
		for (let i = 0; i < count; i++) {
			distance += cards[i].offsetWidth + GAP;
		}
		return distance;
	}

	function scrollRight() {
		if (!scrollEl) return;
		scrollEl.scrollBy({ left: batchDistance(), behavior: 'smooth' });
	}

	function scrollLeft() {
		if (!scrollEl) return;
		scrollEl.scrollBy({ left: -batchDistance(), behavior: 'smooth' });
	}
</script>

<div class="relative {className}">
	<div
		bind:this={scrollEl}
		onscroll={updateScrollState}
		class="flex gap-4 overflow-x-auto pb-2 scrollbar-none"
		style="scroll-snap-type: x mandatory;"
	>
		{#each series as item (item.id)}
			<div style="scroll-snap-align: start;">
				<SeriesCard
					id={item.id}
					title={item.title}
					cover_url={item.cover_url}
					genre={item.genre}
					views_count={item.views_count}
					hearts_count={item.hearts_count}
					badge={item.badge}
					overlay_title={item.overlay_title}
					{size}
					href="/series/{item.id}"
				/>
			</div>
		{/each}
	</div>

	{#if showArrow && canScrollLeft}
		<button
			type="button"
			onclick={scrollLeft}
			class="absolute -left-3 top-1/3 z-10 hidden h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-background text-muted-foreground shadow-lg ring-1 ring-border transition hover:bg-muted hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-black sm:flex"
			aria-label="Scroll left"
		>
			<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 5l-7 7 7 7" />
			</svg>
		</button>
	{/if}

	{#if showArrow && canScrollRight}
		<button
			type="button"
			onclick={scrollRight}
			class="absolute -right-3 top-1/3 z-10 hidden h-10 w-10 -translate-y-1/2 items-center justify-center rounded-full bg-background text-muted-foreground shadow-lg ring-1 ring-border transition hover:bg-muted hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-black sm:flex"
			aria-label="Scroll right"
		>
			<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
			</svg>
		</button>
	{/if}
</div>

<style>
	.scrollbar-none {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
	.scrollbar-none::-webkit-scrollbar {
		display: none;
	}
</style>
