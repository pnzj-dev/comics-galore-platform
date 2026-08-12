<script lang="ts">
	import TableFilter from './TableFilter.svelte';

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
		columns,
		sortKey,
		sortDir,
		filters,
		onSort,
		onFilter,
		showFilters = false,
		onToggleFilters,
		headerClass = 'px-4 py-2.5 font-medium whitespace-nowrap align-top',
		sortBtnClass = 'hover:text-primary transition-colors cursor-pointer border-0 bg-transparent p-0 text-inherit select-none',
		headerRowClass = 'border-b bg-muted/50 text-left',
		filterRowClass = 'border-b bg-muted/30',
		filterCellClass = 'px-3 py-1.5',
		toggleBtnClass = 'text-xs text-muted-foreground hover:text-primary border-0 bg-transparent cursor-pointer',
	}: {
		columns: ColumnDef[];
		sortKey: string | null;
		sortDir: 'asc' | 'desc';
		filters: Record<string, string>;
		onSort: (key: string, dir: 'asc' | 'desc') => void;
		onFilter: (key: string, value: string) => void;
		showFilters?: boolean;
		onToggleFilters?: () => void;
		headerClass?: string;
		sortBtnClass?: string;
		headerRowClass?: string;
		filterRowClass?: string;
		filterCellClass?: string;
		toggleBtnClass?: string;
	} = $props();

	const hasFilters = $derived(columns.some((c) => c.filterable));

	function sortArrow(key: string): string {
		if (sortKey !== key) return '↕';
		return sortDir === 'asc' ? '↑' : '↓';
	}

	function toggleSort(col: ColumnDef) {
		if (!col.sortable) return;
		const dir: 'asc' | 'desc' = sortKey === col.key && sortDir === 'asc' ? 'desc' : 'asc';
		onSort(col.key, dir);
	}
</script>

<tr class={headerRowClass}>
	{#each columns as col}
		<th class={headerClass}>
			{#if col.sortable}
				<button class={sortBtnClass} onclick={() => toggleSort(col)}>
					{col.label}{sortArrow(col.key)}
				</button>
			{:else}
				{col.label}
			{/if}
		</th>
	{/each}
	<th class="px-4 py-2.5 w-24">
		{#if hasFilters}
			<button class={toggleBtnClass} onclick={() => onToggleFilters?.()}>
				{showFilters ? 'Hide' : 'Filter'}
			</button>
		{/if}
	</th>
</tr>
{#if showFilters && hasFilters}
	<tr class={filterRowClass}>
		{#each columns as col}
			<td class={filterCellClass}>
				{#if col.filterable}
					<TableFilter column={col} value={filters[col.key] || ''} {onFilter} />
				{/if}
			</td>
		{/each}
		<td class={filterCellClass}></td>
	</tr>
{/if}
