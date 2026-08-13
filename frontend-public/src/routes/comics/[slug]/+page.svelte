<script lang="ts">
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import LikeButton from '$lib/components/LikeButton.svelte';
	import FavoriteButton from '$lib/components/FavoriteButton.svelte';
	import DislikeButton from '$lib/components/DislikeButton.svelte';
	import Reader from '$lib/components/Reader.svelte';
	import CompactComicCard from '$lib/components/CompactComicCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import AgeGate from '$lib/components/AgeGate.svelte';
	import CommentList from '$lib/components/CommentList.svelte';
	import type { Comment } from '$lib/components/CommentList.svelte';
	import CommentForm from '$lib/components/CommentForm.svelte';
	import { currentUser } from '$lib/stores/auth';
	import { createCommentStream } from '$lib/stores/live-comments';
	import { isAgeConfirmed, confirmAge } from '$lib/ageGate';
	import { Eye, Download, BookOpen, Globe, Clock, ThumbsDown, BookOpenCheck, Bell, BellOff } from 'lucide-svelte';

	let { data } = $props();

	let comic = $derived(data.comic);
	let loading = $state(false);
	let error = $state('');
	let reading = $state(false);
	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);
	let coverIndex = $state(0);
	
	let likeStatus = $derived(data.likeStatus);

	let reactionLiked = $derived(data.likeStatus?.liked ?? false);
	let reactionDisliked = $derived(data.likeStatus?.disliked ?? false);
	let likeCount = $derived(data.comic?.like_count ?? 0);
	let dislikeCount = $derived(data.comic?.dislike_count ?? 0);
	let likeLoading = $state(false);
	let dislikeLoading = $state(false);

	let followingUploader = $state(false);
	let followLoading = $state(false);


	let related = $derived(data.related);
	let relatedLoading = $state(false);
	let relatedPage = $state(1);
	let relatedTotal = $state(0);
	let relatedHasMore = $state(false);

	let comments = $derived<Comment[]>(data.comments);
	let replyTarget = $state('');

	const slug = $derived(page.params.slug!);
	const user = $derived($currentUser);
	const coverSrc = $derived(comic?.cover_url || '');
	const pageImages = $derived(comic?.page_urls || []);
	const lightboxImages = $derived(coverSrc ? [coverSrc, ...pageImages] : pageImages);
	const previewSlots = $derived(pageImages.slice(0, 4));
	const hasMorePreviews = $derived(pageImages.length > 4);
	const isMature = $derived(comic?.age_rating === 'mature' || comic?.age_rating === 'explicit');
	const authorName = $derived(comic?.author || 'Unknown');

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

	function coverNext() {
		if (coverIndex < lightboxImages.length - 1) coverIndex++;
	}

	function coverPrev() {
		if (coverIndex > 0) coverIndex--;
	}

	function goCover(idx: number) {
		coverIndex = idx;
	}

	function coverKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowLeft') { e.preventDefault(); coverPrev(); }
		if (e.key === 'ArrowRight') { e.preventDefault(); coverNext(); }
	}

	$effect(() => {
		const stream = createCommentStream(comic.id);
		const unsub = stream.subscribe(() => {
			loadCommentsFallback();
		});
		return () => { stream.close(); unsub(); };
	});

	$effect(() => {
		if (user && comic?.uploader_id && user.id !== comic.uploader_id) {
			encore.comics.GetUploaderFollowStatus(comic.uploader_id)
				.then((res) => { followingUploader = res.following; })
				.catch(() => {});
		}
	});

	async function toggleLike() {
		if (likeLoading) return;
		likeLoading = true;
		try {
			const res = await encore.comics.ToggleLike(comic.id);
			reactionLiked = res.liked;
			likeCount = res.like_count;
			if (res.liked) { reactionDisliked = false; dislikeCount = Math.max(0, dislikeCount - 1); }
		} catch {} finally { likeLoading = false; }
	}

	async function toggleDislike() {
		if (dislikeLoading) return;
		dislikeLoading = true;
		try {
			const res = await encore.comics.ToggleDislike(comic.id);
			reactionDisliked = res.disliked;
			dislikeCount = res.dislike_count;
			if (res.disliked) { reactionLiked = false; likeCount = Math.max(0, likeCount - 1); }
		} catch {} finally { dislikeLoading = false; }
	}

	async function loadRelated(page?: number) {
		const p = page || 1;
		if (p === 1) relatedLoading = true;
		const timeout = setTimeout(() => { relatedLoading = false; }, 5000);
		try {
			const relRes = await encore.comics.ListComics({ Limit: 4, Page: p, Language: '', Search: '', SearchField: '', Tag: '', ExcludeMature: '', Sort: '' });
			clearTimeout(timeout);
			const filtered = relRes.comics.filter((c) => c.id !== comic.id);
			if (p === 1) {
				related = filtered;
			} else {
				related = [...related, ...filtered];
			}
			relatedTotal = relRes.total;
			relatedHasMore = filtered.length >= 4 && related.length < relRes.total;
			relatedPage = p;
		} catch {
			clearTimeout(timeout);
		}
		relatedLoading = false;
	}

	async function loadMoreRelated() {
		await loadRelated(relatedPage + 1);
	}

	async function loadCommentsFallback() {
		try {
			const commentRes = await encore.comics.ListComments(comic.id);
			comments = commentRes.comments || [];
		} catch {}
	}

	async function submitComment(bodyText: string, parentId?: string) {
		replyTarget = '';
		try {
			await encore.comics.CreateComment(comic.id, { body_text: bodyText, parent_id: parentId || '' });
			await loadCommentsFallback();
		} catch (e) {
			console.error('[comments] submit failed:', e);
		}
	}

	async function deleteComment(commentId: string) {
		try {
			await encore.comics.DeleteComment(commentId);
			await loadCommentsFallback();
		} catch (e) {
			console.error('[comments] delete failed:', e);
		}
	}

	async function flagComment(commentId: string) {
		try {
			await encore.comics.FlagComment(commentId, { reason: '' });
		} catch (e) {
			console.error('[comments] flag failed:', e);
		}
	}

	async function toggleFollowUploader() {
		if (!comic?.uploader_id || !user) return;
		followLoading = true;
		try {
			if (followingUploader) {
				await encore.comics.UnfollowUploader(comic.uploader_id);
				followingUploader = false;
			} else {
				await encore.comics.FollowUploader(comic.uploader_id);
				followingUploader = true;
			}
		} catch (e) {
			console.error('[follow] toggle failed:', e);
		} finally {
			followLoading = false;
		}
	}

	async function handleDownload() {
		try {
			const res = await encore.reading.RecordDownload(comic.id);
			if (!res.allowed) {
				alert(`${res.message}\nYou've used ${res.used} of ${res.limit} downloads.`);
			} else if (res.used >= res.limit * 0.8) {
				alert(`Download started!\nQuota warning: ${res.used} of ${res.limit} downloads used.`);
			}
		} catch {}
	}

	let ageGateOpen = $state(false);
	let agePendingAction = $state<'read' | 'download' | null>(null);

	function startReading() {
		if (isMature && !isAgeConfirmed(comic.id)) {
			agePendingAction = 'read';
			ageGateOpen = true;
		} else {
			reading = true;
		}
	}

	function startDownload() {
		if (isMature && !isAgeConfirmed(comic.id)) {
			agePendingAction = 'download';
			ageGateOpen = true;
		} else {
			handleDownload();
		}
	}

	function onAgeConfirm() {
		confirmAge(comic.id);
		ageGateOpen = false;
		const action = agePendingAction;
		agePendingAction = null;
		if (action === 'read') reading = true;
		else if (action === 'download') handleDownload();
	}

	function onAgeClose() {
		ageGateOpen = false;
		agePendingAction = null;
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
	<Reader comicId={comic.id} comicSlug={comic.slug} pageKeys={comic.page_keys || []} pageUrls={comic.page_urls} totalPages={comic.page_keys?.length || 1} />
{/if}

<Lightbox images={lightboxImages} open={lightboxOpen} startIndex={lightboxIndex} onClose={() => lightboxOpen = false} />

{#if comic}
	<AgeGate
		open={ageGateOpen}
		title={comic.title}
		author={authorName}
		ageRating={comic.age_rating}
		onConfirm={onAgeConfirm}
		onClose={onAgeClose}
	/>
{/if}

{#if loading}
	<div class="max-w-5xl mx-auto py-8">
		<div class="grid md:grid-cols-[3fr_2fr] gap-8 animate-pulse">
			<div class="space-y-3">
				<div class="aspect-3/4 rounded-xl bg-gray-200 dark:bg-gray-700"></div>
				<div class="flex gap-3">
					{#each Array(4) as _}
						<div class="flex-1 aspect-3/4 rounded bg-gray-200 dark:bg-gray-700"></div>
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
			<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div class="space-y-3" onkeydown={coverKeydown} tabindex="0" role="region" aria-label="Image carousel">
				<div class="aspect-3/4 rounded-xl bg-muted overflow-hidden relative group/cover">
					<button onclick={() => openLightbox(coverIndex)} class="w-full h-full p-0 border-0 bg-transparent cursor-zoom-in" aria-label="Open current image in lightbox">
						<img
							src={lightboxImages[coverIndex]}
							alt={coverIndex === 0 ? comic.title : `Page ${coverIndex}`}
							class="w-full h-full object-cover"
							loading="eager"
							onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; (document.querySelector('.cover-fallback') as HTMLElement).style.display = 'flex'; }}
						/>
					</button>
					<div class="cover-fallback w-full h-full items-center justify-center text-muted-foreground hidden absolute inset-0">
						<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
					</div>

					<button onclick={coverPrev} disabled={coverIndex === 0} class="absolute left-2 top-1/2 -translate-y-1/2 p-1.5 rounded-full bg-black/40 hover:bg-black/60 text-white/80 hover:text-white transition-all opacity-0 group-hover/cover:opacity-100 disabled:opacity-30 disabled:cursor-default" aria-label="Previous image">&larr;</button>
					<button onclick={coverNext} disabled={coverIndex >= lightboxImages.length - 1} class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-full bg-black/40 hover:bg-black/60 text-white/80 hover:text-white transition-all opacity-0 group-hover/cover:opacity-100 disabled:opacity-30 disabled:cursor-default" aria-label="Next image">&rarr;</button>

					<div class="absolute bottom-2 left-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{coverIndex + 1} / {lightboxImages.length}</div>
					<div class="absolute bottom-2 right-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{comic.page_keys?.length || 0} pages</div>
				</div>

				{#if lightboxImages.length > 1}
					<div class="grid grid-cols-5 gap-2">
						<button onclick={() => goCover(0)} class="aspect-3/4 rounded-lg bg-muted overflow-hidden border-2 transition-all {coverIndex === 0 ? 'border-primary' : 'border-border hover:border-primary/50'}">
							<img src={coverSrc} alt="Cover" class="w-full h-full object-cover" loading="lazy" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
						</button>
						{#each previewSlots as imgSrc, i}
							<button onclick={() => goCover(i + 1)} class="aspect-3/4 rounded-lg bg-muted overflow-hidden border-2 transition-all {coverIndex === i + 1 ? 'border-primary' : 'border-border hover:border-primary/50'}">
								<img src={imgSrc} alt={`Preview ${i + 1}`} class="w-full h-full object-cover" loading="lazy" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
							</button>
						{/each}
						{#if hasMorePreviews}
							<button onclick={() => goCover(previewSlots.length + 1)} class="aspect-3/4 rounded-lg bg-muted border border-border hover:border-primary/50 flex items-center justify-center text-muted-foreground text-sm font-medium transition-colors">
								+{pageImages.length - 4}
							</button>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Info Column -->
			<div>
				<div class="flex items-center gap-2">
					<h1 class="text-2xl font-bold leading-tight">{comic.title}</h1>
				</div>

				<p class="text-sm text-muted-foreground mt-1">By {authorName}</p>

				{#if comic.uploader_id}
					<div class="flex items-center gap-2 mt-1">
						<div class="size-6 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-[10px] font-medium text-purple-600 dark:text-purple-300 shrink-0">
							{(comic.author || 'U').charAt(0).toUpperCase()}
						</div>
						<span class="text-xs text-muted-foreground">Uploader</span>
						{#if user && user.id !== comic.uploader_id}
							<button
								onclick={toggleFollowUploader}
								disabled={followLoading}
								class="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full border border-border hover:bg-muted transition-colors"
								aria-label={followingUploader ? 'Unfollow uploader' : 'Follow uploader'}
							>
								{#if followingUploader}
									<BellOff class="size-3" /> Following
								{:else}
									<Bell class="size-3" /> Follow
								{/if}
							</button>
						{/if}
					</div>
				{/if}

				{#if comic.is_premium}
					<span class="inline-block mt-2 text-xs font-semibold text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900/30 px-2 py-0.5 rounded-full">Premium</span>
				{/if}

				<div class="flex flex-wrap items-center gap-1 mt-2">
						<span class="text-[10px] px-2 py-0.5 rounded-full bg-muted">{comic.age_rating?.replace('_', ' ')}</span>
						{#each (comic.tags || []) as tag}
							<span class="text-[10px] px-2 py-0.5 rounded-full bg-primary/10 text-primary">{tag}</span>
						{/each}
						{#if comic.status === 'published'}
							<span class="text-[10px] px-2 py-0.5 rounded-full font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 flex items-center gap-1"><BookOpenCheck class="size-3" />published</span>
						{/if}
						{#if comic.status === 'pending_review'}
							<span class="text-[10px] px-2 py-0.5 rounded-full font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">{comic.status?.replace('_', ' ')}</span>
						{/if}
						{#if comic.status === 'rejected'}
							<span class="text-[10px] px-2 py-0.5 rounded-full font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">{comic.status?.replace('_', ' ')}</span>
						{/if}
					</div>

				<p class="mt-3 text-muted-foreground text-sm leading-relaxed">{comic.description}</p>

				<div class="flex items-center gap-4 mt-3 py-2 border-y border-border">
					<LikeButton active={reactionLiked} count={likeCount} loading={likeLoading} onToggle={toggleLike} />
					<DislikeButton active={reactionDisliked} count={dislikeCount} loading={dislikeLoading} onToggle={toggleDislike} />
					<FavoriteButton comicId={comic.id} initialFavorited={likeStatus?.favorited} initialCount={comic.fav_count} />
				</div>

				<div class="grid grid-cols-2 gap-x-4 gap-y-1.5 mt-3 text-sm">
					<div class="flex items-center gap-2"><Eye class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Views</span><span class="ml-auto font-medium">{compactNum(comic.view_count)}</span></div>
					<div class="flex items-center gap-2"><Download class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Downloads</span><span class="ml-auto font-medium">{compactNum(comic.download_count)}</span></div>
					<div class="flex items-center gap-2"><ThumbsDown class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Dislikes</span><span class="ml-auto font-medium">{compactNum(comic.dislike_count)}</span></div>
					<div class="flex items-center gap-2"><BookOpen class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Pages</span><span class="ml-auto font-medium">{comic.page_keys?.length || 0}</span></div>
					<div class="flex items-center gap-2"><BookOpenCheck class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Published</span><span class="ml-auto font-medium text-xs">{formatDate(comic.published_at) || '-'}</span></div>
					<div class="flex items-center gap-2"><Clock class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Reading</span><span class="ml-auto font-medium">5 min</span></div>
					<div class="flex items-center gap-2"><Globe class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Language</span><span class="ml-auto font-medium uppercase">{comic.content_language}</span></div>
				</div>

				<div class="flex flex-col gap-2 mt-4">
					{#if comic.status === 'published'}
						<Button size="lg" class="w-full bg-emerald-600 hover:bg-emerald-700" onclick={startReading}>Start Reading</Button>
					{/if}
					{#if user}
						<Button size="lg" variant="outline" class="w-full" onclick={startDownload}>Download</Button>
					{:else}
						<Button size="lg" variant="outline" class="w-full" href="/login">Sign in</Button>
					{/if}
				</div>

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
						<div class="rounded-xl overflow-hidden animate-pulse">
							<div class="p-2 space-y-1">
								<div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
								<div class="h-2 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
							</div>
							<div class="aspect-3/4 bg-gray-200 dark:bg-gray-700 mx-2 mb-2 rounded"></div>
						</div>
					{/each}
				</div>
			{:else if related.length > 0}
				<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
					{#each related as r}<CompactComicCard {...r} />{/each}
				</div>
				{#if relatedHasMore}
					<div class="text-center mt-4">
						<Button variant="outline" size="sm" onclick={loadMoreRelated}>Load More</Button>
					</div>
				{/if}
			{:else}
				<p class="text-sm text-muted-foreground text-center py-8">No related comics found.</p>
			{/if}
		</section>

		<!-- Comments -->
		<section class="mt-12">
			<h2 class="text-xl font-semibold mb-4">Comments</h2>

			{#if user}
				<div class="mb-4">
					<CommentForm onSubmit={submitComment} placeholder="Write a comment..." />
				</div>
			{/if}

			{#if comments.length > 0}
				<CommentList comments={comments} onReply={(id) => replyTarget = id} onSubmitComment={submitComment} onDelete={deleteComment} onFlag={flagComment} userId={user?.id} role={user?.role} />
			{:else}
				<p class="text-sm text-muted-foreground text-center py-4">No comments yet. Be the first!</p>
			{/if}
		</section>
	</div>
{/if}
