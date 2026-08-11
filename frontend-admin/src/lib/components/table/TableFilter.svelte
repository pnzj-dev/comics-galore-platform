<script lang="ts">
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
		<input
			type="text"
			defaultValue={value}
			placeholder={col.filterPlaceholder ?? col.label}
			class={inputClass}
			onblur={(e) => onFilter(col.key, (e.target as HTMLInputElement).value)}
			onkeydown={(e) => { if (e.key === 'Enter') onFilter(col.key, (e.currentTarget as HTMLInputElement).value); }}
		/>
	{/if}
</div>

