<script lang="ts">
	import { untrack } from 'svelte';
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { checkoutPlan, clearCheckoutPlan } from '$lib/stores/checkout.svelte';
	import PlanGrid from '$lib/components/billing/PlanGrid.svelte';
	import CryptoSelector from '$lib/components/billing/CryptoSelector.svelte';
	import CheckoutStatus from '$lib/components/billing/CheckoutStatus.svelte';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('checkout'));

	type Screen = 'plans' | 'crypto' | 'checkout';

	let screen = $state<Screen>('plans');
	let selectedPlanId = $state('');
	let selectedPriceUsdCents = $state(0);
	let selectedPlanName = $state('');
	let selectedInterval = $state('');
	let selectedCrypto = $state('');
	let checkoutId = $state('');

	function close() {
		modal.close('checkout');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	$effect(() => {
		if (!open) return;
		const planId = untrack(() => checkoutPlan.planId);
		const price = untrack(() => checkoutPlan.priceUsdCents);
		const name = untrack(() => checkoutPlan.planName);
		const interval = untrack(() => checkoutPlan.interval);
		if (planId) {
			selectedPlanId = planId;
			selectedPriceUsdCents = price;
			selectedPlanName = name;
			selectedInterval = interval;
			clearCheckoutPlan();
			screen = 'crypto';
		} else {
			screen = 'plans';
		}
	});

	function goToCrypto(selection: { planId: string; priceUsdCents: number; name: string; interval: string }) {
		selectedPlanId = selection.planId;
		selectedPriceUsdCents = selection.priceUsdCents;
		selectedPlanName = selection.name;
		selectedInterval = selection.interval;
		screen = 'crypto';
	}

	async function startCheckout(crypto: string) {
		selectedCrypto = crypto;
		try {
			const res = await encore.billing.StartSubscription({ plan_id: selectedPlanId, crypto });
			checkoutId = res.checkout_id;
			screen = 'checkout';
		} catch (err) {
			alert((err as Error).message);
		}
	}

	function onCheckoutSuccess() {
		window.location.reload();
	}

	function onCheckoutRetry() {
		startCheckout(selectedCrypto);
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()} role="presentation">

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">
					{#if screen === 'plans'}Choose a Plan{:else if screen === 'crypto'}Pay with Crypto{:else}Checkout{/if}
				</h2>
				<button onclick={close} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto flex-1">
				{#if screen === 'plans'}
					<PlanGrid onSelect={goToCrypto} />
				{:else if screen === 'crypto'}
					<CryptoSelector planId={selectedPlanId} priceUsdCents={selectedPriceUsdCents} planName={selectedPlanName} interval={selectedInterval} onBack={() => screen = 'plans'} onContinue={startCheckout} />
				{:else if screen === 'checkout'}
					<CheckoutStatus checkoutId={checkoutId} onSuccess={onCheckoutSuccess} onRetry={onCheckoutRetry} />
				{/if}
			</div>
		</div>
	</div>
{/if}
