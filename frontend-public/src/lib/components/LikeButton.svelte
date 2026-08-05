<script lang="ts">
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		comicId: string;
		initialLiked?: boolean;
		initialCount?: number;
	}

	let { comicId, initialLiked = false, initialCount = 0 }: Props = $props();

	// svelte-ignore state_referenced_locally
	let liked = $state(initialLiked);
	// svelte-ignore state_referenced_locally
	let count = $state(initialCount);

	async function toggle() {
		const next = !liked;
		liked = next;
		count += next ? 1 : -1;

		try {
			if (next) {
				const res = await api.post<{ liked: boolean; like_count: number }>(`/comics/${comicId}/like`);
				liked = res.liked;
				count = res.like_count;
			} else {
				await api.delete(`/comics/${comicId}/like`);
			}
		} catch {
			liked = !next;
			count += next ? -1 : 1;
		}
	}
</script>

<Button variant="ghost" size="sm" onclick={toggle}>
	{#if liked}
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="2"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/></svg>
	{:else}
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/></svg>
	{/if}
	<span class="ml-1">{count}</span>
</Button>
