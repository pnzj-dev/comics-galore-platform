<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';

	let { planId, onBack, onContinue }: {
		planId: string;
		onBack: () => void;
		onContinue: (crypto: string) => void;
	} = $props();

	const cryptos = [
		{ code: 'btc', label: 'BTC', color: 'bg-orange-500' },
		{ code: 'eth', label: 'ETH', color: 'bg-blue-500' },
		{ code: 'usdt', label: 'USDT', color: 'bg-green-500' },
		{ code: 'ltc', label: 'LTC', color: 'bg-gray-400' },
	];

	let selectedCrypto = $state('');
	let estimate = $state<number | null>(null);
	let estimating = $state(false);
	let error = $state('');

	async function selectCrypto(code: string) {
		if (selectedCrypto === code) {
			selectedCrypto = '';
			estimate = null;
			return;
		}
		selectedCrypto = code;
		estimating = true;
		error = '';
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

	<div class="grid grid-cols-4 gap-3">
		{#each cryptos as c}
			<button
				onclick={() => selectCrypto(c.code)}
				class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 transition-all
					{selectedCrypto === c.code
						? 'border-primary bg-primary/5'
						: selectedCrypto
							? 'border-border opacity-40 cursor-not-allowed'
							: 'border-border hover:border-primary/50'}"
				disabled={!!selectedCrypto && selectedCrypto !== c.code}
			>
				<div class="w-10 h-10 rounded-full {c.color} flex items-center justify-center text-white font-bold text-sm">{c.label}</div>
				<span class="text-xs font-medium">{c.label}</span>
			</button>
		{/each}
	</div>

	{#if estimating}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<svg class="animate-spin size-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
			Getting live price estimate...
		</div>
	{:else if estimate !== null}
		<div class="p-4 rounded-lg bg-primary/5 border border-primary/20">
			<p class="text-sm text-muted-foreground">Estimated price</p>
			<p class="text-2xl font-bold">{estimate.toFixed(6)} <span class="text-sm uppercase">{selectedCrypto}</span></p>
		</div>
	{:else if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}

	<Button class="w-full" disabled={!selectedCrypto || estimate === null} onclick={() => onContinue(selectedCrypto)}>
		Continue
	</Button>
</div>
