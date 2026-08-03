<script lang="ts">
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';

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
		comment_count?: number;
		is_premium?: boolean;
		published_at?: string;
		created_at?: string;
		initial_favorited?: boolean;
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
		page_keys = [],
		tags = [],
		age_rating = '',
		view_count = 0,
		download_count = 0,
		like_count = 0,
		fav_count = 0,
		comment_count = 0,
		is_premium = false,
		published_at = '',
		created_at = '',
		initial_favorited = false
	}: Props = $props();

	const user = $derived($currentUser);
	const authed = $derived(!!user);
	const pageCount = $derived(page_keys?.length || 0);

	// svelte-ignore state_referenced_locally
	let favorited = $state(initial_favorited);
	// svelte-ignore state_referenced_locally
	let favCount = $state(fav_count);

	async function toggleFavorite(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		const next = !favorited;
		favorited = next;
		favCount += next ? 1 : -1;
		try {
			if (next) {
				const res = await api.post<{ favorited: boolean; fav_count: number }>(`/comics/${id}/favorite`);
				favorited = res.favorited;
				favCount = res.fav_count;
			} else {
				await api.delete(`/comics/${id}/favorite`);
			}
		} catch {
			favorited = !next;
			favCount += next ? -1 : 1;
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

	const visibleTags = $derived(tags.slice(0, 2));
	const overflowTags = $derived(tags.slice(2));
	const hasOverflow = $derived(overflowTags.length > 0);

	function coverSrc(): string {
		if (cover_url) return cover_url;
		if (cover_key) return `/media/${cover_key}`;
		return '';
	}
</script>

<div class="group rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden hover:border-purple-300 dark:hover:border-purple-700 transition-all hover:shadow-lg flex flex-col">
	<a href="/comics/{id}" class="block relative">
		<div class="aspect-[3/4] bg-gray-100 dark:bg-gray-700 relative overflow-hidden">
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
			<div class="w-full h-full items-center justify-center text-gray-400 dark:text-gray-500 {coverSrc() ? 'hidden' : 'flex'}">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="size-12"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
			</div>

			<div class="absolute bottom-0 inset-x-0 h-10 bg-gradient-to-t from-black/30 to-transparent pointer-events-none"></div>

			{#if description}
				<div class="absolute inset-0 bg-black/70 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-4 z-20">
					<p class="text-gray-200 text-xs text-center line-clamp-4">{description}</p>
				</div>
			{/if}

			{#if is_premium}
				<div class="absolute top-2 right-2 rounded-full bg-yellow-500 px-2 py-0.5 text-xs font-semibold text-white z-10">Premium</div>
			{/if}

			{#if age_rating === 'mature' || age_rating === 'explicit'}
				<div class="absolute top-2 left-2 rounded bg-red-600 px-1.5 py-0.5 text-[10px] font-semibold text-white z-10 uppercase">{age_rating}</div>
			{/if}
		</div>
	</a>

	<div class="p-3 pb-1 flex-1">
		{#if author}
			<span class="text-xs font-bold uppercase tracking-wide text-purple-500">{author}</span>
		{/if}

		<a href="/comics/{id}" class="block">
			<h3 class="text-sm font-semibold leading-snug line-clamp-2 min-h-[2.5rem] hover:text-purple-500 transition-colors">{title}</h3>
		</a>

		<div class="flex items-center gap-1 mt-1 min-h-0 flex-wrap">
			{#each visibleTags as tag}
				<span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">{tag}</span>
			{/each}
			{#if hasOverflow}
				<span title={overflowTags.join('\n')} class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-purple-500 dark:text-purple-400 cursor-default">+{overflowTags.length}</span>
			{/if}
		</div>

		<div class="flex items-center justify-between mt-1.5">
			<span class="text-xs text-gray-500">{formatDate(published_at || created_at)}</span>
			{#if authed}
				<button
					onclick={toggleFavorite}
					aria-label="Toggle favorite"
					class="flex-shrink-0 {favorited ? 'text-yellow-500 hover:text-yellow-600' : 'text-gray-400 hover:text-yellow-500'}"
				>
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill={favorited ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
				</button>
			{/if}
		</div>
	</div>

	<div class="px-3 pb-3 pt-2 mt-auto border-t border-gray-100 dark:border-gray-700/50">
		<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
			<div class="flex items-center gap-3">
				<span class="flex items-center gap-1">
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/></svg>
					<span>{compactNum(view_count)}</span>
				</span>
				<span class="flex items-center gap-1">
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
					<span>{compactNum(download_count)}</span>
				</span>
				<span class="flex items-center gap-1">
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/></svg>
					<span>{compactNum(comment_count)}</span>
				</span>
				<span class="flex items-center gap-1">
					<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H19a1 1 0 0 1 1 1v18a1 1 0 0 1-1 1H6.5A1.5 1.5 0 0 1 5 20.5a1.5 1.5 0 0 1 1.5-1.5H20"/></svg>
					<span>{pageCount}</span>
				</span>
			</div>
		</div>
	</div>
</div>
