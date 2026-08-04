<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LikeButton from '$lib/components/LikeButton.svelte';
	import FavoriteButton from '$lib/components/FavoriteButton.svelte';
	import Reader from '$lib/components/Reader.svelte';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import { currentUser } from '$lib/stores/auth';
	import { Eye, Download, BookOpen, Globe } from 'lucide-svelte';

	let comic = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let reading = $state(false);
	let likeStatus = $state<{ liked: boolean; favorited: boolean } | null>(null);
	let related = $state<any[]>([]);
	let relatedLoading = $state(true);

	const id = $derived($page.params.id);
	const user = $derived($currentUser);

	function compactNum(n: number): string {
		if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
		if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k';
		return String(n);
	}

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	onMount(async () => {
		try {
			comic = await api.get(`/comics/${id}`);
			if (user) {
				try { likeStatus = await api.get(`/comics/${id}/like-status`); } catch {}
			}
			loadRelated();
		} catch (err) {
			error = (err as Error).message;
		}
		loading = false;
	});

	async function loadRelated() {
		try {
			const res = await api.get<{ comics: any[] }>('/comics?limit=4');
			related = res.comics.filter((c: any) => c.id !== id).slice(0, 4);
		} catch {}
		relatedLoading = false;
	}
</script>

<svelte:head>
	<title>{comic?.title || 'Comic'} - Comics Galore</title>
</svelte:head>

{#if reading && comic}
	<Reader comicId={comic.id} pageKeys={comic.page_keys || []} totalPages={(comic.page_keys || []).length || 1} />
{/if}

{#if loading}
	<div class="max-w-4xl mx-auto py-8">
		<div class="grid md:grid-cols-[320px_1fr] gap-8 animate-pulse">
			<div class="aspect-[3/4] rounded-xl bg-gray-200 dark:bg-gray-700"></div>
			<div class="space-y-4">
				<div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-2/3"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-5/6"></div>
			</div>
		</div>
	</div>
{:else if error}
	<p class="text-destructive py-20 text-center">{error}</p>
{:else if comic}
	<div class="max-w-4xl mx-auto py-8">
		<div class="grid md:grid-cols-[320px_1fr] gap-8">
			<!-- Cover + Thumbnail Gallery -->
			<div class="space-y-3">
				<div class="aspect-[3/4] rounded-xl bg-muted overflow-hidden relative">
					<div class="w-full h-full flex items-center justify-center text-muted-foreground">
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="size-16"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
					</div>
					{#if comic.status === 'published'}
						<div class="absolute bottom-2 right-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{comic.page_keys?.length || 0} pages</div>
					{/if}
				</div>
				{#if comic.is_premium}
					<div class="text-center">
						<span class="inline-flex items-center gap-1 text-xs font-semibold text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900/30 px-3 py-1 rounded-full">Premium</span>
					</div>
				{/if}
			</div>

			<!-- Info Column -->
			<div>
				<h1 class="text-3xl font-bold">{comic.title}</h1>

				<div class="flex flex-wrap items-center gap-2 mt-2">
					<span class="text-xs rounded-full px-2 py-0.5 font-medium {comic.status === 'published' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : comic.status === 'pending_review' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'}">
						{comic.status?.replace('_', ' ')}
					</span>
					<span class="text-xs bg-muted px-2 py-0.5 rounded-full">{comic.age_rating?.replace('_', ' ')}</span>
				</div>

				<!-- Author -->
				<div class="flex items-center gap-2 mt-3">
					<div class="size-8 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-xs font-medium text-purple-600 dark:text-purple-300">
						{comic.title?.charAt(0).toUpperCase() || '?'}
					</div>
					<div>
						<p class="text-sm font-medium">Author</p>
					</div>
				</div>

				<p class="mt-4 text-muted-foreground text-sm">{comic.description}</p>

				<!-- Reactions -->
				<div class="flex items-center gap-4 mt-4 py-2 border-y border-border">
					<LikeButton comicId={comic.id} initialLiked={likeStatus?.liked} initialCount={comic.like_count} />
					<FavoriteButton comicId={comic.id} initialFavorited={likeStatus?.favorited} initialCount={comic.fav_count} />
				</div>

				<!-- Metadata Panel -->
				<div class="grid grid-cols-2 gap-x-6 gap-y-2 mt-4 text-sm">
					<div class="flex items-center gap-2">
						<Eye class="size-3.5 text-muted-foreground" />
						<span class="text-muted-foreground">Views</span>
						<span class="ml-auto font-medium">{compactNum(comic.view_count)}</span>
					</div>
					<div class="flex items-center gap-2">
						<Download class="size-3.5 text-muted-foreground" />
						<span class="text-muted-foreground">Downloads</span>
						<span class="ml-auto font-medium">{compactNum(comic.download_count)}</span>
					</div>
					<div class="flex items-center gap-2">
						<BookOpen class="size-3.5 text-muted-foreground" />
						<span class="text-muted-foreground">Pages</span>
						<span class="ml-auto font-medium">{comic.page_keys?.length || 0}</span>
					</div>
					<div class="flex items-center gap-2">
						<span class="text-muted-foreground">Published</span>
						<span class="ml-auto font-medium text-xs">{formatDate(comic.published_at) || '-'}</span>
					</div>
					<div class="flex items-center gap-2">
						<span class="text-muted-foreground">Reading</span>
						<span class="ml-auto font-medium">5 min</span>
					</div>
					<div class="flex items-center gap-2">
						<Globe class="size-3.5 text-muted-foreground" />
						<span class="text-muted-foreground">Language</span>
						<span class="ml-auto font-medium uppercase">{comic.content_language}</span>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="flex gap-2 mt-6">
					{#if comic.status === 'published'}
						<Button size="lg" class="bg-emerald-600 hover:bg-emerald-700" onclick={() => reading = true}>Start Reading</Button>
					{/if}
					{#if user}
						<Button size="lg" variant="outline">Download</Button>
					{:else}
						<Button size="lg" variant="outline" href="/login">Sign in to Download</Button>
					{/if}
				</div>

				<!-- Tags -->
				{#if (comic.tags || []).length > 0}
					<div class="flex flex-wrap gap-1 mt-4">
						{#each comic.tags as tag}
							<span class="text-[10px] px-2 py-0.5 rounded-full bg-primary/10 text-primary">{tag}</span>
						{/each}
					</div>
				{/if}

				{#if comic.rejection_reason}
					<div class="mt-4 p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
						Rejection reason: {comic.rejection_reason}
					</div>
				{/if}
			</div>
		</div>

		<!-- Related Comics -->
		<section class="mt-12">
			<h2 class="text-xl font-semibold mb-4">Related Comics</h2>
			{#if relatedLoading}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					{#each Array(4) as _}
						<div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden animate-pulse">
							<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
							<div class="p-3 space-y-2">
								<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
								<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
							</div>
						</div>
					{/each}
				</div>
			{:else if related.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					{#each related as r}
						<ComicCard {...r} />
					{/each}
				</div>
			{:else}
				<p class="text-sm text-muted-foreground text-center py-8">No related comics found.</p>
			{/if}
		</section>
	</div>
{/if}
