<script lang="ts">
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import ReactionButton from '$lib/components/social/ReactionButton.svelte';
	import FavoriteButton from '$lib/components/social/FavoriteButton.svelte';
	import Reader from '$lib/components/comics/Reader.svelte';
	import CompactComicCard from '$lib/components/comics/CompactComicCard.svelte';
	import Lightbox from '$lib/components/comics/Lightbox.svelte';
	import AgeGate from '$lib/components/comics/AgeGate.svelte';
	import CommentList from '$lib/components/social/CommentList.svelte';
	import type { Comment } from '$lib/components/social/CommentList.svelte';
	import CommentForm from '$lib/components/social/CommentForm.svelte';
	import AdBanner from '$lib/components/home/AdBanner.svelte';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { currentUser } from '$lib/stores/auth';
	import { createCommentStream } from '$lib/stores/live-comments';
	import { modal } from '$lib/stores/modal.svelte';
	import { quotaRefresh } from '$lib/stores/quota.svelte';
	import { setAddToList } from '$lib/stores/add-to-list.svelte';
	import { setNewMessage } from '$lib/stores/new-message.svelte';
	import { isAgeConfirmed, confirmAge } from '$lib/ageGate';
	import { formatDate, formatCompactNumber, buildDownloadFilename } from '$lib/utils/format';
	import { t } from '$lib/i18n';
	import { Eye, Download, BookOpen, Globe, Clock, ThumbsDown, BookOpenCheck, Bell, BellOff, ListPlus, Lock, MessageCircle } from 'lucide-svelte';

	let { data } = $props();

	let comic = $derived(data.comic);
	let loading = $state(false);
	let error = $state('');
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;
	let reading = $state(false);
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
	const canReact = $derived(!!user);
	// The web reader requires a paid tier (Bronze+); staff are always allowed.
	const canRead = $derived(!!user && (
		user.role === 'admin' || user.role === 'moderator' || user.role === 'uploader' ||
		(!!user.tier && user.tier !== 'free')
	));
	const coverSrc = $derived(comic?.cover_url || '');
	const pageCount = $derived(comic?.page_count || 0);

	// Download quota: fetched for signed-in users and refreshed whenever a
	// boost purchase completes.
	let quota = $state<{ used: number; limit: number } | null>(null);
	const quotaExhausted = $derived(quota !== null && quota.limit > 0 && quota.used >= quota.limit);

	async function loadQuota() {
		if (!user) { quota = null; return; }
		try {
			const res = await encore.reading.GetQuotaStatus();
			quota = { used: res.used, limit: res.limit };
		} catch {
			quota = null;
		}
	}

	$effect(() => {
		quotaRefresh.tick;
		if (user) loadQuota();
	});

	// Preview gallery: fetch the first few pages lazily (the reader fetches its
	// own window). Free/anonymous callers receive locked (URL-less) pages past
	// the tier preview limit; paid tiers get the full set sharp.
	type GalleryEntry = { url: string; locked: boolean };
	let pageEntries = $state<GalleryEntry[]>([]);

	const gallery = $derived<GalleryEntry[]>(coverSrc ? [{ url: coverSrc, locked: false }, ...pageEntries] : pageEntries);
	const lightboxImages = $derived<string[]>(gallery.filter((g) => !g.locked).map((g) => g.url));
	const previewSlots = $derived(pageEntries.slice(0, 4));
	const hasMorePreviews = $derived(pageEntries.length > 4);
	const isMature = $derived(comic?.age_rating === 'mature' || comic?.age_rating === 'explicit');
	const matureLocked = $derived(!!comic?.mature_locked);
	const authorName = $derived(comic?.author || 'Unknown');

	$effect(() => {
		if (!comic || matureLocked) return;
		encore.comics.GetComicPages(slug, { Offset: 0, Limit: 10, Preview: true })
			.then((res) => {
				pageEntries = (res.pages || []).map((p) => ({ url: p.url || '', locked: !!p.locked }));
			})
			.catch(() => {});
	});

	function unlockPreviews() {
		modal.open('checkout');
	}

	function openLightbox(index: number) {
		const entry = gallery[index];
		if (!entry || entry.locked) {
			unlockPreviews();
			return;
		}
		// Map gallery index → sharp-only lightbox index (locked entries excluded).
		let sharpIndex = 0;
		for (let i = 0; i < index; i++) {
			if (!gallery[i].locked) sharpIndex++;
		}
		lightboxIndex = sharpIndex;
		modal.open('lightbox');
	}

	function coverNext() {
		if (coverIndex < gallery.length - 1) coverIndex++;
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
			await encore.comics.CreateComment(comic.id, { body_text: bodyText, parent_id: parentId || '', turnstile_token: turnstileToken || '' });
			await loadCommentsFallback();
		} catch (e) {
			console.error('[comments] submit failed:', e);
		} finally {
			turnstileToken = null;
			turnstileReset++;
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

	function messageUploader() {
		if (!comic?.uploader_id || !user || user.id === comic.uploader_id) return;
		setNewMessage(comic.uploader_id, authorName);
		modal.open('new-message');
	}

	async function handleDownload() {
		try {
			const res = await encore.reading.RecordDownload(comic.id);
			quota = { used: res.used, limit: res.limit };
			if (!res.allowed) {
				modal.open('boost');
				return;
			}
			const name = buildDownloadFilename([comic.author, comic.title, comic.volume, comic.issue_number]);
			window.location.href = `/api/download/${comic.file_key}?size=${comic.file_size_bytes}&name=${encodeURIComponent(name)}`;
		} catch {}
	}

	let agePendingAction = $state<'read' | 'download' | null>(null);

	function startReading() {
		if (matureLocked || !canRead) return;
		if (isMature && !isAgeConfirmed(comic.id)) {
			agePendingAction = 'read';
			modal.open('agegate');
		} else {
			reading = true;
		}
	}

	function startDownload() {
		if (matureLocked) return;
		if (isMature && !isAgeConfirmed(comic.id)) {
			agePendingAction = 'download';
			modal.open('agegate');
		} else {
			handleDownload();
		}
	}

	function onAgeConfirm() {
		confirmAge(comic.id);
		modal.close('agegate');
		const action = agePendingAction;
		agePendingAction = null;
		if (action === 'read') reading = true;
		else if (action === 'download') handleDownload();
	}

	function onAgeClose() {
		modal.close('agegate');
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
	<Reader comicId={comic.id} comicSlug={comic.slug} pageCount={comic.page_count || 0} readingDirection={comic.reading_direction || 'ltr'} />
{/if}

<Lightbox images={lightboxImages} startIndex={lightboxIndex} />

{#if comic}
	<AgeGate
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
					{#if gallery[coverIndex]?.locked}
						<button onclick={unlockPreviews} class="w-full h-full p-0 border-0 bg-transparent cursor-pointer" aria-label={t('comic.upgradeForMorePreviews')}>
							<div class="w-full h-full bg-muted flex flex-col items-center justify-center gap-2 text-muted-foreground p-4 text-center">
								<Lock class="size-6" />
								<span class="text-xs font-semibold">{t('comic.upgradeForMorePreviews')}</span>
								<span class="text-xs underline text-primary">View plans</span>
							</div>
						</button>
					{:else}
						<button onclick={() => !matureLocked && openLightbox(coverIndex)} class="w-full h-full p-0 border-0 bg-transparent {matureLocked ? 'cursor-default' : 'cursor-zoom-in'}" aria-label="Open current image in lightbox">
							<img
								src={gallery[coverIndex]?.url}
								alt={coverIndex === 0 ? comic.title : `Page ${coverIndex}`}
								class="w-full h-full object-cover {matureLocked ? 'blur-xl scale-110' : ''}"
								loading="eager"
								onerror={(e) => { (e.target as HTMLInputElement).style.display = 'none'; (document.querySelector('.cover-fallback') as HTMLElement).style.display = 'flex'; }}
							/>
						</button>
					{/if}
					{#if matureLocked}
						<div class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-black/40 text-white p-4 text-center">
							<span class="text-xs font-bold uppercase tracking-wide px-2 py-0.5 rounded-full bg-red-600">{comic.age_rating?.replace('_', ' ')}</span>
							<span class="text-sm font-semibold">Mature content</span>
							<span class="text-xs text-white/80">Upgrade to a paid tier to read this comic.</span>
						</div>
					{/if}
					<div class="cover-fallback w-full h-full items-center justify-center text-muted-foreground hidden absolute inset-0">
						<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
					</div>

					<button onclick={coverPrev} disabled={coverIndex === 0} class="absolute left-2 top-1/2 -translate-y-1/2 p-1.5 rounded-full bg-black/40 hover:bg-black/60 text-white/80 hover:text-white transition-all opacity-0 group-hover/cover:opacity-100 disabled:opacity-30 disabled:cursor-default" aria-label="Previous image">&larr;</button>
					<button onclick={coverNext} disabled={coverIndex >= gallery.length - 1} class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-full bg-black/40 hover:bg-black/60 text-white/80 hover:text-white transition-all opacity-0 group-hover/cover:opacity-100 disabled:opacity-30 disabled:cursor-default" aria-label="Next image">&rarr;</button>

					<div class="absolute bottom-2 left-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{coverIndex + 1} / {gallery.length}</div>
					<div class="absolute bottom-2 right-2 bg-black/60 text-white text-xs px-2 py-1 rounded">{pageCount} pages</div>
				</div>

				{#if gallery.length > 1}
					<div class="grid grid-cols-5 gap-2">
						<button onclick={() => goCover(0)} class="aspect-3/4 rounded-lg bg-muted overflow-hidden border-2 transition-all {coverIndex === 0 ? 'border-primary' : 'border-border hover:border-primary/50'}">
							<img src={coverSrc} alt="Cover" class="w-full h-full object-cover" loading="lazy" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
						</button>
						{#each previewSlots as slot, i}
							{#if slot.locked}
								<button onclick={unlockPreviews} class="aspect-3/4 rounded-lg bg-muted overflow-hidden border-2 border-border hover:border-primary/50 flex items-center justify-center" aria-label={t('comic.upgradeForMorePreviews')}>
									<Lock class="size-4 text-muted-foreground" />
								</button>
							{:else}
								<button onclick={() => goCover(i + 1)} class="aspect-3/4 rounded-lg bg-muted overflow-hidden border-2 transition-all {coverIndex === i + 1 ? 'border-primary' : 'border-border hover:border-primary/50'}">
									<img src={slot.url} alt={`Preview ${i + 1}`} class="w-full h-full object-cover" loading="lazy" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
								</button>
							{/if}
						{/each}
						{#if hasMorePreviews}
							<button onclick={() => goCover(previewSlots.length + 1)} class="aspect-3/4 rounded-lg bg-muted border border-border hover:border-primary/50 flex items-center justify-center text-muted-foreground text-sm font-medium transition-colors">
								+{pageEntries.length - 4}
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

				<p class="text-sm text-muted-foreground mt-1">
					By
					{#if user && user.id !== comic.uploader_id}
						<button onclick={messageUploader} class="font-medium text-foreground hover:text-primary underline-offset-2 hover:underline transition-colors" title="Message uploader">{authorName}</button>
					{:else}
						{authorName}
					{/if}
				</p>

				{#if comic.uploader_id}
					<div class="flex items-center gap-2 mt-1">
						{#if user && user.id !== comic.uploader_id}
							<button onclick={messageUploader} class="size-6 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-[10px] font-medium text-purple-600 dark:text-purple-300 shrink-0 hover:opacity-80 transition-opacity" aria-label="Message uploader">
								{(comic.author || 'U').charAt(0).toUpperCase()}
							</button>
						{:else}
							<div class="size-6 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center text-[10px] font-medium text-purple-600 dark:text-purple-300 shrink-0">
								{(comic.author || 'U').charAt(0).toUpperCase()}
							</div>
						{/if}
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
							<button
								onclick={messageUploader}
								class="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full border border-border hover:bg-muted transition-colors"
								aria-label="Message uploader"
							>
								<MessageCircle class="size-3" /> Message
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
					<span title={!canReact ? t('comic.signInToLike') : undefined} class="inline-flex">
						<ReactionButton kind="like" active={reactionLiked} count={likeCount} loading={likeLoading} disabled={!canReact} onToggle={toggleLike} />
					</span>
					<span title={!canReact ? t('comic.signInToDislike') : undefined} class="inline-flex">
						<ReactionButton kind="dislike" active={reactionDisliked} count={dislikeCount} loading={dislikeLoading} disabled={!canReact} onToggle={toggleDislike} />
					</span>
					<span title={!canReact ? t('comic.signInToFavorite') : undefined} class="inline-flex">
						<FavoriteButton comicId={comic.id} initialFavorited={likeStatus?.favorited} initialCount={comic.fav_count} disabled={!canReact} />
					</span>
					{#if user}
						<button
							onclick={() => { setAddToList(comic.id, comic.title); modal.open('add-to-list'); }}
							class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
						>
							<ListPlus class="size-4" /> Add to list
						</button>
					{/if}
				</div>

				<div class="grid grid-cols-2 gap-x-4 gap-y-1.5 mt-3 text-sm">
					<div class="flex items-center gap-2"><Eye class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Views</span><span class="ml-auto font-medium">{formatCompactNumber(comic.view_count)}</span></div>
					<div class="flex items-center gap-2"><Download class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Downloads</span><span class="ml-auto font-medium">{formatCompactNumber(comic.download_count)}</span></div>
					<div class="flex items-center gap-2"><ThumbsDown class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Dislikes</span><span class="ml-auto font-medium">{formatCompactNumber(comic.dislike_count)}</span></div>
					<div class="flex items-center gap-2"><BookOpen class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Pages</span><span class="ml-auto font-medium">{pageCount}</span></div>
					<div class="flex items-center gap-2"><BookOpenCheck class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Published</span><span class="ml-auto font-medium text-xs">{formatDate(comic.published_at) || '-'}</span></div>
					<div class="flex items-center gap-2"><Clock class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Reading</span><span class="ml-auto font-medium">5 min</span></div>
					<div class="flex items-center gap-2"><Globe class="size-3.5 text-muted-foreground shrink-0" /><span class="text-muted-foreground">Language</span><span class="ml-auto font-medium uppercase">{comic.content_language}</span></div>
				</div>

				<div class="flex flex-col gap-2 mt-4">
					{#if matureLocked}
						{#if user}
							<Button size="lg" class="w-full" href="/pricing">Upgrade to read mature content</Button>
						{:else}
							<Button size="lg" class="w-full" onclick={() => modal.open('login')}>Sign in</Button>
						{/if}
					{:else}
						{#if comic.status === 'published' && comic.extraction_status === 'processing'}
							<Button size="lg" class="w-full" disabled>
								Processing pages…
							</Button>
						{:else if comic.status === 'published' && canRead}
							<Button size="lg" class="w-full bg-emerald-600 hover:bg-emerald-700" onclick={startReading}>Start Reading</Button>
						{:else if comic.status === 'published' && user}
							<Button size="lg" class="w-full bg-emerald-600 hover:bg-emerald-700" onclick={() => modal.open('checkout')}>{t('comic.upgradeToRead')}</Button>
						{/if}
						{#if user}
							{#if quotaExhausted}
								<Button size="lg" class="w-full" onclick={() => modal.open('boost')}>Boost Quota</Button>
							{:else}
								<Button size="lg" variant="outline" class="w-full" onclick={startDownload}>Download</Button>
							{/if}
						{:else}
							<Button size="lg" variant="outline" class="w-full" onclick={() => modal.open('login')}>Sign in</Button>
						{/if}
					{/if}
				</div>

				<AdBanner class="mt-4" />

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

			{#if comic.comments_enabled !== true}
				<p class="text-sm text-muted-foreground text-center py-4">Comments are currently disabled.</p>
			{:else}
				{#if user}
					<div class="mb-4">
						<CommentForm onSubmit={submitComment} placeholder="Write a comment..." />
						<Turnstile action="comment" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />
					</div>
				{/if}

				{#if comments.length > 0}
					<CommentList comments={comments} onReply={(id) => replyTarget = id} onSubmitComment={submitComment} onDelete={deleteComment} onFlag={flagComment} userId={user?.id} role={user?.role} />
				{:else}
					<p class="text-sm text-muted-foreground text-center py-4">No comments yet. Be the first!</p>
				{/if}
			{/if}
		</section>
	</div>
{/if}
