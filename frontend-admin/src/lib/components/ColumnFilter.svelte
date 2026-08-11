<script lang="ts" generics="TData extends Record<string, unknown>">
  import DebouncedInput from "$lib/components/DebouncedInput.svelte";
  import type { Column, Table } from "@tanstack/svelte-table";
  import { features } from "$lib/components/tableFeatures.svelte";

  let {
    column,
    table,
  }: {
    column: Column<typeof features, TData, unknown>;
    table: Table<typeof features, TData>;
  } = $props();

  const meta = column.columnDef.meta ?? {};
  const filterType = meta.filterType ?? "text";
  const filterOptions = meta.filterOptions ?? [];
  const filterPlaceholder = meta.filterPlaceholder ?? "Filter...";

  const firstValue = $derived(
    table.getPreFilteredRowModel().flatRows[0]?.getValue(column.id),
  );

  const isNumber = $derived(typeof firstValue === "number");
</script>

{#if filterType === "select" && filterOptions.length > 0}
  <select
    value={(column.getFilterValue() as string) ?? ""}
    onchange={(e) =>
      column.setFilterValue((e.target as HTMLSelectElement).value || undefined)}
    class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
  >
    <option value="">All</option>
    {#each filterOptions as opt}
      <option value={opt.value}>{opt.label}</option>
    {/each}
  </select>
{:else}
  {#if isNumber}
    <div>
      <div class="filter-row">
        <DebouncedInput
          type="number"
          min={0}
          max={100}
          value={(
            column.getFilterValue() as [number, number] | undefined
          )?.[0] ?? ""}
          onchange={(value) =>
            column.setFilterValue((old: [number, number] | undefined) => [
              value,
              old?.[1] ?? "",
            ])}
          debounce={500}
          placeholder="Min"
          class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
        />
        <DebouncedInput
          type="number"
          min={0}
          max={100}
          value={(
            column.getFilterValue() as [number, number] | undefined
          )?.[1] ?? ""}
          onchange={(value) =>
            column.setFilterValue((old: [number, number] | undefined) => [
              old?.[0] ?? "",
              value,
            ])}
          debounce={500}
          placeholder="Max"
          class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
        />
      </div>
    </div>
  {:else}
    <div>
      <DebouncedInput
        type="text"
        value={(column.getFilterValue() ?? "") as string}
        onchange={(value) => column.setFilterValue(value)}
        debounce={500}
        placeholder={filterPlaceholder}
        class="mt-1 block w-full rounded border border-input bg-background px-2 py-1 text-xs"
      />
    </div>
  {/if}
{/if}
