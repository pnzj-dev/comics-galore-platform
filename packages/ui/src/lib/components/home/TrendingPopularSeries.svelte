<!--
  TrendingPopularSeries.svelte

  Full-width, horizontally-scrolling rail of series cards with
  Trending / Popular tabs.

  Usage:
  ```svelte
  <script>
    import TrendingPopularSeries from '$lib/components/home/TrendingPopularSeries.svelte';
    import { goto } from '$app/navigation';
  </script>

  <TrendingPopularSeries
    trending={data.trendingSeries}
    popular={data.popularSeries}
    onViewAll={() => goto('/series')}
  />
  ```
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import SeriesCard from './SeriesCard.svelte';
	import { t } from '$lib/i18n';

	interface SeriesItem {
		id: string;
		slug: string;
		rank: number;
		title: string;
		genre: string;
		cover_url: string;
		overlay_title?: string;
		views_count?: number;
		hearts_count?: number;
	}

	interface Props {
		trending: SeriesItem[];
		popular: SeriesItem[];
		activeTab?: 'trending' | 'popular';
		onTabChange?: (tab: 'trending' | 'popular') => void;
		onViewAll?: () => void;
	}

	let {
		trending = [],
		popular = [],
		activeTab = $bindable('trending'),
		onTabChange,
		onViewAll,
	}: Props = $props();

	const cards = $derived(activeTab === 'trending' ? trending : popular);

	let scrollerEl: HTMLDivElement | undefined = $state();
	let canScrollLeft = $state(false);
	let canScrollRight = $state(false);

	const BATCH_SIZE = 5;
	const GAP = 16;

	function updateScrollState() {
		if (!scrollerEl) return;
		canScrollLeft = scrollerEl.scrollLeft > 1;
		canScrollRight = scrollerEl.scrollLeft < scrollerEl.scrollWidth - scrollerEl.clientWidth - 1;
	}

	function setTab(tab: 'trending' | 'popular') {
		if (tab === activeTab) return;
		activeTab = tab;
		onTabChange?.(tab);
	}

	function onTabKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
			e.preventDefault();
			const next: 'trending' | 'popular' = activeTab === 'trending' ? 'popular' : 'trending';
			setTab(next);
		}
	}

	// Reset scroll position + arrow state whenever the visible card set changes
	// (i.e. the active tab flips). Runs after the DOM has updated.
	$effect(() => {
		cards;
		if (scrollerEl) {
			scrollerEl.scrollLeft = 0;
			updateScrollState();
		}
	});

	onMount(() => {
		updateScrollState();
		window.addEventListener('resize', updateScrollState);
		return () => window.removeEventListener('resize', updateScrollState);
	});

	function batchDistance(): number {
		if (!scrollerEl) return 0;
		const children = Array.from(scrollerEl.children) as HTMLElement[];
		if (children.length === 0) return 0;
		const count = Math.min(BATCH_SIZE, children.length);
		let distance = 0;
		for (let i = 0; i < count; i++) {
			distance += children[i].offsetWidth + GAP;
		}
		return distance;
	}

	function scrollRight() {
		scrollerEl?.scrollBy({ left: batchDistance(), behavior: 'smooth' });
	}

	function scrollLeft() {
		scrollerEl?.scrollBy({ left: -batchDistance(), behavior: 'smooth' });
	}
</script>

<section aria-label={t('series.trendingPopular')} class="py-8">
	<!-- Header -->
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-xl font-semibold">{t('series.trendingPopular')}</h2>
		{#if onViewAll}
			<button
				type="button"
				onclick={onViewAll}
				class="text-sm text-primary hover:underline flex items-center gap-1"
			>
				{t('series.viewAll')}
				<span aria-hidden="true">&gt;</span>
			</button>
		{/if}
	</div>

	<!-- Tabs -->
	<div
		role="tablist"
		aria-label={t('series.trendingPopular')}
		class="flex items-center gap-2 mb-4"
	>
		<button
			type="button"
			role="tab"
			aria-selected={activeTab === 'trending'}
			onclick={() => setTab('trending')}
			onkeydown={onTabKeydown}
			class="px-4 py-1.5 rounded-full text-sm font-medium transition-colors {activeTab === 'trending'
				? 'bg-primary text-primary-foreground'
				: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
		>
			{t('series.trending')}
		</button>
		<button
			type="button"
			role="tab"
			aria-selected={activeTab === 'popular'}
			onclick={() => setTab('popular')}
			onkeydown={onTabKeydown}
			class="px-4 py-1.5 rounded-full text-sm font-medium transition-colors {activeTab === 'popular'
				? 'bg-primary text-primary-foreground'
				: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
		>
			{t('series.popular')}
		</button>
	</div>

	<!-- Scrolling row -->
	<div class="relative group/row">
		<div
			bind:this={scrollerEl}
			onscroll={updateScrollState}
			class="flex gap-4 overflow-x-auto pb-2 snap-x snap-mandatory scroll-smooth no-scrollbar"
		>
			{#each cards as s (s.id)}
				<div class="snap-start">
					<SeriesCard
						id={s.id}
						title={s.title}
						cover_url={s.cover_url}
						genre={s.genre}
						views_count={s.views_count}
						hearts_count={s.hearts_count}
						href="/series/{s.slug}"
						size="md"
					/>
				</div>
			{/each}
		</div>

		<!-- Left arrow -->
		{#if canScrollLeft}
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

		<!-- Right arrow -->
		{#if canScrollRight}
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
</section>

<style>
	.no-scrollbar {
		scrollbar-width: none; /* Firefox */
		-ms-overflow-style: none; /* IE/Edge */
	}
	.no-scrollbar::-webkit-scrollbar {
		display: none; /* Chrome/Safari */
	}
</style>
