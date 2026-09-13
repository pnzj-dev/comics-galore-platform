<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { checkoutId, onSuccess, onRetry }: {
		checkoutId: string;
		onSuccess: () => void;
		onRetry: () => void;
	} = $props();

	const POLL_MS = 3000;

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

	let checkout = $state({
		step: 'checking',
		pay_address: '',
		pay_amount: '',
		pay_currency: '',
		payin_extra_id: '',
		network: '',
		qr_data_url: '',
		expires_at: ''
	});
	let failed = $state(false);
	let copied = $state(false);
	let copiedTag = $state(false);
	let now = $state(Date.now());

	let pollInterval: ReturnType<typeof setInterval>;
	let tickInterval: ReturnType<typeof setInterval>;

	onMount(() => {
		tickInterval = setInterval(() => (now = Date.now()), 1000);
		pollInterval = setInterval(poll, POLL_MS);
		poll();
		return () => {
			clearInterval(pollInterval);
			clearInterval(tickInterval);
		};
	});

	async function poll() {
		try {
			const s = await encore.billing.GetCheckoutState(checkoutId);
			checkout = {
				step: s.step,
				pay_address: s.pay_address,
				pay_amount: s.pay_amount,
				pay_currency: s.pay_currency,
				payin_extra_id: s.payin_extra_id,
				network: s.network,
				qr_data_url: s.qr_data_url,
				expires_at: s.expires_at
			};
			if (s.step === 'active') {
				clearInterval(pollInterval);
				onSuccess();
			} else if (s.step === 'expired' || s.step === 'failed') {
				clearInterval(pollInterval);
				failed = true;
			}
		} catch {
			// keep polling
		}
	}

	function copyAddress() {
		navigator.clipboard.writeText(checkout.pay_address);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function copyTag() {
		if (!checkout.payin_extra_id) return;
		navigator.clipboard.writeText(checkout.payin_extra_id);
		copiedTag = true;
		setTimeout(() => (copiedTag = false), 2000);
	}

	function formatAmount(s: string): string {
		const n = parseFloat(s);
		if (!n) return '0';
		return n.toFixed(8).replace(/\.?0+$/, '');
	}

	function formatCountdown(s: number): string {
		const sec = Math.max(0, Math.floor(s));
		const m = Math.floor(sec / 60);
		const rem = sec % 60;
		return `${String(m).padStart(2, '0')}:${String(rem).padStart(2, '0')}`;
	}

	const networkLabel = $derived(checkout.network ? (NETWORK_NAMES[checkout.network.toLowerCase()] || checkout.network.toUpperCase()) : checkout.pay_currency.toUpperCase());
	const expiresAtMs = $derived(checkout.expires_at ? new Date(checkout.expires_at).getTime() : 0);
	const remainingSecs = $derived(expiresAtMs ? Math.max(0, (expiresAtMs - now) / 1000) : 0);
	const progressPct = $derived(expiresAtMs ? Math.min(100, ((Date.now() - (expiresAtMs - 30 * 60 * 1000)) / (30 * 60 * 1000)) * 100) : 0);

	const awaitingDeposit = $derived(checkout.step === 'awaiting_deposit');
	const processing = $derived(checkout.step === 'checking' || checkout.step === 'awaiting_subscription' || checkout.step === 'completed');
</script>

{#if awaitingDeposit}
	<div class="space-y-6">
		<h3 class="text-lg font-semibold">Send Payment</h3>

		<div class="grid gap-6 md:grid-cols-[auto_1fr] md:items-start">
			<div class="flex flex-col items-center gap-4">
				{#if checkout.qr_data_url}
					<div class="bg-white p-3 rounded-xl inline-block">
						<img src={checkout.qr_data_url} alt="Payment QR Code" class="w-44 h-44 rounded-lg" />
					</div>
				{/if}
				<div class="text-center">
					<p class="text-xs text-muted-foreground">Amount to send</p>
					<p class="text-2xl font-bold">{formatAmount(checkout.pay_amount)} <span class="text-sm font-medium uppercase">{checkout.pay_currency}</span></p>
				</div>
			</div>

			<div class="space-y-4">
				<div class="flex items-center justify-between gap-4">
					<span class="text-sm text-muted-foreground">Network</span>
					<span class="text-sm font-medium">{networkLabel}</span>
				</div>

				<div>
					<p class="text-xs text-muted-foreground mb-1.5">Address</p>
					<div class="flex items-start gap-2">
						<code class="flex-1 text-xs bg-muted rounded-lg p-2.5 break-all">{checkout.pay_address}</code>
						<Button size="sm" variant="outline" onclick={copyAddress}>{copied ? 'Copied' : 'Copy'}</Button>
					</div>
				</div>

				{#if checkout.payin_extra_id}
					<div>
						<p class="text-xs text-muted-foreground mb-1.5">Destination tag / memo</p>
						<div class="flex items-start gap-2">
							<code class="flex-1 text-xs bg-muted rounded-lg p-2.5">{checkout.payin_extra_id}</code>
							<Button size="sm" variant="outline" onclick={copyTag}>{copiedTag ? 'Copied' : 'Copy'}</Button>
						</div>
						<p class="text-[11px] text-muted-foreground mt-1">Included automatically when scanning the QR code.</p>
					</div>
				{/if}

				<div class="pt-2 border-t border-border">
					<div class="flex items-center justify-between text-xs text-muted-foreground mb-1.5">
						<span>Time remaining</span>
						<span class="font-mono tabular-nums">{formatCountdown(remainingSecs)}</span>
					</div>
					<div class="w-full h-1.5 bg-muted rounded-full overflow-hidden">
						<div class="h-full bg-primary transition-all" style="width: {Math.min(100, progressPct)}%"></div>
					</div>
				</div>
			</div>
		</div>
	</div>
{:else if processing}
	<div class="space-y-6 text-center">
		<div class="flex justify-center">
			<svg class="animate-spin size-12 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
		</div>
		<h3 class="text-lg font-semibold">Processing Subscription</h3>
		<p class="text-sm text-muted-foreground">Please wait while your subscription is being processed...</p>
	</div>
{:else}
	<div class="space-y-6 text-center">
		<p class="text-sm text-destructive">{failed ? 'Checkout failed or timed out.' : 'Something went wrong.'}</p>
		<Button onclick={onRetry}>Retry</Button>
	</div>
{/if}
