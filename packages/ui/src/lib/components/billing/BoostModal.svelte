<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { quotaRefresh } from '$lib/stores/quota.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CoinIcon from '$lib/components/billing/CoinIcon.svelte';
	import DepositScreen from '$lib/components/billing/DepositScreen.svelte';

	let { onBoosted }: { onBoosted?: () => void } = $props();

	const open = $derived(modal.isOpen('boost'));

	type Screen = 'select' | 'deposit';

	const cryptos = [
		{ code: 'btc', label: 'BTC' },
		{ code: 'eth', label: 'ETH' },
		{ code: 'usdt', label: 'USDT' },
		{ code: 'ltc', label: 'LTC' }
	];

	let screen = $state<Screen>('select');
	let boosts = $state<{ downloads: number; price_usd: number }[]>([]);
	let loadingBoosts = $state(true);
	let selectedDownloads = $state(0);
	let selectedCrypto = $state('');
	let paying = $state(false);
	let error = $state('');
	let depositData = $state<any>(null);

	$effect(() => {
		if (open) loadBoosts();
	});

	async function loadBoosts() {
		loadingBoosts = true;
		error = '';
		screen = 'select';
		depositData = null;
		try {
			const res = await encore.billing.GetBoostOptions();
			boosts = res.boosts || [];
		} catch (e) {
			error = (e as Error).message || 'Failed to load boost options';
		}
		loadingBoosts = false;
	}

	function close() {
		modal.close('boost');
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	async function pay() {
		if (!selectedDownloads || !selectedCrypto) return;
		paying = true;
		error = '';
		try {
			const res: any = await encore.billing.CreateQuotaBoost({
				downloads: selectedDownloads,
				crypto: selectedCrypto,
			});
			depositData = res;
			screen = 'deposit';
		} catch (e) {
			error = (e as Error).message || 'Failed to start boost payment';
		}
		paying = false;
	}

	function onDepositSuccess() {
		quotaRefresh.bump();
		onBoosted?.();
		close();
	}

	function onRetry() {
		pay();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Boost Download Quota</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-5">
				{#if screen === 'select'}
					<p class="text-sm text-muted-foreground">
						You've used all your monthly downloads. Buy a one-time boost to keep downloading.
					</p>

					{#if loadingBoosts}
						<div class="flex items-center gap-2 text-sm text-muted-foreground">
							<svg class="animate-spin size-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
							Loading boost options…
						</div>
					{:else}
						<div class="grid grid-cols-3 gap-3">
							{#each boosts as boost}
								<button
									onclick={() => selectedDownloads = boost.downloads}
									class="rounded-xl border-2 p-3 text-center transition-all
										{selectedDownloads === boost.downloads ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'}"
								>
									<p class="text-sm font-bold">+{boost.downloads} downloads</p>
									<p class="text-xs text-muted-foreground">${boost.price_usd.toFixed(2)}</p>
								</button>
							{/each}
						</div>

						<div>
							<p class="text-sm font-medium mb-2">Pay with</p>
							<div class="grid grid-cols-4 gap-2">
								{#each cryptos as c}
									<button
										onclick={() => selectedCrypto = c.code}
										class="flex flex-col items-center gap-1 p-3 rounded-xl border-2 transition-all
											{selectedCrypto === c.code ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'}"
									>
										<div class="w-8 h-8 flex items-center justify-center">
											<CoinIcon code={c.code} size="32" />
										</div>
										<span class="text-xs font-medium">{c.label}</span>
									</button>
								{/each}
							</div>
						</div>
					{/if}

					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}

					<Button class="w-full" disabled={!selectedDownloads || !selectedCrypto || paying} onclick={pay}>
						{paying ? 'Preparing payment…' : 'Continue to payment'}
					</Button>
				{:else if screen === 'deposit' && depositData}
					<DepositScreen
						depositId={depositData.deposit_id}
						payAddress={depositData.pay_address}
						payAmount={depositData.pay_amount}
						payCurrency={depositData.pay_currency}
						planId={''}
						crypto={selectedCrypto}
						onSuccess={onDepositSuccess}
						onTimeout={onRetry}
						onRetry={onRetry}
					/>
				{/if}
			</div>
		</div>
	</div>
{/if}
