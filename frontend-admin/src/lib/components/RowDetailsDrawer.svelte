<script lang="ts">
	let {
		open,
		title = 'Details',
		onClose,
		children,
	}: {
		open: boolean;
		title?: string;
		onClose: () => void;
		children: import('svelte').Snippet;
	} = $props();

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50" role="dialog" aria-modal="true" aria-label={title}>
		<div class="absolute inset-0 bg-black/40" onclick={onClose} aria-hidden="true"></div>
		<div class="absolute inset-y-0 right-0 w-full max-w-md bg-background shadow-xl flex flex-col" role="presentation">
			<div class="flex items-center justify-between p-4 border-b flex-shrink-0">
				<h2 class="text-lg font-semibold">{title}</h2>
				<button onclick={onClose} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-6">
				{@render children()}
			</div>
		</div>
	</div>
{/if}
