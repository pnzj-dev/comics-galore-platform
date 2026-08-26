export type CheckoutPlan = {
	planId: string | null;
	priceUsdCents: number;
};

export const checkoutPlan = $state<CheckoutPlan>({ planId: null, priceUsdCents: 0 });

export function setCheckoutPlan(planId: string, priceUsdCents: number) {
	checkoutPlan.planId = planId;
	checkoutPlan.priceUsdCents = priceUsdCents;
}

export function clearCheckoutPlan() {
	checkoutPlan.planId = null;
	checkoutPlan.priceUsdCents = 0;
}
