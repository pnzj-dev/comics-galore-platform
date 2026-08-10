<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	type Column = {
		key: string;
		label: string;
		sortable?: boolean;
		filterable?: boolean;
		filterType?: 'text' | 'select';
		filterOptions?: { value: string; label: string }[];
		filterPlaceholder?: string;
	};

	interface Props {
		columns: Column[];
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
		children: import('svelte').Snippet<[item: Record<string, unknown>, col: Column]>;
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
		loading,
		emptyMessage = 'No results found.',
		onSort,
		onSearch,
		onFilter,
		onPage,
		children,
		actions
	}: Props = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
	// svelte-ignore state_referenced_locally
	let searchTerm = $state(search);
	let searchTimer = $state<ReturnType<typeof setTimeout>>();

	function handleSearchInput(value: string) {
		searchTerm = value;
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => onSearch(value), 300);
	}

	function handleSort(key: string) {
		if (!columns.find(c => c.key === key)?.sortable) return;
		let nextDir: 'asc' | 'desc' = 'asc';
		if (sortKey === key) {
			nextDir = sortDir === 'asc' ? 'desc' : 'asc';
		}
		onSort(key, nextDir);
	}

	function sortArrow(key: string): string {
		if (sortKey !== key) return ' ↕';
		return sortDir === 'asc' ? ' ↑' : ' ↓';
	}
</script>

{#if loading}
	<div class="rounded-xl border border-border overflow-hidden">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b bg-muted/50 text-left">
					{#each columns as col}
						<th class="px-4 py-2.5 font-medium">
							<div class="h-3 bg-muted rounded w-16 animate-pulse"></div>
						</th>
					{/each}
					<th class="px-4 py-2.5 w-24"></th>
				</tr>
			</thead>
			<tbody class="divide-y divide-border">
				{#each Array(5) as _}
					<tr>
						{#each columns as _}
							<td class="px-4 py-2.5"><div class="h-3 bg-muted rounded w-full animate-pulse"></div></td>
						{/each}
						<td class="px-4 py-2.5"><div class="h-3 bg-muted rounded w-16 animate-pulse"></div></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
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
					<tr class="border-b bg-muted/50 text-left">
						{#each columns as col}
							<th class="px-4 py-2.5 font-medium whitespace-nowrap">
								<button
									onclick={() => handleSort(col.key)}
									class="hover:text-primary transition-colors cursor-pointer border-0 bg-transparent p-0 text-inherit select-none"
									disabled={!col.sortable}
								>
									{col.label}{sortArrow(col.key)}
								</button>
							</th>
						{/each}
						<th class="px-4 py-2.5 w-24"></th>
					</tr>
					<tr class="border-b bg-muted/30">
						{#each columns as col}
							<td class="px-3 py-1.5">
								{#if col.filterable && col.filterType === 'text'}
									<input
										type="text"
										placeholder={col.filterPlaceholder || col.label}
										value={filters[col.key] || ''}
										oninput={(e) => onFilter(col.key, (e.target as HTMLInputElement).value)}
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
							<td colspan={columns.length + 1} class="px-4 py-8 text-center text-sm text-muted-foreground">
								{emptyMessage}
							</td>
						</tr>
					{:else}
						{#each data as row (row.id as string)}
							<tr class="hover:bg-muted/30">
								{#each columns as col}
									<td class="px-4 py-2.5">
										{#if children}
											{@render children(row, col)}
										{:else}
											<span class="text-xs">{row[col.key] as string}</span>
										{/if}
									</td>
								{/each}
								<td class="px-4 py-2.5">
									{#if actions}
										<div class="flex items-center gap-1">
											{@render actions(row)}
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
{/if}
