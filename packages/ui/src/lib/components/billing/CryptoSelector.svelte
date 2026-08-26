<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CoinIcon from '$lib/components/billing/CoinIcon.svelte';
	import { isStablecoin, coinLabel } from '$lib/utils/crypto';

	let { planId, priceUsdCents, onBack, onContinue }: {
		planId: string;
		priceUsdCents: number;
		onBack: () => void;
		onContinue: (crypto: string) => void;
	} = $props();

	let cryptos = $state<string[]>([]);
	let loading = $state(true);
	let selectedCrypto = $state('');
	let estimate = $state<number | null>(null);
	let estimating = $state(false);
	let stableSelected = $state(false);
	let error = $state('');

	let cachedCurrencies: string[] | null = null;

	onMount(async () => {
		if (cachedCurrencies) {
			cryptos = cachedCurrencies;
			loading = false;
			return;
		}
		try {
			const res = await encore.billing.ListCurrencies();
			cachedCurrencies = res.currencies || [];
			cryptos = cachedCurrencies;
		} catch {
			cryptos = [];
		}
		loading = false;
	});

	async function selectCrypto(code: string) {
		if (selectedCrypto === code) {
			selectedCrypto = '';
			estimate = null;
			stableSelected = false;
			return;
		}

		selectedCrypto = code;
		estimate = null;
		stableSelected = false;
		error = '';

		if (isStablecoin(code)) {
			stableSelected = true;
			estimate = priceUsdCents / 100;
			return;
		}

		estimating = true;
		try {
			const res = await encore.billing.EstimatePrice({ plan_id: planId, crypto: code });
			estimate = res.estimated_amount;
		} catch (err) {
			error = (err as Error).message;
			estimate = null;
		} finally {
			estimating = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex items-center gap-2">
		<button onclick={onBack} class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to plans</button>
	</div>

	<h3 class="text-lg font-semibold">Choose Payment Currency</h3>

	{#if loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<svg class="animate-spin size-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
			Loading currencies…
		</div>
	{:else if cryptos.length === 0}
		<p class="text-sm text-muted-foreground">No payment currencies are currently available.</p>
	{:else}
		<div class="grid grid-cols-4 gap-3">
			{#each cryptos as code (code)}
				<button
					onclick={() => selectCrypto(code)}
					class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 transition-all
						{selectedCrypto === code
							? 'border-primary bg-primary/5'
							: selectedCrypto
								? 'border-border opacity-40 cursor-not-allowed'
								: 'border-border hover:border-primary/50'}"
					disabled={!!selectedCrypto && selectedCrypto !== code}
				>
					<div class="w-10 h-10 flex items-center justify-center">
						<CoinIcon code={code} size="40" />
					</div>
					<span class="text-xs font-medium">{coinLabel(code)}</span>
				</button>
			{/each}
		</div>
	{/if}

	{#if estimating}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<svg class="animate-spin size-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
			Getting live price estimate...
		</div>
	{:else if estimate !== null}
		<div class="p-4 rounded-lg bg-primary/5 border border-primary/20">
			{#if stableSelected}
				<p class="text-sm text-muted-foreground">Price</p>
				<p class="text-2xl font-bold">${estimate.toFixed(2)}</p>
			{:else}
				<p class="text-sm text-muted-foreground">Estimated price</p>
				<p class="text-2xl font-bold">{estimate.toFixed(6)} <span class="text-sm uppercase">{coinLabel(selectedCrypto)}</span></p>
			{/if}
		</div>
	{:else if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	<Button class="w-full" disabled={!selectedCrypto || estimate === null || estimating} onclick={() => onContinue(selectedCrypto)}>
		Continue
	</Button>
</div>
