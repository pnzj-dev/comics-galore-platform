<script lang="ts" generics="TData">
	import TableHeader from './TableHeader.svelte';
	import TablePagination from './TablePagination.svelte';
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

	interface Props {
		columns: ColumnDef[];
		data: TData[];
		total: number;
		page: number;
		limit: number;
		sortKey: string | null;
		sortDir: 'asc' | 'desc';
		search: string;
		searchColumns?: string[];
		filters: Record<string, string>;
		emptyMessage?: string;
		onSort: (key: string, dir: 'asc' | 'desc') => void;
		onSearch: (value: string) => void;
		onFilter: (key: string, value: string) => void;
		onPage: (page: number) => void;
		children: import('svelte').Snippet<[item: TData, col: ColumnDef]>;
		actions: import('svelte').Snippet<[item: TData]>;
		searchPlaceholder?: string;
		searchClass?: string;
		containerClass?: string;
		tableClass?: string;
		tableWrapperClass?: string;
		headerClass?: string;
		sortBtnClass?: string;
		headerRowClass?: string;
		filterRowClass?: string;
		filterCellClass?: string;
		rowClass?: string;
		cellClass?: string;
		actionsCellClass?: string;
		searchWrapperClass?: string;
	}

	let {
		columns,
		data,
		total,
		page,
		limit,
		sortKey,
		sortDir,
		search,
		searchColumns = columns.filter((c) => !c.filterType || c.filterType === 'text').map((c) => c.key),
		filters,
		emptyMessage = 'No results found.',
		onSort,
		onSearch,
		onFilter,
		onPage,
		children,
		actions,
		searchPlaceholder = 'Search...',
		searchClass = 'max-w-xs rounded-md border border-input bg-background px-3 py-1.5 text-sm',
		containerClass = 'space-y-3',
		tableClass = 'w-full text-sm',
		tableWrapperClass = 'rounded-xl border border-border overflow-hidden',
		headerClass = 'px-4 py-2.5 font-medium whitespace-nowrap align-top',
		sortBtnClass = 'hover:text-primary transition-colors cursor-pointer border-0 bg-transparent p-0 text-inherit select-none',
		headerRowClass = 'border-b bg-muted/50 text-left',
		filterRowClass = 'border-b bg-muted/30',
		filterCellClass = 'px-3 py-1.5',
		rowClass = 'hover:bg-muted/30',
		cellClass = 'px-4 py-2.5',
		actionsCellClass = 'px-4 py-2.5',
		searchWrapperClass = 'flex items-center gap-3',
	}: Props = $props();

	const searchableCols = $derived(searchColumns.length > 0 ? searchColumns : columns.map((c) => c.key));

	function handleSearch(value: string) {
		onSearch(value);
	}
</script>

<div class={containerClass}>
	<div class={searchWrapperClass}>
		<DebouncedInput
			type="text"
			value={search}
			onchange={handleSearch}
			debounce={300}
			placeholder={searchPlaceholder}
			class={searchClass}
		/>
	</div>

	<div class={tableWrapperClass}>
		<table class={tableClass}>
			<thead>
				<TableHeader
					{columns}
					{sortKey}
					{sortDir}
					{filters}
					{onSort}
					{onFilter}
					{headerClass}
					{sortBtnClass}
					{headerRowClass}
					{filterRowClass}
					{filterCellClass}
				/>
			</thead>
			<tbody class="divide-y divide-border">
				{#if data.length === 0}
					<tr>
						<td colspan={columns.length + 1} class="px-4 py-8 text-center text-sm text-muted-foreground">
							{emptyMessage}
						</td>
					</tr>
				{:else}
					{#each data as row (row.id as string)}
						<tr class={rowClass}>
							{#each columns as col}
								<td class={cellClass}>
									{#if children}
										{@render children(row, col)}
									{:else}
										<span class="text-xs">{(row as Record<string, unknown>)[col.key] as string}</span>
									{/if}
								</td>
							{/each}
							{#if actions}
								<td class={actionsCellClass}>
									<div class="flex items-center gap-1">
										{@render actions(row)}
									</div>
								</td>
							{/if}
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<TablePagination {page} {limit} {total} {onPage} />
</div>
