<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { currentUser } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
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
		page_keys?: string[];
		tags?: string[];
		age_rating?: string;
		view_count?: number;
		download_count?: number;
		like_count?: number;
		fav_count?: number;
		dislike_count?: number;
		comment_count?: number;
		is_premium?: boolean;
		published_at?: string;
		created_at?: string;
		is_favorited?: boolean;
	}

	let {
		id,
		title,
		slug,
		description = '',
		author = '',
		cover_key = '',
		cover_url = '',
		page_keys = [],
		tags = [],
		age_rating = '',
		view_count = 0,
		download_count = 0,
		like_count = 0,
		fav_count = 0,
		dislike_count = 0,
		comment_count = 0,
		is_premium = false,
		published_at = '',
		is_favorited = false
	}: Props = $props();

	const user = $derived($currentUser);
	const authed = $derived(!!user);
	const pageCount = $derived(page_keys?.length || 0);

	let favorited = $state(is_favorited);
	let favCount = $state(fav_count);
	let liking = $state(false);
	let favHovered = $state(false);

	async function toggleFavorite(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (liking) return;
		const next = !favorited;
		favorited = next;
		favCount += next ? 1 : -1;
		liking = true;
		try {
			const res = await encore.comics.ToggleFavorite(id);
			favorited = res.favorited;
			favCount = res.fav_count;
		} catch {
			favorited = !next;
			favCount += next ? -1 : 1;
		} finally {
			liking = false;
		}
	}

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function compactNum(n: number): string {
		if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
		if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k';
		return String(n);
	}

	const visibleTags = $derived(tags.slice(0, 3));
	const overflowTags = $derived(tags.slice(3));

	function coverSrc(): string {
		if (cover_url) return cover_url;
		if (cover_key) return '/media/${cover_key}';
		return '';
	}
</script>

<a href="/comics/{slug}" class="group block border border-border rounded-xl overflow-hidden hover:shadow-lg transition-all">
	<div class="flex flex-col sm:flex-row">
		<div class="relative sm:w-[35%] aspect-[3/4] sm:aspect-auto sm:min-h-[320px] overflow-hidden bg-muted">
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

			<div class="absolute bottom-0 inset-x-0 h-10 bg-gradient-to-t from-black/30 to-transparent pointer-events-none sm:hidden"></div>

			{#if is_premium}
				<div class="absolute top-8 left-2 rounded-full bg-yellow-500 px-2 py-0.5 text-xs font-semibold text-white z-10">Premium</div>
			{/if}

			{#if age_rating === 'mature' || age_rating === 'explicit'}
				<div class="absolute top-2 left-2 rounded bg-red-600 px-1.5 py-0.5 text-[10px] font-semibold text-white z-10 uppercase">{age_rating}</div>
			{/if}
		</div>

		<div class="flex-1 flex flex-col p-5">
			<div class="flex items-start justify-between gap-3">
				<div class="flex-1 min-w-0">
					{#if is_premium}
						<span class="hidden sm:inline-block rounded-full bg-yellow-500 px-2 py-0.5 text-xs font-semibold text-white mb-1">Premium</span>
					{/if}
					{#if author}
						<p class="text-[10px] font-bold uppercase tracking-wide text-purple-500">{author}</p>
					{/if}
					<h2 class="text-lg font-bold leading-tight mt-0.5 group-hover:text-purple-500 transition-colors">{title}</h2>
				</div>

				{#if authed}
					{@const favHighlight = favorited !== favHovered}
					<Button
						variant="ghost"
						size="icon"
						onclick={toggleFavorite}
						onmouseenter={() => favHovered = true}
						onmouseleave={() => favHovered = false}
						disabled={liking}
						aria-label={favorited ? 'Remove from favorites' : 'Add to favorites'}
						class="shrink-0 size-8 rounded-full {favHighlight ? 'text-yellow-400' : 'text-muted-foreground hover:text-yellow-400'} {liking ? 'opacity-60' : ''}"
					>
						<svg class="size-4" viewBox="0 0 24 24" fill={favHighlight ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
					</Button>
				{/if}
			</div>

			<p class="text-xs text-muted-foreground mt-1">{formatDate(published_at)}</p>

			{#if description}
				<p class="text-sm text-muted-foreground mt-3 leading-relaxed line-clamp-3">{description}</p>
			{/if}

			{#if tags.length > 0}
				<div class="flex items-center gap-1 flex-wrap mt-3">
					{#each visibleTags as tag}
						<Badge variant="secondary" class="text-[10px] px-1.5 py-0.5 rounded">{tag}</Badge>
					{/each}
					{#if overflowTags.length > 0}
						<span title={overflowTags.join(', ')} class="inline-flex items-center justify-center rounded-4xl border border-border px-2 py-0.5 text-[10px] font-medium text-foreground cursor-default">+{overflowTags.length}</span>
					{/if}
				</div>
			{/if}

			<div class="mt-auto pt-4 flex items-center justify-between border-t text-xs text-muted-foreground">
				<div class="flex items-center gap-3">
					<span class="flex items-center gap-1"><Eye class="size-3.5" />{compactNum(view_count)}</span>
					<span class="flex items-center gap-1"><Download class="size-3.5" />{compactNum(download_count)}</span>
					<span class="flex items-center gap-1"><MessageCircle class="size-3.5" />{compactNum(comment_count)}</span>
					<span class="flex items-center gap-1"><BookOpen class="size-3.5" />{pageCount}</span>
				</div>
				<div class="flex items-center gap-3">
					<span class="flex items-center gap-1"><ThumbsUp class="size-3.5" />{compactNum(like_count)}</span>
					<span class="flex items-center gap-1"><ThumbsDown class="size-3.5" />{compactNum(dislike_count)}</span>
				</div>
			</div>
		</div>
	</div>
</a>
