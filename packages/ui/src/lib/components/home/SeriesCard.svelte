<!--
  SeriesCard.svelte
  Flexible card used in all carousels.
  Variants controlled by props (showViews, showHearts, showBadge, size).
-->
<script lang="ts">
	import { formatCompactNumber } from '$lib/utils/format';
	import { Card, CardContent } from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';

	interface Props {
		id: string;
		title: string;
		cover_url: string;
		genre?: string;
		views_count?: number;
		hearts_count?: number;
		badge?: string | null;
		overlay_title?: string;
		size?: 'sm' | 'md' | 'lg';
		href?: string;
		class?: string;
	}

	let {
		id,
		title,
		cover_url,
		genre = '',
		views_count = 0,
		hearts_count = 0,
		badge = null,
		overlay_title = '',
		size = 'md',
		href = '#',
		class: className = ''
	}: Props = $props();

	const sizeClasses = {
		sm: 'w-32 sm:w-36',
		md: 'w-40 sm:w-44',
		lg: 'w-44 sm:w-52'
	};

	const viewsLabel = $derived(views_count > 0 ? `${formatCompactNumber(views_count)} Views` : '');
	const heartsLabel = $derived(hearts_count > 0 ? formatCompactNumber(hearts_count) : '');
</script>

<Card size="sm" class="group flex shrink-0 flex-col pt-0 transition hover:shadow-md {sizeClasses[size]} {className}">
	<a href={href} class="relative block aspect-[3/4] overflow-hidden bg-muted" aria-label={title}>
		{#if cover_url}
			<img
				src={cover_url}
				alt={title}
				class="h-full w-full object-cover transition duration-300 group-hover:scale-105"
				loading="lazy"
			/>
		{/if}

		{#if badge}
			<Badge variant="secondary" class="absolute left-2 top-2">{badge}</Badge>
		{/if}
	</a>

	<CardContent class="flex flex-col gap-1 p-3">
		<h3 class="truncate text-sm font-medium text-foreground">{title}</h3>

		{#if genre}
			<p class="truncate text-xs text-muted-foreground">{genre}</p>
		{/if}

		{#if viewsLabel || heartsLabel}
			<div class="flex items-center gap-x-2 text-xs text-muted-foreground">
				{#if viewsLabel}<span>{viewsLabel}</span>{/if}
				{#if heartsLabel}
					<span class="inline-flex items-center gap-0.5">
						<svg class="h-3 w-3 text-muted-foreground" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"
								clip-rule="evenodd"
							/>
						</svg>
						{heartsLabel}
					</span>
				{/if}
			</div>
		{/if}
	</CardContent>
</Card>
