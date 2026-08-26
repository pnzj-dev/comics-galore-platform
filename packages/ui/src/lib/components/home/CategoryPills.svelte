<!--
  CategoryPills.svelte
  Horizontal scrollable category filter pills.
-->
<script lang="ts">
	interface Props {
		categories: string[];
		activeId: string;
		onChange: (id: string) => void;
		class?: string;
	}

	let { categories = [], activeId = '', onChange, class: className = '' }: Props = $props();
</script>

<div
	class="flex gap-2 overflow-x-auto pb-1 scrollbar-none {className}"
	role="tablist"
	aria-label="Series categories"
>
	<button
		type="button"
		role="tab"
		aria-selected={activeId === ''}
		onclick={() => onChange('')}
		class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition
			focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-black
			{activeId === ''
				? 'bg-primary text-primary-foreground'
				: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
	>
		All
	</button>

	{#each categories as cat (cat)}
		<button
			type="button"
			role="tab"
			aria-selected={activeId === cat}
			onclick={() => onChange(cat)}
			class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition
				focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-black
				{activeId === cat
					? 'bg-primary text-primary-foreground'
					: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
		>
			{cat}
		</button>
	{/each}
</div>

<style>
	.scrollbar-none {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
	.scrollbar-none::-webkit-scrollbar {
		display: none;
	}
</style>
