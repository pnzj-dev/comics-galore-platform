<script lang="ts" generics="TData extends { id: string }">
	import TableHeader from './TableHeader.svelte';
	import TablePagination from './TablePagination.svelte';
	import DebouncedInput from './DebouncedInput.svelte';
	import RowDetailsDrawer from '$lib/components/RowDetailsDrawer.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Eye } from 'lucide-svelte';

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
		actions?: import('svelte').Snippet<[item: TData]>;
		details?: import('svelte').Snippet<[item: TData]>;
		detailsTitle?: string;
		showFilters?: boolean;
		onToggleFilters?: () => void;
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
		details,
		detailsTitle = 'Details',
		showFilters = false,
		onToggleFilters,
		searchPlaceholder = 'Search...',
		searchClass = 'max-w-xs rounded-md border border-input bg-background px-3 py-1.5 text-sm',
		containerClass = 'space-y-3',
		tableClass = 'w-full text-sm table-fixed',
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

	let selectedRow = $state<TData | null>(null);

	const hasActionsColumn = $derived(!!(actions || details));

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

	{#if columns.some((c) => c.filterable) && onToggleFilters}
		<div class="flex justify-end">
			<button
				class="text-xs font-medium text-primary/70 hover:text-primary border border-primary/30 rounded px-2.5 py-1 bg-transparent cursor-pointer"
				onclick={onToggleFilters}
			>
				{showFilters ? 'Hide Filters ↑' : 'Show Filters ↓'}
			</button>
		</div>
	{/if}

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
					{showFilters}
					{onToggleFilters}
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
						<td colspan={columns.length + (hasActionsColumn ? 1 : 0)} class="px-4 py-8 text-center text-sm text-muted-foreground">
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
							{#if hasActionsColumn}
								<td class={actionsCellClass}>
									<div class="flex items-center gap-1">
										{#if details}
											<Button size="sm" variant="ghost" onclick={() => selectedRow = row}>
												<Eye class="size-3.5 mr-1" /> Details
											</Button>
										{/if}
										{#if actions}
											{@render actions(row)}
										{/if}
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

{#if details}
	<RowDetailsDrawer open={selectedRow !== null} title={detailsTitle} onClose={() => selectedRow = null}>
		{#if selectedRow}
			{@render details(selectedRow)}
		{/if}
	</RowDetailsDrawer>
{/if}
