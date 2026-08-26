<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import FavoriteStar from '$lib/components/social/FavoriteStar.svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { formatDate } from '$lib/utils/format';

	interface Props {
		id: string;
		title: string;
		slug?: string;
		author?: string;
		cover_key?: string;
		cover_url?: string;
		tags?: string[];
		age_rating?: string;
		is_premium?: boolean;
		published_at?: string;
		created_at?: string;
		is_favorited?: boolean;
	}

	let {
		id,
		title,
		slug = '',
		author = '',
		cover_key = '',
		cover_url = '',
		tags = [],
		age_rating = '',
		is_premium = false,
		published_at = '',
		created_at = '',
		is_favorited = false
	}: Props = $props();

	const user = $derived($currentUser);
	const authed = $derived(!!user);

	const visibleTags = $derived(tags.slice(0, 2));
	const overflowTags = $derived(tags.slice(2));

	function coverSrc(): string {
		if (cover_url) return cover_url;
		if (cover_key) return '';
		return '';
	}
</script>

<div class="rounded-xl overflow-hidden border hover:border-purple-300 dark:hover:border-purple-700 transition-all hover:shadow-sm flex flex-col bg-card">
	<a href="/comics/{slug}" class="block">
		<div class="p-2 pb-0 space-y-0.5">
			<p class="text-xs font-semibold line-clamp-1 hover:text-purple-500 transition-colors">{title}</p>
			{#if author}
				<p class="text-[10px] font-bold uppercase tracking-wide text-purple-500">{author}</p>
			{/if}
			<p class="text-[10px] text-muted-foreground">{formatDate(published_at || created_at)}</p>
		</div>

		<div class="p-2 pt-1">
			<div class="aspect-[3/4] rounded-lg overflow-hidden bg-muted relative group/cmp">
				{#if coverSrc()}
					<img src={coverSrc()} alt={title} class="w-full h-full object-cover" loading="lazy"
						onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
				{/if}
				<div class="w-full h-full items-center justify-center text-muted-foreground {coverSrc() ? 'hidden' : 'flex'}">
					<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
				</div>

				{#if is_premium}
					<div class="absolute top-1.5 left-1.5 rounded-full bg-yellow-500 px-1.5 py-0.5 text-[10px] font-semibold text-white">Premium</div>
				{/if}

				{#if age_rating === 'mature' || age_rating === 'explicit'}
					<div class="absolute top-1.5 left-1.5 rounded bg-red-600 px-1 py-0.5 text-[10px] font-semibold text-white uppercase {is_premium ? 'top-7' : ''}">{age_rating}</div>
				{/if}

				{#if authed}
					<FavoriteStar
						comicId={id}
						initialFavorited={is_favorited}
						size="sm"
						class="absolute top-1.5 right-1.5"
					/>
				{/if}
			</div>
		</div>
	</a>

	{#if visibleTags.length > 0}
		<div class="px-2 pb-2 flex items-center gap-1 flex-wrap">
			{#each visibleTags as tag}
				<Badge variant="secondary" class="text-[10px] px-1 py-0 h-4 rounded">{tag}</Badge>
			{/each}
			{#if overflowTags.length > 0}
				<span title={overflowTags.join(', ')} class="inline-flex items-center justify-center rounded-4xl border border-border px-1.5 py-0 h-4 text-[10px] font-medium text-foreground cursor-default">+{overflowTags.length}</span>
			{/if}
		</div>
	{/if}
</div>
