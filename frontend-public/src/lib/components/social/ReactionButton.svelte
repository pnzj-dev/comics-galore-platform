<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		kind: 'like' | 'dislike';
		active?: boolean;
		count?: number;
		loading?: boolean;
		disabled?: boolean;
		onToggle?: () => void;
	}

	let { kind, active = $bindable(false), count = $bindable(0), loading = false, disabled = false, onToggle }: Props = $props();

	const isLike = $derived(kind === 'like');
</script>

<Button variant="ghost" size="sm" onclick={onToggle} disabled={loading || disabled}>
	<svg class="size-3.5 {isLike ? '' : 'rotate-180'}" viewBox="0 0 24 24" fill={active ? 'currentColor' : 'none'} stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 10v12"/><path d="M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2h0a3.13 3.13 0 0 1 3 3.88Z"/></svg>
	<span class="ml-1 {loading ? 'opacity-50' : ''}">{count}</span>
</Button>
