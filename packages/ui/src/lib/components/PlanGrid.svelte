<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';

	let { onSelect }: { onSelect: (planId: string) => void } = $props();

	let tiers = $state<any[]>([]);
	let plans = $state<any[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const [tRes, pRes] = await Promise.all([
				encore.tiers.ListTiers(),
				encore.tiers.ListPlans()
			]);
			tiers = tRes.tiers;
			plans = pRes.plans;
		} catch { /* */ }
		loading = false;
	});

	function intervalLabel(i: string): string {
		return i.charAt(0).toUpperCase() + i.slice(1);
	}

	function tierPlans(tierId: string) {
		return plans.filter((p: any) => p.tier_id === tierId);
	}

	function parseFeatures(f: any): string[] {
		if (Array.isArray(f)) return f;
		if (typeof f === 'string') {
			try { return JSON.parse(f); } catch { return []; }
		}
		return [];
	}
</script>

{#if loading}
	<div class="py-12 text-center text-muted-foreground">Loading plans...</div>
{:else}
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5 max-h-[60vh] overflow-y-auto p-1">
		{#each tiers as tier}
			<Card class="h-full">
				<CardHeader>
					<CardTitle>{tier.name}</CardTitle>
					<CardDescription>{tier.description}</CardDescription>
				</CardHeader>
				<CardContent class="space-y-3">
					{#each tierPlans(tier.id) as plan}
						<button class="w-full text-left p-2 rounded-lg border border-border hover:border-primary/50 transition-colors cursor-pointer bg-transparent" onclick={() => onSelect(plan.id)}>
							<div class="flex justify-between items-center">
								<span class="text-sm font-medium">{intervalLabel(plan.interval)}</span>
								<span class="text-sm font-semibold">
									{#if plan.price_usd_cents === 0}
										Free
									{:else}
										${(plan.price_usd_cents / 100).toFixed(2)}
									{/if}
								</span>
						</button>
							{#if plan.features && parseFeatures(plan.features).length > 0}
								<ul class="mt-1 space-y-0.5">
									{#each parseFeatures(plan.features) as feat}
										<li class="text-[10px] text-muted-foreground flex items-center gap-1">
											<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" class="text-green-500 flex-shrink-0"><path d="M20 6 9 17l-5-5"/></svg>
											{feat}
										</li>
									{/each}
								</ul>
							{/if}
						</div>
					{/each}
					<Button class="w-full text-xs" size="sm" onclick={() => {
						const firstPlan = tierPlans(tier.id)[0];
						if (firstPlan) onSelect(firstPlan.id);
					}}>
						Select {tier.name}
					</Button>
				</CardContent>
			</Card>
		{/each}
	</div>
{/if}
