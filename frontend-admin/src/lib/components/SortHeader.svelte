<script lang="ts">
	import { FlexRender } from '@tanstack/svelte-table';
	import type { Header, Table, TableFeatures, RowData } from '@tanstack/svelte-table';

	type Props = {
		label?: string;
		header: Header<TableFeatures, unknown, unknown>;
	};

	let { label, header }: Props = $props();
</script>

<button
	onclick={header.column.getToggleSortingHandler()}
	class="hover:text-primary transition-colors cursor-pointer border-0 bg-transparent p-0 text-inherit select-none"
>
	{#if label}
		{label}
	{:else}
		<FlexRender header={header} content={header.column.columnDef.header} />
	{/if}
	{#if header.column.getIsSorted() === 'asc'} ↑
	{:else if header.column.getIsSorted() === 'desc'} ↓
	{:else} ↕
	{/if}
</button>
