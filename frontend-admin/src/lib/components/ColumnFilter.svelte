<script lang="ts" generics="TData extends Record<string, unknown>">
	import DebouncedInput from '$lib/components/DebouncedInput.svelte';
	import type { Column, TableFeatures } from '@tanstack/svelte-table';

	let {
		column,
	}: {
		column: Column<TableFeatures, TData, unknown>;
	} = $props();

	const meta = ((column.columnDef.meta ?? {}) as Record<string, unknown>);
	const filterType = (meta.filterType as string) ?? 'text';
	const filterOptions = (meta.filterOptions as Array<{ value: string; label: string }>) ?? [];
	const filterPlaceholder = (meta.filterPlaceholder as string) ?? 'Filter...';
</script>

{#if filterType === 'select' && filterOptions.length > 0}
	<select
		value={column.getFilterValue() as string ?? ''}
		onchange={(e) => column.setFilterValue((e.target as HTMLSelectElement).value || undefined)}
		class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
	>
		<option value="">All</option>
		{#each filterOptions as opt}
			<option value={opt.value}>{opt.label}</option>
		{/each}
	</select>
{:else}
	<DebouncedInput
		type="text"
		value={(column.getFilterValue() as string ?? '')}
		onchange={(value) => column.setFilterValue(String(value) || undefined)}
		debounce={300}
		placeholder={filterPlaceholder}
		class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
	/>
{/if}
