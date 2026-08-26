<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { checkoutPlan, clearCheckoutPlan } from '$lib/stores/checkout.svelte';
	import PlanGrid from '$lib/components/billing/PlanGrid.svelte';
	import CryptoSelector from '$lib/components/billing/CryptoSelector.svelte';
	import ProcessingScreen from '$lib/components/billing/ProcessingScreen.svelte';
	import DepositScreen from '$lib/components/billing/DepositScreen.svelte';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('checkout'));

	type Screen = 'plans' | 'crypto' | 'processing' | 'deposit';

	let screen = $state<Screen>('plans');
	let selectedPlanId = $state('');
	let selectedPriceUsdCents = $state(0);
	let selectedCrypto = $state('');
	let subscriptionId = $state('');
	let depositData = $state<any>(null);

	function close() {
		modal.close('checkout');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	$effect(() => {
		if (!open) return;
		if (checkoutPlan.planId) {
			selectedPlanId = checkoutPlan.planId;
			selectedPriceUsdCents = checkoutPlan.priceUsdCents;
			clearCheckoutPlan();
			screen = 'crypto';
		} else {
			screen = 'plans';
		}
	});

	function goToCrypto(planId: string, priceUsdCents: number) {
		selectedPlanId = planId;
		selectedPriceUsdCents = priceUsdCents;
		screen = 'crypto';
	}

	async function goToCheckout(crypto: string) {
		selectedCrypto = crypto;
		try {
			const res = await encore.billing.CheckBalance();
			const balances = res.balances || {};
			const balance = balances[crypto] || balances[crypto.toUpperCase()];
			const hasBalance = (balance?.amount || 0) > 0;

			if (hasBalance) {
				await fundSubscription();
			} else {
				await createDeposit();
			}
		} catch {
			// If balance check fails, try subscription anyway
			await fundSubscription();
		}
	}

	async function fundSubscription() {
		try {
			const subRes = await encore.billing.CreateSubscription({ plan_id: selectedPlanId });
			subscriptionId = subRes.subscription_id;
			screen = 'processing';
		} catch {
			await createDeposit();
		}
	}

	async function createDeposit() {
		try {
			const res: any = await encore.billing.CreateDeposit({
				plan_id: selectedPlanId,
				crypto: selectedCrypto
			});
			depositData = res;
			screen = 'deposit';
		} catch (err) {
			alert((err as Error).message);
			close();
		}
	}

	function onProcessingSuccess() {
		window.location.reload();
	}

	function onDepositSuccess() {
		fundSubscription();
	}

	function onRetryProcessing() {
		fundSubscription();
	}

	function onRetryDeposit() {
		createDeposit();
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()} role="presentation">

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">
					{#if screen === 'plans'}Choose a Plan{:else if screen === 'crypto'}Pay with Crypto{:else if screen === 'processing'}Processing{:else}Send Payment{/if}
				</h2>
				<button onclick={close} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto flex-1">
				{#if screen === 'plans'}
					<PlanGrid onSelect={goToCrypto} />
				{:else if screen === 'crypto'}
					<CryptoSelector planId={selectedPlanId} priceUsdCents={selectedPriceUsdCents} onBack={() => screen = 'plans'} onContinue={goToCheckout} />
				{:else if screen === 'processing'}
					<ProcessingScreen subscriptionId={subscriptionId} onSuccess={onProcessingSuccess} onRetry={onRetryProcessing} />
				{:else if screen === 'deposit'}
					<DepositScreen
						depositId={depositData.deposit_id}
						payAddress={depositData.pay_address}
						payAmount={depositData.pay_amount}
						payCurrency={depositData.pay_currency}
						planId={selectedPlanId}
						crypto={selectedCrypto}
						onSuccess={onDepositSuccess}
						onTimeout={onRetryDeposit}
						onRetry={onRetryDeposit}
					/>
				{/if}
			</div>
		</div>
	</div>
{/if}
