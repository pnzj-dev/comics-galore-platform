<!--
  DayPills.svelte
  Day-of-week + Completed filter pills for the Daily section.
-->
<script lang="ts">
	interface Day {
		id: string;
		name: string;
	}

	interface Props {
		days: Day[];
		activeId: string;
		onChange: (id: string) => void;
		class?: string;
	}

	let { days = [], activeId = '', onChange, class: className = '' }: Props = $props();
</script>

<div
	class="flex gap-2 overflow-x-auto pb-1 scrollbar-none {className}"
	role="tablist"
	aria-label="Daily schedule"
>
	{#each days as day (day.id)}
		<button
			type="button"
			role="tab"
			aria-selected={activeId === day.id}
			onclick={() => onChange(day.id)}
			class="shrink-0 rounded-full px-4 py-2 text-sm font-medium transition
				focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-black
				{activeId === day.id
					? 'bg-black text-white dark:bg-white dark:text-black'
					: 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700'}"
		>
			{day.name}
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
