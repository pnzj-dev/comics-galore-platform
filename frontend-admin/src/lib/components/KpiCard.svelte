<script lang="ts">
	import type { Component } from 'svelte';

	let {
		title,
		value,
		icon: Icon,
		hint,
		accent = false,
	}: {
		title: string;
		value: string;
		icon?: Component;
		hint?: string;
		accent?: boolean;
	} = $props();
</script>

<div class="rounded-xl border border-border bg-background p-4 flex flex-col gap-1 {accent ? 'border-primary/30 bg-primary/5' : ''}">
	<div class="flex items-center justify-between">
		<span class="text-xs font-medium text-muted-foreground">{title}</span>
		{#if Icon}
			<Icon class="size-4 text-muted-foreground" />
		{/if}
	</div>
	<span class="text-2xl font-bold">{value}</span>
	{#if hint}
		<span class="text-xs text-muted-foreground">{hint}</span>
	{/if}
</div>
