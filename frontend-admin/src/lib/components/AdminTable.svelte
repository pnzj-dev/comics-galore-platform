<script lang="ts">
	import { createTable, FlexRender } from '@tanstack/svelte-table';
	import {
		tableFeatures,
		rowSortingFeature,
		rowPaginationFeature,
		createColumnHelper,
	} from '@tanstack/svelte-table';
	import { Button } from '$lib/components/ui/button/index.js';

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
		loading: boolean;
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
		loading,
		emptyMessage = 'No results found.',
		onSort,
		onSearch,
		onFilter,
		onPage,
		children,
		actions,
	}: Props = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
	let searchTerm = $state(search);
	let searchTimer = $state<ReturnType<typeof setTimeout>>();
	let filterTimers = $state<Record<string, ReturnType<typeof setTimeout>>>({});

	function handleSearchInput(value: string) {
		searchTerm = value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => onSearch(value), 300);
	}

	function handleFilterInput(key: string, value: string) {
		if (filterTimers[key]) clearTimeout(filterTimers[key]);
		filterTimers[key] = setTimeout(() => onFilter(key, value), 300);
	}

	const columnHelper = createColumnHelper<Record<string, unknown>>();

	const tableColumns = $derived(
		columnDefs.map((col) =>
			columnHelper.accessor(col.key, {
				header: col.label,
				enableSorting: col.sortable ?? false,
			})
		)
	);

	const sortingState = $derived(
		sortKey ? [{ id: sortKey, desc: sortDir === 'desc' }] : []
	);

	const table = createTable({
		features: tableFeatures({
			rowSortingFeature,
			rowPaginationFeature,
		}),
		columns: tableColumns,
		get data() { return data; },
		get rowCount() { return total; },
		manualPagination: true,
		manualSorting: true,
		state: {
			get sorting() { return sortingState as any; },
			get pagination() { return { pageIndex: page - 1, pageSize: limit }; },
		},
		onSortingChange: (updater) => {
			const prev = table.getState().sorting;
			const next = typeof updater === 'function' ? updater(prev) : updater;
			if (next.length > 0) {
				onSort(next[0].id, next[0].desc ? 'desc' : 'asc');
			}
		},
	});
</script>

<div class="space-y-3">
	<div class="flex items-center gap-3">
		<input
			type="text"
			placeholder="Search..."
			value={searchTerm}
			oninput={(e) => handleSearchInput((e.target as HTMLInputElement).value)}
			class="max-w-xs rounded-md border border-input bg-background px-3 py-1.5 text-sm"
		/>
	</div>

	<div class="rounded-xl border border-border overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
					<tr class="border-b bg-muted/50 text-left">
						{#each headerGroup.headers as header (header.id)}
							<th class="px-4 py-2.5 font-medium whitespace-nowrap">
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
							</th>
						{/each}
						<th class="px-4 py-2.5 w-24"></th>
					</tr>
				{/each}
				<tr class="border-b bg-muted/30">
					{#each columnDefs as col}
						<td class="px-3 py-1.5">
							{#if col.filterable && col.filterType === 'text'}
								<input
									type="text"
									placeholder={col.filterPlaceholder || col.label}
									value={filters[col.key] || ''}
									oninput={(e) => handleFilterInput(col.key, (e.target as HTMLInputElement).value)}
									class="w-full rounded border border-input bg-background px-2 py-1 text-xs"
								/>
							{:else if col.filterable && col.filterType === 'select' && col.filterOptions}
								<select
									value={filters[col.key] || ''}
									onchange={(e) => onFilter(col.key, (e.target as HTMLSelectElement).value)}
									class="w-full rounded border border-input bg-background px-2 py-1 text-xs"
								>
									<option value="">All</option>
									{#each col.filterOptions as opt}
										<option value={opt.value}>{opt.label}</option>
									{/each}
								</select>
							{/if}
						</td>
					{/each}
					<td class="px-3 py-1.5"></td>
				</tr>
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
							{#each tableColumns as colDef, i}
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
			<Button variant="outline" size="sm" disabled={page <= 1} onclick={() => onPage(page - 1)}>
				← Prev
			</Button>
			<span class="text-xs">Page {page} of {totalPages}</span>
			<Button variant="outline" size="sm" disabled={page >= totalPages} onclick={() => onPage(page + 1)}>
				Next →
			</Button>
		</div>
	</div>
</div>
