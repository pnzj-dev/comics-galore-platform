<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { formatDate, formatUSD } from '$lib/utils/format';

	let { deposit }: { deposit: Record<string, unknown> } = $props();

	let copied = $state(false);

	const payAddress = $derived((deposit.pay_address as string) || '');
	const qrUrl = $derived(
		payAddress ? `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(payAddress)}` : ''
	);

	async function copyAddress() {
		if (!payAddress) return;
		try {
			await navigator.clipboard.writeText(payAddress);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {}
	}
</script>

<div class="space-y-4">
	{#if qrUrl}
		<div class="text-center">
			<img src={qrUrl} alt="Payment QR code" class="mx-auto size-48 rounded-lg border border-border" />
		</div>
	{/if}

	<div>
		<p class="text-xs text-muted-foreground mb-1">Pay to this address</p>
		<div class="flex items-center gap-2">
			<code class="flex-1 text-xs font-mono break-all bg-muted/50 rounded px-2 py-1.5 text-left">{payAddress || '—'}</code>
			<Button size="sm" variant="outline" onclick={copyAddress}>{copied ? 'Copied!' : 'Copy'}</Button>
		</div>
	</div>

	<div class="text-sm space-y-2 border-t pt-3">
		<div class="flex justify-between"><span class="text-muted-foreground">Status</span><span class="capitalize">{deposit.status as string}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Currency</span><span class="uppercase">{deposit.currency_crypto as string}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Amount (crypto)</span><span class="font-mono">{deposit.amount_crypto as string || '—'}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Amount (USD)</span><span>{formatUSD(deposit.amount_usd_cents as number)}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Provider ID</span><span class="font-mono text-xs">{deposit.provider_deposit_id as string || '—'}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Created</span><span>{formatDate(deposit.created_at as string)}</span></div>
		<div class="flex justify-between"><span class="text-muted-foreground">Completed</span><span>{formatDate(deposit.completed_at as string)}</span></div>
	</div>
</div>
