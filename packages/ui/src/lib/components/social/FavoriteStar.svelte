<script lang="ts">
	import { encore } from '$lib/api/encore';

	interface Props {
		comicId: string;
		initialFavorited?: boolean;
		size?: 'sm' | 'md';
		variant?: 'overlay' | 'subtle';
		class?: string;
		onUnfavorite?: (id: string) => void;
	}

	let {
		comicId,
		initialFavorited = false,
		size = 'md',
		variant = 'overlay',
		class: className = '',
		onUnfavorite
	}: Props = $props();

	// svelte-ignore state_referenced_locally
	let favorited = $state(initialFavorited);
	let favHovered = $state(false);
	let liking = $state(false);

	// Hover previews the state a click would produce:
	// favorited + hovered → unfilled (click unfavorites); not favorited + hovered → filled (click favorites).
	const filled = $derived(favHovered ? !favorited : favorited);
	const iconSize = $derived(size === 'sm' ? 'size-3' : 'size-4');
	const btnSize = $derived(size === 'sm' ? 'size-6' : 'size-8');
	const unfilledClass = $derived(
		variant === 'overlay' ? 'text-white' : 'text-muted-foreground hover:text-yellow-400'
	);
	const bgClass = $derived(
		variant === 'overlay' ? 'bg-black/40 backdrop-blur-sm hover:bg-black/60' : ''
	);

	async function toggleFavorite(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (liking) return;
		favHovered = false;
		const next = !favorited;
		favorited = next;
		liking = true;
		try {
			const res = await encore.comics.ToggleFavorite(comicId);
			favorited = res.favorited;
			if (!res.favorited) onUnfavorite?.(comicId);
		} catch {
			favorited = !next;
		} finally {
			liking = false;
		}
	}
</script>

<button
	type="button"
	onclick={toggleFavorite}
	onmouseenter={() => (favHovered = true)}
	onmouseleave={() => (favHovered = false)}
	disabled={liking}
	aria-label={favorited ? 'Remove from favorites' : 'Add to favorites'}
	class="flex items-center justify-center rounded-full {btnSize} {bgClass} {filled ? 'text-yellow-400' : unfilledClass} {liking ? 'opacity-60' : ''} {className}"
>
	<svg class={iconSize} viewBox="0 0 24 24" fill={filled ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
</button>
