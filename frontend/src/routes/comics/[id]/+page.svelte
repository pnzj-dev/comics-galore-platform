<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LikeButton from '$lib/components/LikeButton.svelte';
	import FavoriteButton from '$lib/components/FavoriteButton.svelte';
	import Reader from '$lib/components/Reader.svelte';
	import ComicCard from '$lib/components/ComicCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import CommentList from '$lib/components/CommentList.svelte';
	import type { Comment } from '$lib/components/CommentList.svelte';
	import CommentForm from '$lib/components/CommentForm.svelte';
	import { currentUser } from '$lib/stores/auth';
	import { Eye, Download, BookOpen, Globe, Clock } from 'lucide-svelte';

	let comic = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let reading = $state(false);
	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);
	let likeStatus = $state<{ liked: boolean; favorited: boolean } | null>(null);
	let related = $state<any[]>([]);
	let relatedLoading = $state(true);

	let comments = $state<Comment[]>([]);
	let commentsLoading = $state(true);
	let replyTarget = $state('');

	const id = $derived($page.params.id);
	const user = $derived($currentUser);
	const coverSrc = $derived(comic?.cover_key ? `/media/${comic.cover_key}` : '');
	const pageImages = $derived((comic?.page_keys || []).map((k: string) => `/media/${k}`));
	const previewSlots = $derived(pageImages.slice(0, 4));
	const hasMorePreviews = $derived(pageImages.length > 4);

	function compactNum(n: number): string {
		if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
		if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k';
		return String(n);
	}

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function openLightbox(index: number) {
		lightboxIndex = index;
		lightboxOpen = true;
	}

	onMount(async () => {
		try {
			comic = await api.get(`/comics/${id}`);
			if (user) { try { likeStatus = await api.get(`/comics/${id}/like-status`); } catch {} }
			loadRelated();
			loadComments();
		} catch (err) { error = (err as Error).message; }
		loading = false;
	});

	async function loadRelated() {
		try {
			const res = await api.get<{ comics: any[] }>('/comics?limit=4');
			related = res.comics.filter((c: any) => c.id !== id).slice(0, 4);
		} catch {}
		relatedLoading = false;
	}

			async function loadComments() {
		try {
			const res = await api.get<{ comments: Comment[] }>(`/comics/${id}/comments`);
			comments = res.comments;
		} catch {}
		commentsLoading = false;
	}

	async function submitComment(bodyText: string, parentId?: string) {
		await api.post(`/comics/${id}/comments`, { body_text: bodyText, parent_id: parentId || '' });
		replyTarget = '';
		await loadComments();
	}

	async function deleteComment(commentId: string) {
		await api.delete(`/comments/${commentId}`);
		await loadComments();
	}

	async function handleDownload() {
		try {
			const res = await api.post<{ allowed: boolean; used: number; limit: number; message: string }>(`/comics/${comic.id}/download`);
			if (!res.allowed) {
				alert(`${res.message}\nYou've used ${res.used} of ${res.limit} downloads.`);
			} else if (res.used >= res.limit * 0.8) {
				alert(`Download started!\nQuota warning: ${res.used} of ${res.limit} downloads used.`);
			}
		} catch {}
	}
</script>

<svelte:head>
	<title>{comic?.title || 'Comic'} - Comics Galore</title>
	<meta property="og:title" content={comic?.title || 'Comic'} />
	<meta property="og:description" content={comic?.description || 'Discover and read comics on Comics Galore'} />
	<meta property="og:type" content="article" />
	<meta name="description" content={comic?.description || ''} />
</svelte:head>

