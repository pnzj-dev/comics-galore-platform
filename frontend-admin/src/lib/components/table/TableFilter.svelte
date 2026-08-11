<script lang="ts">
	import DebouncedInput from './DebouncedInput.svelte';

	type ColumnDef = {
		key: string;
		label: string;
		sortable?: boolean;
		filterable?: boolean;
		filterType?: 'text' | 'select';
		filterOptions?: { value: string; label: string }[];
		filterPlaceholder?: string;
	};

	let {
		column: col,
		value,
		onFilter,
		class: className = '',
		inputClass = 'block w-full rounded border border-input bg-background px-2 py-1 text-xs',
	}: {
		column: ColumnDef;
		value: string;
		onFilter: (key: string, value: string) => void;
		class?: string;
		inputClass?: string;
	} = $props();
</script>

<div class={className || 'mt-1'}>
	{#if col.filterType === 'select' && col.filterOptions}
		<select
			value={value}
			onchange={(e) => onFilter(col.key, (e.target as HTMLSelectElement).value)}
			class={inputClass}
		>
			<option value="">All</option>
			{#each col.filterOptions as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	{:else}
		<DebouncedInput
			type="text"
			value={value}
			onchange={(v) => onFilter(col.key, v)}
			debounce={300}
			placeholder={col.filterPlaceholder ?? col.label}
			class={inputClass}
		/>
	{/if}
</div>
