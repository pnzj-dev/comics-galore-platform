<script lang="ts">
	let { label, value }: { label: string; value: unknown } = $props();
</script>

<div class="flex justify-between gap-4 py-1.5 border-b border-border/50 last:border-0">
	<span class="text-sm text-muted-foreground flex-shrink-0">{label}</span>
	<span class="text-sm font-medium text-right break-all">{value ?? '—'}</span>
</div>
