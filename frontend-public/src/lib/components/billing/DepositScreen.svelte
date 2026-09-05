<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { depositId, payAddress, payAmount, payCurrency, planId, crypto, payinExtraId, network, qrDataUrl, onSuccess, onTimeout, onRetry }: {
		depositId: string;
		payAddress: string;
		payAmount: number;
		payCurrency: string;
		planId: string;
		crypto: string;
		payinExtraId?: string;
		network?: string;
		qrDataUrl?: string;
		onSuccess: () => void;
		onTimeout: () => void;
		onRetry: () => void;
	} = $props();

	const MAX_SECONDS = 1800; // 30 minutes
	const POLL_MS = 5000;

	const NETWORK_NAMES: Record<string, string> = {
		btc: 'Bitcoin',
		ltc: 'Litecoin',
		eth: 'Ethereum',
		trx: 'Tron',
		sol: 'Solana',
		xrp: 'Ripple',
		xlm: 'Stellar',
		bch: 'Bitcoin Cash',
		doge: 'Dogecoin',
		bnb: 'BNB Chain'
	};

	let copied = $state(false);
	let copiedTag = $state(false);
	let remaining = $state(MAX_SECONDS);
	let timedOut = $state(false);

	let pollInterval: ReturnType<typeof setInterval>;
	let tickInterval: ReturnType<typeof setInterval>;

	onMount(() => {
		tickInterval = setInterval(() => {
			remaining = Math.max(0, remaining - 1);
			if (remaining <= 0) {
				clearInterval(tickInterval);
				clearInterval(pollInterval);
				timedOut = true;
				onTimeout?.();
			}
		}, 1000);

		pollInterval = setInterval(async () => {
			try {
				const res = await encore.billing.PollDeposit(depositId);
				if (res.completed) {
					clearInterval(pollInterval);
					clearInterval(tickInterval);
					onSuccess();
				}
			} catch {
				// keep polling
			}
		}, POLL_MS);

		return () => {
			clearInterval(pollInterval);
			clearInterval(tickInterval);
		};
	});

	function copyAddress() {
		navigator.clipboard.writeText(payAddress);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function copyTag() {
		if (!payinExtraId) return;
		navigator.clipboard.writeText(payinExtraId);
		copiedTag = true;
		setTimeout(() => (copiedTag = false), 2000);
	}

	function formatAmount(n: number): string {
		if (!n) return '0';
		return n.toFixed(8).replace(/\.?0+$/, '');
	}

	function formatCountdown(s: number): string {
		const m = Math.floor(s / 60);
		const sec = s % 60;
		return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
	}

	const networkLabel = $derived(network ? (NETWORK_NAMES[network.toLowerCase()] || network.toUpperCase()) : payCurrency.toUpperCase());
	const progressPct = $derived(((MAX_SECONDS - remaining) / MAX_SECONDS) * 100);
</script>

<div class="space-y-6">
	<h3 class="text-lg font-semibold">Send Payment</h3>

	<div class="grid gap-6 md:grid-cols-[auto_1fr] md:items-start">
		<!-- QR + amount -->
		<div class="flex flex-col items-center gap-4">
			{#if qrDataUrl}
				<div class="bg-white p-3 rounded-xl inline-block">
					<img src={qrDataUrl} alt="Payment QR Code" class="w-44 h-44 rounded-lg" />
				</div>
			{/if}
			<div class="text-center">
				<p class="text-xs text-muted-foreground">Amount to send</p>
				<p class="text-2xl font-bold">{formatAmount(payAmount)} <span class="text-sm font-medium uppercase">{payCurrency}</span></p>
			</div>
		</div>

		<!-- Details -->
		<div class="space-y-4">
			<div class="flex items-center justify-between gap-4">
				<span class="text-sm text-muted-foreground">Network</span>
				<span class="text-sm font-medium">{networkLabel}</span>
			</div>

			<div>
				<p class="text-xs text-muted-foreground mb-1.5">Address</p>
				<div class="flex items-start gap-2">
					<code class="flex-1 text-xs bg-muted rounded-lg p-2.5 break-all">{payAddress}</code>
					<Button size="sm" variant="outline" onclick={copyAddress}>{copied ? 'Copied' : 'Copy'}</Button>
				</div>
			</div>

			{#if payinExtraId}
				<div>
					<p class="text-xs text-muted-foreground mb-1.5">Destination tag / memo</p>
					<div class="flex items-start gap-2">
						<code class="flex-1 text-xs bg-muted rounded-lg p-2.5">{payinExtraId}</code>
						<Button size="sm" variant="outline" onclick={copyTag}>{copiedTag ? 'Copied' : 'Copy'}</Button>
					</div>
					<p class="text-[11px] text-muted-foreground mt-1">Included automatically when scanning the QR code.</p>
				</div>
			{/if}

			<div class="pt-2 border-t border-border">
				<div class="flex items-center justify-between text-xs text-muted-foreground mb-1.5">
					<span>Time remaining</span>
					<span class="font-mono tabular-nums">{formatCountdown(remaining)}</span>
				</div>
				<div class="w-full h-1.5 bg-muted rounded-full overflow-hidden">
					<div class="h-full bg-primary transition-all" style="width: {progressPct}%"></div>
				</div>
			</div>

			{#if timedOut}
				<div class="space-y-3">
					<p class="text-sm text-destructive">Transaction not detected yet.</p>
					<Button onclick={onRetry}>Retry</Button>
				</div>
			{/if}
		</div>
	</div>
</div>
