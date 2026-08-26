<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		page: number;
		totalPages: number;
		onPage: (p: number) => void;
	}

	let { page, totalPages, onPage }: Props = $props();

	function pageNumbers(): (number | '…')[] {
		const nums: (number | '…')[] = [];
		const start = Math.max(1, page - 2);
		const end = Math.min(totalPages, page + 2);
		if (start > 1) nums.push(1);
		if (start > 2) nums.push('…');
		for (let i = start; i <= end; i++) nums.push(i);
		if (end < totalPages - 1) nums.push('…');
		if (end < totalPages) nums.push(totalPages);
		return nums;
	}
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-center gap-1 mt-12 pt-6 border-t border-border">
		<Button variant="outline" size="sm" disabled={page <= 1} onclick={() => onPage(page - 1)}>← Prev</Button>
		{#each pageNumbers() as p (p)}
			{#if p === '…'}
				<span class="px-2 text-sm text-muted-foreground">…</span>
			{:else}
				<Button
					variant={p === page ? 'default' : 'outline'}
					size="sm"
					onclick={() => onPage(p)}
				>
					{p}
				</Button>
			{/if}
		{/each}
		<Button variant="outline" size="sm" disabled={page >= totalPages} onclick={() => onPage(page + 1)}>Next →</Button>
	</div>
{/if}
