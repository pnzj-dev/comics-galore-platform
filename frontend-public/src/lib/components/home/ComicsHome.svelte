<!--
  ComicsHome.svelte
  Main homepage layout that composes the advertising, category, newly released,
  daily, and indie sections.

  Usage example:
    <ComicsHome
      ad={{ title: "Your story belongs here", cta_text: "Learn more" }}
      categories={['Drama', 'Fantasy', 'Comedy']}
      popular_by_category={[...]}
      newly_released={[...]}
      daily_series={[...]}
      indie_series={[...]}
      activeCategory="Drama"
      activeDay="fri"
      onCategoryChange={(id) => ...}
      onDayChange={(day) => ...}
      onViewAll={(section) => ...}
    />
-->
<script lang="ts">
	import AdBanner from './AdBanner.svelte';
	import CategoryPills from './CategoryPills.svelte';
	import DayPills from './DayPills.svelte';
	import SeriesCarousel from './SeriesCarousel.svelte';

	export interface Series {
		id: string;
		slug: string;
		title: string;
		cover_url: string;
		genre?: string;
		views_count?: number;
		hearts_count?: number;
		badge?: string | null;
		overlay_title?: string;
		schedule_day?: string;
	}

	export interface AdBannerData {
		image_url?: string;
		title?: string;
		subtitle?: string;
		cta_text?: string;
		cta_href?: string;
	}

	interface Props {
		ad?: AdBannerData;
		categories?: string[];
		popular_by_category?: Series[];
		newly_released?: Series[];
		daily_series?: Series[];
		indie_series?: Series[];
		activeCategory?: string;
		activeDay?: string;
		onCategoryChange?: (id: string) => void;
		onDayChange?: (day: string) => void;
		onViewAll?: (section: string) => void;
	}

	let {
		ad = {},
		categories = [],
		popular_by_category = [],
		newly_released = [],
		daily_series = [],
		indie_series = [],
		activeCategory = '',
		activeDay = 'mon',
		onCategoryChange,
		onDayChange,
		onViewAll,
	}: Props = $props();

	const days = [
		{ id: 'mon', name: 'Mon' },
		{ id: 'tue', name: 'Tue' },
		{ id: 'wed', name: 'Wed' },
		{ id: 'thu', name: 'Thu' },
		{ id: 'fri', name: 'Fri' },
		{ id: 'sat', name: 'Sat' },
		{ id: 'sun', name: 'Sun' },
		{ id: 'completed', name: 'Completed' },
	];

	// Client-side filtering of the already-loaded daily list by schedule day.
	const filteredDaily = $derived(
		activeDay === 'completed'
			? daily_series.filter((s) => s.schedule_day === 'completed')
			: daily_series.filter((s) => s.schedule_day === activeDay),
	);

	function handleCategoryChange(id: string) {
		onCategoryChange?.(id);
	}

	function handleDayChange(day: string) {
		onDayChange?.(day);
	}
</script>

<div class="space-y-10 py-6">
	<!-- 1. Advertising Section -->
	<AdBanner
		imageUrl={ad.image_url}
		title={ad.title}
		subtitle={ad.subtitle}
		ctaText={ad.cta_text}
		ctaHref={ad.cta_href}
	/>

	<!-- 2. Popular Series by Category -->
	<section aria-labelledby="popular-heading">
		<div class="mb-4 flex items-center justify-between">
			<h2 id="popular-heading" class="text-xl font-bold text-gray-900 dark:text-white">
				Popular Series by Category
			</h2>
			{#if onViewAll}
				<button
					type="button"
					onclick={() => onViewAll('popular')}
					class="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
				>
					View all &gt;
				</button>
			{/if}
		</div>

		<CategoryPills {categories} activeId={activeCategory} onChange={handleCategoryChange} class="mb-5" />

		<SeriesCarousel series={popular_by_category} size="md" />
	</section>

	<!-- 3. Newly Released -->
	<section aria-labelledby="newly-heading">
		<div class="mb-4 flex items-center justify-between">
			<h2 id="newly-heading" class="text-xl font-bold text-gray-900 dark:text-white">
				Newly Released
			</h2>
			{#if onViewAll}
				<button
					type="button"
					onclick={() => onViewAll('newly')}
					class="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
				>
					View all &gt;
				</button>
			{/if}
		</div>

		<SeriesCarousel series={newly_released} size="lg" />
	</section>

	<!-- 4. Daily -->
	<section aria-labelledby="daily-heading">
		<div class="mb-4 flex items-center justify-between">
			<h2 id="daily-heading" class="text-xl font-bold text-gray-900 dark:text-white">Daily</h2>
			{#if onViewAll}
				<button
					type="button"
					onclick={() => onViewAll('daily')}
					class="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
				>
					View all &gt;
				</button>
			{/if}
		</div>

		<DayPills {days} activeId={activeDay} onChange={handleDayChange} class="mb-5" />

		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
			{#each filteredDaily as item (item.id)}
				<a href="/series/{item.slug}" class="group flex flex-col gap-2" aria-label={item.title}>
					<div class="relative aspect-[3/4] overflow-hidden rounded-xl bg-muted transition group-hover:shadow-md">
						<img
							src={item.cover_url}
							alt={item.title}
							class="h-full w-full object-cover transition duration-300 group-hover:scale-105"
							loading="lazy"
						/>
						{#if item.badge}
							<span
								class="absolute left-2 top-2 rounded bg-secondary px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-secondary-foreground"
							>
								{item.badge}
							</span>
						{/if}
					</div>
					<div class="min-w-0">
						{#if item.genre}
							<p class="text-xs text-muted-foreground">{item.genre}</p>
						{/if}
						<h3 class="truncate text-sm font-medium text-foreground">{item.title}</h3>
					</div>
				</a>
			{/each}
		</div>
	</section>

	<!-- 5. More stories from indie creators -->
	<section aria-labelledby="indie-heading">
		<div class="mb-4 flex items-center justify-between">
			<h2 id="indie-heading" class="text-xl font-bold text-gray-900 dark:text-white">
				More stories from indie creators
			</h2>
			{#if onViewAll}
				<button
					type="button"
					onclick={() => onViewAll('indie')}
					class="text-sm font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
				>
					View all &gt;
				</button>
			{/if}
		</div>

		<SeriesCarousel series={indie_series} size="sm" />
	</section>
</div>
