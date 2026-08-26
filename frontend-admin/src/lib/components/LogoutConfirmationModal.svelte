<script lang="ts">
	import { logout } from '$lib/stores/auth';
	import { modal } from '$lib/stores/modal.svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('logout'));

	function close() {
		modal.close('logout');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	function handleLogout() {
		logout('/login');
		close();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-sm" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Log out</h2>
				<button onclick={close} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>
			<div class="p-6 space-y-4">
				<p class="text-sm text-muted-foreground">Are you sure you want to log out?</p>
				<div class="flex gap-2 justify-end">
					<Button variant="outline" size="sm" onclick={close}>Cancel</Button>
					<Button variant="destructive" size="sm" onclick={handleLogout}>Yes, log out</Button>
				</div>
			</div>
		</div>
	</div>
{/if}
