<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';

	interface Props {
		mode?: 'modal' | 'page';
		onSelect?: (planId: string, priceUsdCents: number) => void;
	}

	let { mode = 'modal', onSelect }: Props = $props();

	let tiers = $state<any[]>([]);
	let plans = $state<any[]>([]);
	let loading = $state(true);

	let selectedIntervals = $state<Record<string, string>>({});

	const isModal = $derived(mode === 'modal');

	const intervalLabels: Record<string, string> = {
		monthly: 'monthly',
		quarterly: 'quarterly',
		semesterly: 'semesterly',
		yearly: 'yearly'
	};

	onMount(async () => {
		try {
			const [tRes, pRes] = await Promise.all([
				encore.tiers.ListTiers(),
				encore.tiers.ListPlans()
			]);
			tiers = tRes.tiers;
			plans = pRes.plans;
			for (const t of tiers) {
				const tp = pRes.plans.filter((p: any) => p.tier_id === t.id);
				if (tp.length > 0) selectedIntervals[t.id] = tp[0].interval;
			}
		} catch {}
		loading = false;
	});

	function parseFeatures(f: any): string[] {
		if (Array.isArray(f)) return f;
		if (typeof f === 'string') { try { return JSON.parse(f); } catch { return []; } }
		return [];
	}

	function tierPlans(tierId: string): any[] {
		return plans.filter((p: any) => p.tier_id === tierId);
	}

	function priceForInterval(tierId: string, interval: string): number {
		const p = plans.find((pl: any) => pl.tier_id === tierId && pl.interval === interval);
		return p ? p.price_usd_cents : 0;
	}

	function planIdFor(tierId: string, interval: string): string {
		const p = plans.find((pl: any) => pl.tier_id === tierId && pl.interval === interval);
		return p ? p.id : '';
	}

	function tierFeatures(tierId: string): string[] {
		const tp = tierPlans(tierId);
		if (tp.length === 0) return [];
		return parseFeatures(tp[0].features);
	}

	function previousFeatures(tierIdx: number): string[] {
		if (tierIdx <= 0) return [];
		const prevId = tiers[tierIdx - 1]?.id;
		return tierFeatures(prevId);
	}

	function formatQuota(downloads: number): string {
		if (downloads >= 999999) return 'Unlimited downloads';
		return `${downloads} downloads/month`;
	}
</script>

{#if loading}
	<div class="py-12 text-center text-muted-foreground">Loading plans...</div>
{:else}
	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
		{#each tiers as tier, idx}
			{@const isFree = tier.name === 'Free'}
			{@const tp = tierPlans(tier.id)}
			{@const allFeatures = tierFeatures(tier.id)}
			{@const prevFeatures = previousFeatures(idx)}

			<Card class="flex flex-col {isFree ? 'opacity-80' : ''}">
				<CardHeader>
					<CardTitle>{tier.name}</CardTitle>
				</CardHeader>
				<CardContent class="flex-1 flex flex-col justify-between gap-3">
					{#if !isFree && tp.length > 1}
						<Select type="single" bind:value={selectedIntervals[tier.id]}>
							<SelectTrigger class="w-full">
								{intervalLabels[selectedIntervals[tier.id]] || selectedIntervals[tier.id]}
							</SelectTrigger>
							<SelectContent>
								{#each tp as plan}
									<SelectItem value={plan.interval}>{intervalLabels[plan.interval] || plan.interval}</SelectItem>
								{/each}
							</SelectContent>
						</Select>
					{/if}

					<ul class="space-y-1 flex-1">
						{#each allFeatures as feature}
							{@const isNew = !prevFeatures.includes(feature)}
							<li class="text-[11px] flex items-start gap-1.5">
								{#if isFree || isNew}
									<span class="text-primary font-bold text-xs leading-tight mt-px">+</span>
								{:else}
									<span class="text-green-500 text-xs leading-tight mt-px">✓</span>
								{/if}
								<span class={isFree || isNew ? 'text-foreground font-medium' : 'text-muted-foreground'}>{feature}</span>
							</li>
						{/each}
						<li class="text-[11px] flex items-start gap-1.5 pt-1 border-t border-border">
							<span class="text-primary font-bold text-xs leading-tight mt-px">+</span>
							<span class="text-foreground font-semibold">{formatQuota(tier.quota_downloads)}</span>
						</li>
					</ul>

					<div class="pt-2 border-t border-border">
						<div class="flex items-center justify-between">
							<div>
								<span class="text-lg font-bold">
									{#if priceForInterval(tier.id, selectedIntervals[tier.id] || 'monthly') === 0}Free
									{:else}${(priceForInterval(tier.id, selectedIntervals[tier.id] || 'monthly') / 100).toFixed(2)}
									{/if}
								</span>
								{#if priceForInterval(tier.id, selectedIntervals[tier.id] || 'monthly') > 0}
									<span class="text-xs text-muted-foreground">/{selectedIntervals[tier.id] || 'monthly'}</span>
								{/if}
							</div>

							{#if onSelect && !isFree}
								<Button size="sm" onclick={() => onSelect(planIdFor(tier.id, selectedIntervals[tier.id] || 'monthly'), priceForInterval(tier.id, selectedIntervals[tier.id] || 'monthly'))}>{isModal ? 'Select' : 'Subscribe'}</Button>
							{/if}
						</div>
					</div>
				</CardContent>
			</Card>
		{/each}
	</div>
{/if}