{#if reading && comic}
	<Reader comicId={comic.id} pageKeys={comic.page_keys || []} totalPages={comic.page_keys?.length || 1} />
{/if}

<Lightbox images={pageImages} open={lightboxOpen} startIndex={lightboxIndex} onClose={() => lightboxOpen = false} />

{#if loading}
	<div class="max-w-5xl mx-auto py-8">
		<div class="grid md:grid-cols-[3fr_2fr] gap-8 animate-pulse">
			<div class="space-y-3">
				<div class="aspect-[3/4] rounded-xl bg-gray-200 dark:bg-gray-700"></div>
				<div class="flex gap-3">
					{#each Array(4) as _}
						<div class="flex-1 aspect-[3/4] rounded bg-gray-200 dark:bg-gray-700"></div>
					{/each}
				</div>
			</div>
			<div class="space-y-4">
				<div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-5/6"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
				<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
			</div>
		</div>
	</div>
{:else if error}
	<p class="text-destructive py-20 text-center">{error}</p>
{:else if comic}
	<div class="max-w-5xl mx-auto py-8">
		<div class="grid md:grid-cols-[3fr_2fr] gap-8">
			<!-- Cover Column -->
			<div class="space-y-3">
				<div class="aspect-[3/4] rounded-xl bg-muted overflow-hidden relative">
					<button onclick={() => openLightbox(0)} class="w-full h-full p-0 border-0 bg-transparent cursor-zoom-in" aria-label="Open cover image">
						<img
							src={coverSrc}
							alt={comic.title}
							class="w-full h-full object-cover"
							loading="eager"
							onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; (document.querySelector('.cover-fallback') as HTMLElement).style.display = 'flex'; }}
						/>
					</button>
					<div class="cover-fallback w-full h-full items-center justify-center text-muted-foreground hidden absolute inset-0">
						<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
					</div>
					<div class="absolute bottom-2 right-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{comic.page_keys?.length || 0} pages</div>
				</div>

				{#if previewSlots.length > 0}
					<div class="grid grid-cols-4 gap-2">
						{#each previewSlots as imgSrc, i}
							<button onclick={() => openLightbox(i + 1)} class="aspect-[3/4] rounded-lg bg-muted overflow-hidden border border-border hover:border-primary/50 transition-colors">
								<img src={imgSrc} alt={`Preview ${i + 1}`} class="w-full h-full object-cover" loading="lazy" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
								<div class="w-full h-full items-center justify-center text-muted-foreground text-xs hidden">Preview</div>
							</button>
						{/each}
						{#if hasMorePreviews}
							<button onclick={() => openLightbox(4)} class="aspect-[3/4] rounded-lg bg-muted border border-border hover:border-primary/50 flex items-center justify-center text-muted-foreground text-sm font-medium transition-colors">
								+{pageImages.length - 4}
							</button>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Info Column -->
			<div>
				<h1 class="text-2xl font-bold leading-tight">{comic.title}</h1>

				<div class="flex flex-wrap items-center gap-2 mt-2">
					<span class="text-xs rounded-full px-2 py-0.5 font-medium {comic.status === 'published' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : comic.status === 'pending_review' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'}">{comic.status?.replace('_', ' ')}</span>
					<span class="text-xs bg-muted px-2 py-0.5 rounded-full">{comic.age_rating?.replace('_', ' ')}</span>
					{#if comic.is_premium}
						<span class="text-xs font-semibold text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900/30 px-2 py-0.5 rounded-full">Premium</span>
					{/if}
				</div>

				<div class="flex items-center gap-2 mt-3">
					<div class="size-8 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-xs font-medium text-purple-600 dark:text-purple-300 flex-shrink-0">
						{comic.title?.charAt(0).toUpperCase() || '?'}
					</div>
					<span class="text-sm font-medium text-muted-foreground">Author</span>
				</div>

				<p class="mt-3 text-muted-foreground text-sm leading-relaxed">{comic.description}</p>

				<div class="flex items-center gap-4 mt-3 py-2 border-y border-border">
					<LikeButton comicId={comic.id} initialLiked={likeStatus?.liked} initialCount={comic.like_count} />
					<FavoriteButton comicId={comic.id} initialFavorited={likeStatus?.favorited} initialCount={comic.fav_count} />
				</div>

				<div class="grid grid-cols-2 gap-x-4 gap-y-1.5 mt-3 text-sm">
					<div class="flex items-center gap-2"><Eye class="size-3.5 text-muted-foreground flex-shrink-0" /><span class="text-muted-foreground">Views</span><span class="ml-auto font-medium">{compactNum(comic.view_count)}</span></div>
					<div class="flex items-center gap-2"><Download class="size-3.5 text-muted-foreground flex-shrink-0" /><span class="text-muted-foreground">Downloads</span><span class="ml-auto font-medium">{compactNum(comic.download_count)}</span></div>
					<div class="flex items-center gap-2"><BookOpen class="size-3.5 text-muted-foreground flex-shrink-0" /><span class="text-muted-foreground">Pages</span><span class="ml-auto font-medium">{comic.page_keys?.length || 0}</span></div>
					<div class="flex items-center gap-2"><span class="text-muted-foreground ml-[22px]">Published</span><span class="ml-auto font-medium text-xs">{formatDate(comic.published_at) || '-'}</span></div>
					<div class="flex items-center gap-2"><Clock class="size-3.5 text-muted-foreground flex-shrink-0" /><span class="text-muted-foreground">Reading</span><span class="ml-auto font-medium">5 min</span></div>
					<div class="flex items-center gap-2"><Globe class="size-3.5 text-muted-foreground flex-shrink-0" /><span class="text-muted-foreground">Language</span><span class="ml-auto font-medium uppercase">{comic.content_language}</span></div>
				</div>

				<div class="flex gap-2 mt-4">
					{#if comic.status === 'published'}
						<Button size="lg" class="bg-emerald-600 hover:bg-emerald-700" onclick={() => reading = true}>Start Reading</Button>
					{/if}
					{#if user}
						<Button size="lg" variant="outline" onclick={handleDownload}>Download</Button>
					{:else}
						<Button size="lg" variant="outline" href="/login">Sign in</Button>
					{/if}
				</div>

				{#if (comic.tags || []).length > 0}
					<div class="flex flex-wrap gap-1 mt-3">
						{#each comic.tags as tag}
							<span class="text-[10px] px-2 py-0.5 rounded-full bg-primary/10 text-primary">{tag}</span>
						{/each}
					</div>
				{/if}

				{#if comic.rejection_reason}
					<div class="mt-3 p-3 rounded-lg bg-destructive/10 text-destructive text-sm">Rejection reason: {comic.rejection_reason}</div>
				{/if}
			</div>
		</div>

		<!-- Related Comics -->
		<section class="mt-12">
			<h2 class="text-xl font-semibold mb-4">Related Comics</h2>
			{#if relatedLoading}
				<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
					{#each Array(4) as _}
						<div class="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden animate-pulse">
							<div class="aspect-[3/4] bg-gray-200 dark:bg-gray-700"></div>
							<div class="p-3 space-y-2"><div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div><div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div></div>
						</div>
					{/each}
				</div>
			{:else if related.length > 0}
				<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
					{#each related as r}<ComicCard {...r} />{/each}
				</div>
			{:else}
				<p class="text-sm text-muted-foreground text-center py-8">No related comics found.</p>
			{/if}
		</section>

		<!-- Comments -->
		<section class="mt-12">
			<h2 class="text-xl font-semibold mb-4">Comments</h2>

			{#if user}
				<div class="mb-4">
					<CommentForm onSubmit={submitComment} parentId={replyTarget} placeholder={replyTarget ? 'Write a reply...' : 'Write a comment...'} />
				</div>
			{/if}

			{#if commentsLoading}
				<div class="space-y-3">
					{#each Array(3) as _}
						<div class="flex gap-2 animate-pulse">
							<div class="size-7 rounded-full bg-gray-200 dark:bg-gray-700 flex-shrink-0"></div>
							<div class="flex-1 space-y-1.5">
								<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/4"></div>
								<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
							</div>
						</div>
					{/each}
				</div>
			{:else if comments.length > 0}
				<CommentList comments={comments} onReply={(id) => replyTarget = id} onDelete={deleteComment} userId={user?.id} role={user?.role} />
			{:else}
				<p class="text-sm text-muted-foreground text-center py-4">No comments yet. Be the first!</p>
			{/if}
		</section>
	</div>
{/if}
