<script lang="ts">
	import type { ColumnDef, TableFeatures } from '@tanstack/svelte-table';
	import { createTable, FlexRender } from '@tanstack/svelte-table';
	import { Button } from '$lib/components/ui/button/index.js';
	import DebouncedInput from '$lib/components/DebouncedInput.svelte';
	import ColumnFilter from '$lib/components/ColumnFilter.svelte';
	import { features } from '$lib/components/tableFeatures.svelte';

	interface Props {
		columns: ColumnDef<TableFeatures, unknown>[];
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
		actions: import('svelte').Snippet<[item: Record<string, unknown>]>;
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
		filters,
		emptyMessage = 'No results found.',
		onSort,
		onSearch,
		onFilter,
		onPage,
		actions,
	}: Props = $props();

	const sortingState = $derived(
		sortKey ? [{ id: sortKey, desc: sortDir === 'desc' }] : []
	);

	const table = createTable({
		features,
		// svelte-ignore state_referenced_locally
		columns,
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
							<th class="px-4 py-2.5 font-medium whitespace-nowrap align-top" colSpan={header.colSpan}>
								{#if !header.isPlaceholder}
									<div>
										<FlexRender header={header} />
									</div>
									{#if header.column.getCanFilter()}
										<div>
											<ColumnFilter column={header.column} />
										</div>
									{/if}
								{/if}
							</th>
						{/each}
						{#if actions}
							<th class="px-4 py-2.5 w-24"></th>
						{/if}
					</tr>
				{/each}
			</thead>
			<tbody class="divide-y divide-border">
				{#if data.length === 0}
					<tr>
						<td colspan={columns.length + (actions ? 1 : 0)} class="px-4 py-8 text-center text-sm text-muted-foreground">
							{emptyMessage}
						</td>
					</tr>
				{:else}
					{#each table.getRowModel().rows as row (row.id)}
						<tr class="hover:bg-muted/30">
							{#each row.getAllCells() as cell (cell.id)}
								<td class="px-4 py-2.5">
									<FlexRender cell={cell} />
								</td>
							{/each}
							{#if actions}
								<td class="px-4 py-2.5">
									<div class="flex items-center gap-1">
										{@render actions(row.original)}
									</div>
								</td>
							{/if}
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
