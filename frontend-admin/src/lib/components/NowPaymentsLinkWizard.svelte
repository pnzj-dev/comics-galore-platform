<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { AlertCircle, CheckCircle, LoaderCircle, Settings } from 'lucide-svelte';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();

	type Plan = { id: string; name: string; tier_id: string; interval: string; price_usd_cents: number; provider_plan_id: string };

	let unlinked = $state<Plan[]>([]);
	let allLinked = $state(false);
	let inputs = $state<Record<string, string>>({});
	let logLines = $state<Array<{ text: string; ok: boolean; loading: boolean }>>([]);
	let loading = $state(false);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}

	async function loadPlans() {
		try {
			const res = await api.get<{ plans: Plan[] }>('/plans');
			const plans = res.plans || [];
			const missing = plans.filter(p => !p.provider_plan_id);
			unlinked = missing;
			allLinked = missing.length === 0;
			inputs = {};
			logLines = [];
			missing.forEach(p => { inputs[p.id] = ''; });
		} catch {
			unlinked = [];
			allLinked = true;
		}
	}

	$effect(() => {
		if (open) loadPlans();
	});

	async function linkPlan(planId: string, providerPlanId: string) {
		const plan = unlinked.find(p => p.id === planId);
		if (!plan || !providerPlanId.trim()) return;

		logLines = [...logLines, { text: `Linking ${plan.name} → ID ${providerPlanId}...`, ok: false, loading: true }];
		try {
			await api.patch<{ ok: boolean }>(`/admin/plans/${planId}`, { provider_plan_id: providerPlanId });
			logLines = logLines.map(l =>
				l.text.startsWith(`Linking ${plan.name}`) ? { text: `✓ ${plan.name} linked to ID ${providerPlanId}`, ok: true, loading: false } : l
			);
			await loadPlans();
		} catch (err: any) {
			logLines = logLines.map(l =>
				l.text.startsWith(`Linking ${plan.name}`) ? { text: `✗ ${plan.name} failed: ${err?.message || 'unknown error'}`, ok: false, loading: false } : l
			);
		}
	}

	async function linkAll() {
		loading = true;
		logLines = [];
		for (const plan of unlinked) {
			const val = inputs[plan.id];
			if (val?.trim()) {
				await linkPlan(plan.id, val.trim());
			}
		}
		loading = false;
	}

	function formatPrice(cents: number): string {
		return `$${(cents / 100).toFixed(2)}`;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-60 bg-black/50 flex items-center justify-center p-4" onclick={onClose} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-2xl max-h-[85vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b flex-shrink-0">
				<div class="flex items-center gap-2">
					<Settings class="size-5 text-primary" />
					<h2 class="text-lg font-semibold">Link Plans to NowPayments</h2>
				</div>
				<button onclick={onClose} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto flex-1 space-y-4">
				{#if allLinked}
					<div class="flex flex-col items-center py-8 gap-3">
						<CheckCircle class="size-12 text-green-500" />
						<p class="text-lg font-medium">All plans are linked</p>
						<p class="text-sm text-muted-foreground">Your plan matrix is fully configured for NowPayments.</p>
						<Button onclick={onClose}>Done</Button>
					</div>
				{:else}
					<div class="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800">
						<AlertCircle class="size-4 text-amber-600 mt-0.5 flex-shrink-0" />
						<p class="text-sm text-amber-800 dark:text-amber-200">
							{unlinked.length} {unlinked.length === 1 ? 'plan needs' : 'plans need'} a NowPayments plan ID. Get these from your <a href="https://nowpayments.io" target="_blank" class="underline font-medium">NowPayments dashboard</a>.
						</p>
					</div>

					<div class="rounded-xl border border-border overflow-hidden">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b bg-muted/50 text-left">
									<th class="px-4 py-2.5 font-medium">Plan</th>
									<th class="px-4 py-2.5 font-medium">Interval</th>
									<th class="px-4 py-2.5 font-medium">Price</th>
									<th class="px-4 py-2.5 font-medium">NowPayments ID</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-border">
								{#each unlinked as plan}
									<tr>
										<td class="px-4 py-2.5 font-medium">{plan.name}</td>
										<td class="px-4 py-2.5 text-muted-foreground capitalize">{plan.interval}</td>
										<td class="px-4 py-2.5 text-muted-foreground">{formatPrice(plan.price_usd_cents)}</td>
										<td class="px-4 py-2.5">
											<div class="flex items-center gap-2">
												<input
													type="text"
													placeholder="e.g. 12345"
													bind:value={inputs[plan.id]}
													class="w-28 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm"
												/>
												<Button size="sm" variant="outline" disabled={!inputs[plan.id]?.trim() || loading} onclick={() => linkPlan(plan.id, inputs[plan.id])}>
													Link
												</Button>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					{#if unlinked.length > 1}
						<Button onclick={linkAll} disabled={loading || unlinked.every(p => !inputs[p.id]?.trim())} class="w-full">
							{loading ? 'Linking...' : `Link All (${unlinked.filter(p => inputs[p.id]?.trim()).length} ready)`}
						</Button>
					{/if}

					{#if logLines.length > 0}
						<div class="rounded-xl border border-border p-3 space-y-1 max-h-40 overflow-y-auto bg-muted/30">
							{#each logLines as line}
								<div class="flex items-center gap-1.5 text-xs">
									{#if line.loading}
										<LoaderCircle class="size-3 animate-spin text-muted-foreground" />
									{:else if line.ok}
										<CheckCircle class="size-3 text-green-500" />
									{:else}
										<AlertCircle class="size-3 text-red-500" />
									{/if}
									<span class={line.loading ? 'text-muted-foreground' : line.ok ? 'text-green-700 dark:text-green-400' : 'text-red-600 dark:text-red-400'}>
										{line.text}
									</span>
								</div>
							{/each}
						</div>
					{/if}
				{/if}
			</div>
		</div>
	</div>
{/if}
