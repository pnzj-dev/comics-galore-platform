<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { currentUser } from '$lib/stores/auth';

	interface Props {
		mode?: 'modal' | 'page';
		onSelect?: (planId: string, interval: string) => void;
	}

	let { mode = 'modal', onSelect }: Props = $props();

	let tiers = $state<any[]>([]);
	let plans = $state<any[]>([]);
	let loading = $state(true);
	let selectedIntervals = $state<Record<string, string>>({});

	const user = $derived($currentUser);
	const isModal = $derived(mode === 'modal');

	onMount(async () => {
		try {
			const [tRes, pRes] = await Promise.all([
				api.get<{ tiers: any[] }>('/tiers'),
				api.get<{ plans: any[] }>('/plans')
			]);
			tiers = tRes.tiers;
			plans = pRes.plans;
			for (const t of tiers) {
				selectedIntervals[t.id] = 'monthly';
			}
		} catch {}
		loading = false;
	});

	function parseFeatures(f: any): string[] {
		if (Array.isArray(f)) return f;
		if (typeof f === 'string') { try { return JSON.parse(f); } catch { return []; } }
		return [];
	}

	function tierFeatures(tierName: string, interval: string): string[] {
		const plan = plans.find((p: any) => p.tier_id === getTierId(tierName) && p.interval === interval);
		return plan ? parseFeatures(plan.features) : [];
	}

	function tierPrice(tierName: string, interval: string): number {
		const plan = plans.find((p: any) => p.tier_id === getTierId(tierName) && p.interval === interval);
		return plan ? plan.price_usd_cents : 0;
	}

	function tierPlanId(tierName: string, interval: string): string {
		const plan = plans.find((p: any) => p.tier_id === getTierId(tierName) && p.interval === interval);
		return plan ? plan.id : '';
	}

	function getTierId(name: string): string {
		const t = tiers.find((t: any) => t.name === name);
		return t?.id || '';
	}

	function previousTierName(name: string): string | null {
		const idx = tiers.findIndex((t: any) => t.name === name);
		return idx > 0 ? tiers[idx - 1].name : null;
	}

	function diffFeatures(name: string, interval: string): { feature: string; isNew: boolean }[] {
		const current = tierFeatures(name, interval);
		const prevName = previousTierName(name);
		const previous = prevName ? tierFeatures(prevName, interval) : [];
		return current.map(f => ({ feature: f, isNew: !previous.includes(f) }));
	}

	function tierIntervals(tierName: string): string[] {
		return plans
			.filter((p: any) => p.tier_id === getTierId(tierName))
			.map((p: any) => p.interval);
	}
</script>

{#if loading}
	<div class="py-12 text-center text-muted-foreground">Loading plans...</div>
{:else}
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
		{#each tiers as tier}
			{@const isFree = tier.name === 'Free'}
			{@const intervals = tierIntervals(tier.name)}
			{@const selInterval = selectedIntervals[tier.id] || 'monthly'}
			{@const price = tierPrice(tier.name, selInterval)}
			{@const features = diffFeatures(tier.name, selInterval)}
			{@const planId = tierPlanId(tier.name, selInterval)}

			<Card class="flex flex-col {isFree ? 'opacity-80' : ''}">
				<CardHeader>
					<CardTitle>{tier.name}</CardTitle>
				</CardHeader>
				<CardContent class="flex-1 flex flex-col justify-between gap-4">
					{#if !isFree && intervals.length > 1}
						<select
							bind:value={selectedIntervals[tier.id]}
							class="w-full rounded-md border border-input bg-background px-3 py-2 text-xs"
						>
							{#each intervals as intv}
								<option value={intv}>{intv}</option>
							{/each}
						</select>
					{/if}

					<ul class="space-y-1 flex-1">
						{#each features as f}
							<li class="text-[11px] flex items-start gap-1.5">
								{#if f.isNew}
									<span class="text-primary font-bold text-xs leading-tight mt-px">+</span>
								{:else}
									<span class="text-green-500 text-xs leading-tight mt-px">✓</span>
								{/if}
								<span class={f.isNew ? 'text-foreground font-medium' : 'text-muted-foreground'}>{f.feature}</span>
							</li>
						{/each}
					</ul>

					<div class="pt-2 border-t border-border">
						<div class="flex items-center justify-between">
							<div>
								<span class="text-lg font-bold">
									{#if price === 0}
										Free
									{:else}
										${(price / 100).toFixed(2)}
									{/if}
								</span>
								{#if price > 0}
									<span class="text-xs text-muted-foreground">/{selInterval}</span>
								{/if}
							</div>

							{#if isModal && !isFree}
								<Button size="sm" onclick={() => onSelect?.(planId, selInterval)}>Select</Button>
							{/if}
						</div>
					</div>
				</CardContent>
			</Card>
		{/each}
	</div>
{/if}
