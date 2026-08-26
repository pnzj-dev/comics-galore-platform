<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import FavoriteStar from '$lib/components/social/FavoriteStar.svelte';
	import { formatDate, formatCompactNumber } from '$lib/utils/format';
	import { ThumbsUp, ThumbsDown, Eye, Download, BookOpen, MessageCircle } from 'lucide-svelte';

	interface Props {
		id: string;
		title: string;
		slug: string;
		description?: string;
		author?: string;
		status?: string;
		cover_key?: string;
		cover_url?: string;
		page_count?: number;
		tags?: string[];
		age_rating?: string;
		view_count?: number;
		download_count?: number;
		like_count?: number;
		dislike_count?: number;
		comment_count?: number;
		is_premium?: boolean;
		published_at?: string;
		created_at?: string;
		is_favorited?: boolean;
		onUnfavorite?: (id: string) => void;
	}

	let {
		id,
		title,
		slug,
		description = '',
		author = '',
		status = '',
		cover_key = '',
		cover_url = '',
		page_count = 0,
		tags = [],
		age_rating = '',
		view_count = 0,
		download_count = 0,
		like_count = 0,
		dislike_count = 0,
		comment_count = 0,
		is_premium = false,
		published_at = '',
		created_at = '',
		is_favorited = false,
		onUnfavorite,
	}: Props = $props();

	const user = $derived($currentUser);
	const authed = $derived(!!user);
	const pageCount = $derived(page_count || 0);
	const visibleTags = $derived(tags.slice(0, 2));
	const overflowTags = $derived(tags.slice(2));
	const isMature = $derived(age_rating === 'mature' || age_rating === 'explicit');

	function coverSrc(): string {
		if (cover_url) return cover_url;
		if (cover_key) return `/media/${cover_key}`;
		return '';
	}
</script>

<Card class="group overflow-hidden flex flex-col hover:shadow-lg transition-all pt-0">
	<a href="/comics/{slug}" class="relative block aspect-[3/4] overflow-hidden bg-muted">
		{#if coverSrc()}
			<img
				src={coverSrc()}
				alt={title}
				onerror={(e) => {
					const img = e.target as HTMLImageElement;
					img.style.display = 'none';
					const fallback = img.nextElementSibling as HTMLElement;
					if (fallback) fallback.style.display = 'flex';
				}}
				class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
				loading="lazy"
			/>
		{/if}
		<div class="w-full h-full items-center justify-center text-muted-foreground {coverSrc() ? 'hidden' : 'flex'}">
			<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
		</div>

		<div class="absolute bottom-0 inset-x-0 h-10 bg-gradient-to-t from-black/30 to-transparent pointer-events-none"></div>

		{#if description}
			<div class="absolute inset-0 bg-black/70 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-4 z-20">
				<p class="text-gray-200 text-xs text-center line-clamp-4">{description}</p>
			</div>
		{/if}

		{#if is_premium}
			<div class="absolute {isMature ? 'top-8' : 'top-2'} left-2 rounded-full bg-yellow-500 px-2 py-0.5 text-xs font-semibold text-white z-10">Premium</div>
		{/if}

		{#if isMature}
			<div class="absolute top-2 left-2 rounded bg-red-600 px-1.5 py-0.5 text-[10px] font-semibold text-white z-10 uppercase">{age_rating}</div>
		{/if}

		{#if authed}
			<FavoriteStar
				comicId={id}
				initialFavorited={is_favorited}
				onUnfavorite={onUnfavorite}
				class="absolute top-2 right-2 z-30"
			/>
		{/if}
	</a>

	<CardHeader class="p-3 pb-1 space-y-1.5">
		{#if author}
			<span class="text-[10px] font-bold uppercase tracking-wide text-purple-500">{author}</span>
		{/if}
		<a href="/comics/{slug}">
			<CardTitle class="text-sm font-semibold leading-snug h-[2.5rem] line-clamp-2 hover:text-purple-500 transition-colors">{title}</CardTitle>
		</a>
		<div class="flex items-center gap-1 flex-wrap">
			{#each visibleTags as tag}
				<Badge variant="secondary" class="text-[10px] px-1.5 py-0.5 rounded">{tag}</Badge>
			{/each}
			{#if overflowTags.length > 0}
				<span title={overflowTags.join(', ')} class="inline-flex items-center justify-center rounded-4xl border border-border px-2 py-0.5 text-[10px] font-medium text-foreground cursor-default">+{overflowTags.length}</span>
			{/if}
		</div>
	</CardHeader>

	<CardContent class="px-3 pb-0">
		<p class="text-xs text-muted-foreground">{formatDate(published_at || created_at)}</p>
	</CardContent>

	<CardFooter class="px-3 pb-3 pt-2 mt-auto flex items-center justify-between border-t text-xs text-muted-foreground">
		<span class="flex items-center gap-1"><Eye class="size-3.5" />{formatCompactNumber(view_count)}</span>
		<span class="flex items-center gap-1"><Download class="size-3.5" />{formatCompactNumber(download_count)}</span>
		<span class="flex items-center gap-1"><MessageCircle class="size-3.5" />{formatCompactNumber(comment_count)}</span>
		<span class="flex items-center gap-1"><BookOpen class="size-3.5" />{formatCompactNumber(pageCount)}</span>
		<span class="flex items-center gap-1"><ThumbsUp class="size-3.5" />{formatCompactNumber(like_count)}</span>
		<span class="flex items-center gap-1"><ThumbsDown class="size-3.5" />{formatCompactNumber(dislike_count)}</span>
	</CardFooter>
</Card>
