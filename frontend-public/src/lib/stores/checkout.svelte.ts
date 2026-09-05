export type CheckoutPlan = {
	planId: string | null;
	priceUsdCents: number;
	planName: string;
	interval: string;
};

export const checkoutPlan = $state<CheckoutPlan>({ planId: null, priceUsdCents: 0, planName: '', interval: '' });

export function setCheckoutPlan(planId: string, priceUsdCents: number, planName: string, interval: string) {
	checkoutPlan.planId = planId;
	checkoutPlan.priceUsdCents = priceUsdCents;
	checkoutPlan.planName = planName;
	checkoutPlan.interval = interval;
}

export function clearCheckoutPlan() {
	checkoutPlan.planId = null;
	checkoutPlan.priceUsdCents = 0;
	checkoutPlan.planName = '';
	checkoutPlan.interval = '';
}
