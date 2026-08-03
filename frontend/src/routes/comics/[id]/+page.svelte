<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import LikeButton from '$lib/components/LikeButton.svelte';
	import FavoriteButton from '$lib/components/FavoriteButton.svelte';
	import Reader from '$lib/components/Reader.svelte';
	import { currentUser } from '$lib/stores/auth';

	let comic = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let reading = $state(false);
	let likeStatus = $state<{ liked: boolean; favorited: boolean } | null>(null);

	const id = $derived($page.params.id);
	const user = $derived($currentUser);

	onMount(async () => {
		try {
			comic = await api.get(`/comics/${id}`);
			if (user) {
				try {
					likeStatus = await api.get(`/comics/${id}/like-status`);
				} catch { /* not logged in or error */ }
			}
		} catch (err) {
			error = (err as Error).message;
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>{comic?.title || 'Comic'} - Comics Galore</title>
</svelte:head>

{#if reading && comic}
	<Reader comicId={comic.id} pageKeys={comic.page_keys || []} totalPages={(comic.page_keys || []).length || 1} />
{/if}

{#if loading}
	<p class="text-muted-foreground py-20 text-center">Loading...</p>
{:else if error}
	<p class="text-destructive py-20 text-center">{error}</p>
{:else if comic}
	<div class="max-w-4xl mx-auto py-8">
		<div class="grid md:grid-cols-[300px_1fr] gap-8">
			<div class="aspect-[3/4] rounded-xl bg-muted flex items-center justify-center text-muted-foreground">
				Cover
			</div>

			<div>
				<h1 class="text-3xl font-bold">{comic.title}</h1>

				<div class="flex items-center gap-2 mt-2">
					<div class="text-xs rounded-full px-2 py-0.5 {comic.status === 'published' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : comic.status === 'pending_review' ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'}">
						{comic.status?.replace('_', ' ')}
					</div>
					<span class="text-xs bg-muted px-2 py-0.5 rounded-full">{comic.age_rating?.replace('_', ' ')}</span>
				</div>

				<p class="mt-4 text-muted-foreground">{comic.description}</p>

				<div class="flex items-center gap-3 mt-4">
					<LikeButton comicId={comic.id} initialLiked={likeStatus?.liked} initialCount={comic.like_count} />
					<FavoriteButton comicId={comic.id} initialFavorited={likeStatus?.favorited} initialCount={comic.fav_count} />
				</div>

				<div class="flex gap-2 mt-4">
					{#if comic.status === 'published'}
						<Button size="lg" onclick={() => reading = true}>Start Reading</Button>
					{/if}
					<Button size="lg" variant="outline" href={`/comics/${comic.id}/download`}>Download</Button>
				</div>

				<div class="grid grid-cols-2 gap-4 mt-6 text-sm text-muted-foreground">
					<div>Language: {comic.content_language}</div>
					<div>Views: {comic.view_count}</div>
					<div>Downloads: {comic.download_count}</div>
					<div>Pages: {(comic.page_keys || []).length}</div>
				</div>

				{#if (comic.tags || []).length > 0}
					<div class="flex flex-wrap gap-1 mt-4">
						{#each comic.tags as tag}
							<span class="text-xs bg-secondary text-secondary-foreground px-2 py-0.5 rounded-full">{tag}</span>
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
	</div>
{/if}
