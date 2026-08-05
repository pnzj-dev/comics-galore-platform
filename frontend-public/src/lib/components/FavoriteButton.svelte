<script lang="ts">
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		comicId: string;
		initialFavorited?: boolean;
		initialCount?: number;
	}

	let { comicId, initialFavorited = false, initialCount = 0 }: Props = $props();

	// svelte-ignore state_referenced_locally
	let favorited = $state(initialFavorited);
	// svelte-ignore state_referenced_locally
	let count = $state(initialCount);

	async function toggle() {
		const next = !favorited;
		favorited = next;
		count += next ? 1 : -1;

		try {
			if (next) {
				const res = await api.post<{ favorited: boolean; fav_count: number }>(`/comics/${comicId}/favorite`);
				favorited = res.favorited;
				count = res.fav_count;
			} else {
				await api.delete(`/comics/${comicId}/favorite`);
			}
		} catch {
			favorited = !next;
			count += next ? -1 : 1;
		}
	}
</script>

<Button variant="ghost" size="sm" onclick={toggle}>
	{#if favorited}
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
	{:else}
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
	{/if}
</Button>
