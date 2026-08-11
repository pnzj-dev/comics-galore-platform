<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	let {
		page,
		limit,
		total,
		onPage,
		containerClass = 'flex items-center justify-between text-sm text-muted-foreground',
		btnClass = '',
		infoClass = 'text-xs',
	}: {
		page: number;
		limit: number;
		total: number;
		onPage: (page: number) => void;
		containerClass?: string;
		btnClass?: string;
		infoClass?: string;
	} = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
	const canPrev = $derived(page > 1);
	const canNext = $derived(page < totalPages);

	function goPrev() { if (canPrev) onPage(page - 1); }
	function goNext() { if (canNext) onPage(page + 1); }
</script>

<div class={containerClass}>
	<span>{total} result{total !== 1 ? 's' : ''}</span>
	<div class="flex items-center gap-2">
		<Button variant="outline" size="sm" class={btnClass} disabled={!canPrev} onclick={goPrev}>
			← Prev
		</Button>
		<span class={infoClass}>Page {page} of {totalPages}</span>
		<Button variant="outline" size="sm" class={btnClass} disabled={!canNext} onclick={goNext}>
			Next →
		</Button>
	</div>
</div>
