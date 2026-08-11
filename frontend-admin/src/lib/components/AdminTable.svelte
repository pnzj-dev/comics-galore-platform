<script lang="ts">
	import { createTable, FlexRender } from '@tanstack/svelte-table';
	import {
		tableFeatures,
		rowSortingFeature,
		columnFilteringFeature,
		globalFilteringFeature,
		rowPaginationFeature,
		createColumnHelper,
	} from '@tanstack/svelte-table';
	import { Button } from '$lib/components/ui/button/index.js';
	import DebouncedInput from '$lib/components/DebouncedInput.svelte';

	type ColumnDefExt = {
		key: string;
		label: string;
		sortable?: boolean;
		filterable?: boolean;
		filterType?: 'text' | 'select';
		filterOptions?: { value: string; label: string }[];
		filterPlaceholder?: string;
	};

	interface Props {
		columns: ColumnDefExt[];
		data: Record<string, unknown>[];
		total: number;
		page: number;
		limit: number;
		sortKey: string | null;
		sortDir: 'asc' | 'desc';
		search: string;
		filters: Record<string, string>;
		emptyMessage?: string;
		onSort: (key: string, dir: 'asc' | 'desc') => void;
		onSearch: (value: string) => void;
		onFilter: (key: string, value: string) => void;
		onPage: (page: number) => void;
		children: import('svelte').Snippet<[item: Record<string, unknown>, col: ColumnDefExt]>;
		actions: import('svelte').Snippet<[item: Record<string, unknown>]>;
	}

	let {
		columns: columnDefs,
		data,
		total,
		page,
		limit,
		sortKey,
		sortDir,
		search,
		filters,
		emptyMessage = 'No results found.',
		onSort,
		onSearch,
		onFilter,
		onPage,
		children,
		actions,
	}: Props = $props();

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const tableColumns = $derived(
		columnDefs.map((col) =>
			columnHelper.accessor(col.key, {
				header: col.label,
				enableSorting: col.sortable ?? false,
				enableColumnFilter: col.filterable ?? false,
				meta: {
					filterType: col.filterType,
					filterOptions: col.filterOptions,
					filterPlaceholder: col.filterPlaceholder ?? col.label,
				},
			})
		)
	);

	const sortingState = $derived(
		sortKey ? [{ id: sortKey, desc: sortDir === 'desc' }] : []
	);

	const table = createTable({
		features: tableFeatures({
			rowSortingFeature,
			columnFilteringFeature,
			globalFilteringFeature,
			rowPaginationFeature,
		}),
		// svelte-ignore state_referenced_locally
		columns: tableColumns,
		get data() { return data; },
		get rowCount() { return total; },
		manualPagination: true,
		manualSorting: true,
		manualFiltering: true,
		globalFilterFn: 'includesString',
		state: {
			get globalFilter() { return search; },
			get sorting() { return sortingState; },
			get columnFilters() {
				return Object.entries(filters)
					.filter(([_, v]) => v)
					.map(([id, value]) => ({ id, value }));
			},
			get pagination() { return { pageIndex: page - 1, pageSize: limit }; },
		},
		onSortingChange: (updater) => {
			const prev = table.getState().sorting;
			const next = typeof updater === 'function' ? updater(prev) : updater;
			if (next.length > 0) {
				onSort(next[0].id, next[0].desc ? 'desc' : 'asc');
			}
		},
		onGlobalFilterChange: (updater) => {
			const next = typeof updater === 'function'
				? updater(table.getState().globalFilter)
				: updater;
			onSearch(String(next ?? ''));
		},
		onColumnFiltersChange: (updater) => {
			const prev = table.getState().columnFilters;
			const next = typeof updater === 'function' ? updater(prev) : updater;
			const prevIds = new Map(prev.map(f => [f.id as string, f.value]));
			const nextIds = new Map(next.map(f => [f.id as string, f.value]));
			for (const [id, value] of nextIds) {
				if (prevIds.get(id) !== value) onFilter(id, String(value ?? ''));
			}
			for (const [id] of prevIds) {
				if (!nextIds.has(id)) onFilter(id, '');
			}
		},
	});
