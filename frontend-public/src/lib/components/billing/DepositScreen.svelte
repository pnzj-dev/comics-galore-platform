<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { depositId, payAddress, payAmount, payCurrency, planId, crypto, onSuccess, onTimeout, onRetry }: {
		depositId: string;
		payAddress: string;
		payAmount: number;
		payCurrency: string;
		planId: string;
		crypto: string;
		onSuccess: () => void;
		onTimeout: () => void;
		onRetry: () => void;
	} = $props();

	let copied = $state(false);
	let elapsed = $state(0);
	let timedOut = $state(false);
	let interval: ReturnType<typeof setInterval>;
	const MAX_SECONDS = 1800; // 30 minutes
	const POLL_MS = 5000;

	onMount(() => {
		interval = setInterval(async () => {
			elapsed += POLL_MS / 1000;
			if (elapsed >= MAX_SECONDS) {
				clearInterval(interval);
				timedOut = true;
				return;
			}

			try {
				const res = await encore.billing.PollDeposit(depositId);
				if (res.completed) {
					clearInterval(interval);
					onSuccess();
				}
			} catch {
				// keep polling
			}
		}, POLL_MS);

		return () => clearInterval(interval);
	});

	function copyAddress() {
		navigator.clipboard.writeText(payAddress);
		copied = true;
		setTimeout(() => copied = false, 2000);
	}

	let remaining = $derived(Math.ceil(MAX_SECONDS - elapsed));
</script>

<div class="space-y-6 text-center">
	<h3 class="text-lg font-semibold">Send Payment</h3>
	<p class="text-sm text-muted-foreground">
		Send <strong>{payAmount} {payCurrency}</strong> to the address below
	</p>

	<div class="flex justify-center">
		<div class="bg-white p-4 rounded-xl inline-block">
			<img
				src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data={encodeURIComponent(payAddress)}"
				alt="Payment QR Code"
				class="w-40 h-40 rounded-lg"
			/>
		</div>
	</div>

	<div class="max-w-xs mx-auto">
		<div class="flex items-center gap-2">
			<code class="flex-1 text-xs bg-muted rounded-lg p-3 break-all text-left">{payAddress}</code>
			<Button size="sm" variant="outline" onclick={copyAddress}>
				{copied ? 'Copied!' : 'Copy'}
			</Button>
		</div>
	</div>

	<p class="text-xs text-muted-foreground">Time remaining: {Math.ceil(remaining / 60)} min</p>

	{#if timedOut}
		<div class="space-y-3">
			<p class="text-sm text-destructive">Transaction not detected yet.</p>
			<Button onclick={onRetry}>Retry</Button>
		</div>
	{/if}
</div>
