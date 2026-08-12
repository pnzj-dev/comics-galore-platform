<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	let {
		open,
		deposit,
		onClose,
	}: {
		open: boolean;
		deposit: Record<string, unknown> | null;
		onClose: () => void;
	} = $props();

	let copied = $state(false);

	const payAddress = $derived((deposit?.pay_address as string) || '');
	const qrUrl = $derived(
		payAddress ? `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(payAddress)}` : ''
	);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}

	async function copyAddress() {
		if (!payAddress) return;
		try {
			await navigator.clipboard.writeText(payAddress);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open && deposit}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={onClose} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-md" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Deposit Details</h2>
				<button onclick={onClose} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">✕</button>
			</div>

			<div class="p-6 space-y-4">
				<div class="text-center space-y-3">
					{#if qrUrl}
						<img src={qrUrl} alt="Payment QR code" class="mx-auto size-48 rounded-lg border border-border" />
					{/if}

					<div>
						<p class="text-xs text-muted-foreground mb-1">Pay to this address</p>
						<div class="flex items-center gap-2">
							<code class="flex-1 text-xs font-mono break-all bg-muted/50 rounded px-2 py-1.5 text-left">{payAddress || '—'}</code>
							<Button size="sm" variant="outline" onclick={copyAddress}>{copied ? 'Copied!' : 'Copy'}</Button>
						</div>
					</div>
				</div>

				<div class="text-xs space-y-1.5 border-t pt-3">
					<div class="flex justify-between"><span class="text-muted-foreground">Status</span><span>{deposit.status as string}</span></div>
					<div class="flex justify-between"><span class="text-muted-foreground">Currency</span><span class="uppercase">{deposit.currency_crypto as string}</span></div>
					<div class="flex justify-between"><span class="text-muted-foreground">Amount (crypto)</span><span class="font-mono">{deposit.amount_crypto as string}</span></div>
					<div class="flex justify-between"><span class="text-muted-foreground">Provider ID</span><span class="font-mono">{(deposit.provider_deposit_id as string) || '—'}</span></div>
				</div>
			</div>
		</div>
	</div>
{/if}