</script>

<div class="space-y-3">
	<div class="flex items-center gap-3">
		<DebouncedInput
			type="text"
			value={search}
			onchange={(value) => table.setGlobalFilter(String(value))}
			debounce={300}
			placeholder="Search..."
			class="max-w-xs rounded-md border border-input bg-background px-3 py-1.5 text-sm"
		/>
	</div>

	<div class="rounded-xl border border-border overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
					<tr class="border-b bg-muted/50 text-left">
						{#each headerGroup.headers as header (header.id)}
							<th class="px-4 py-2.5 font-medium whitespace-nowrap align-top">
								<div>
									{#if header.column.getCanSort()}
										<button
											onclick={header.column.getToggleSortingHandler()}
											class="hover:text-primary transition-colors cursor-pointer border-0 bg-transparent p-0 text-inherit select-none"
										>
											<FlexRender
												header={header}
												content={header.column.columnDef.header}
											/>
											{#if header.column.getIsSorted() === 'asc'} ↑
											{:else if header.column.getIsSorted() === 'desc'} ↓
											{:else} ↕
											{/if}
										</button>
									{:else}
										<FlexRender
											header={header}
											content={header.column.columnDef.header}
										/>
									{/if}
								</div>
								{#if header.column.getCanFilter()}
									{@const meta = ((header.column.columnDef.meta ?? {}) as Record<string, unknown>)}
									{#if (meta.filterType as string) === 'select' && (meta.filterOptions as any[])}
										<select
											value={header.column.getFilterValue() as string ?? ''}
											onchange={(e) => header.column.setFilterValue((e.target as HTMLSelectElement).value || undefined)}
											class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
										>
											<option value="">All</option>
											{#each (meta.filterOptions as any[]) as opt}
												<option value={opt.value}>{opt.label}</option>
											{/each}
										</select>
									{:else}
										<DebouncedInput
											type="text"
											value={(header.column.getFilterValue() as string ?? '')}
											onchange={(value) => header.column.setFilterValue(String(value) || undefined)}
											debounce={300}
											placeholder={(meta.filterPlaceholder as string) ?? ''}
											class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
										/>
									{/if}
								{/if}
							</th>
						{/each}
						<th class="px-4 py-2.5 w-24"></th>
					</tr>
				{/each}
			</thead>
			<tbody class="divide-y divide-border">
				{#if data.length === 0}
					<tr>
						<td colspan={columnDefs.length + 1} class="px-4 py-8 text-center text-sm text-muted-foreground">
							{emptyMessage}
						</td>
					</tr>
				{:else}
					{#each table.getRowModel().rows as row (row.id)}
						<tr class="hover:bg-muted/30">
							{#each tableColumns as _, i}
								{@const col = columnDefs[i] ?? columnDefs[0]}
								<td class="px-4 py-2.5">
									{#if children}
										{@render children(row.original, col)}
									{:else}
										<span class="text-xs">{row.original[col.key] as string}</span>
									{/if}
								</td>
							{/each}
							<td class="px-4 py-2.5">
								{#if actions}
									<div class="flex items-center gap-1">
										{@render actions(row.original)}
									</div>
								{/if}
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<div class="flex items-center justify-between text-sm text-muted-foreground">
		<span>{total} result{total !== 1 ? 's' : ''}</span>
		<div class="flex items-center gap-2">
			<Button variant="outline" size="sm" disabled={!table.getCanPreviousPage()} onclick={() => table.previousPage()}>
				← Prev
			</Button>
			<span class="text-xs">Page {page} of {table.getPageCount()}</span>
			<Button variant="outline" size="sm" disabled={!table.getCanNextPage()} onclick={() => table.nextPage()}>
				Next →
			</Button>
		</div>
	</div>
</div>
