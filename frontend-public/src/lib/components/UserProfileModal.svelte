<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import CheckoutModal from '$lib/components/CheckoutModal.svelte';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();
	let checkoutOpen = $state(false);

	const user = $derived($currentUser);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={onClose} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-sm max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Profile</h2>
				<button onclick={onClose} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-4">
				<div class="flex items-center gap-3">
					<div class="size-11 rounded-full bg-purple-100 dark:bg-purple-900 flex items-center justify-center flex-shrink-0">
						<span class="text-lg font-semibold text-purple-600 dark:text-purple-300">{user?.email?.charAt(0)?.toUpperCase() || '?'}</span>
					</div>
					<div class="min-w-0">
						<p class="text-sm font-medium truncate">{user?.email || '—'}</p>
						<p class="text-xs text-muted-foreground">Member since {formatDate(user?.created_at || '')}</p>
					</div>
				</div>

			<div class="flex gap-2">
				{#if user?.role !== 'admin'}
					<span class="text-xs rounded-full px-2 py-0.5 capitalize bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Tier: {user?.tier || 'free'}</span>
				{/if}
				<span class="text-xs rounded-full px-2 py-0.5 bg-primary/10 text-primary">Role: {user?.role || 'user'}</span>
			</div>

			<div class="rounded-xl border border-border p-3 space-y-2">
				<p class="text-xs font-medium">Subscription</p>
				<p class="text-xs text-muted-foreground">{user?.tier === 'free' ? 'Free plan · No active subscription' : `Active ${user?.tier} plan`}</p>
				<Button size="sm" class="w-full mt-1" onclick={() => checkoutOpen = true}>
					{user?.tier === 'free' ? 'Upgrade Plan' : 'Manage Subscription'}
				</Button>
			</div>
		</div>
		</div>
	</div>
{/if}

<CheckoutModal open={checkoutOpen} onClose={() => checkoutOpen = false} />
