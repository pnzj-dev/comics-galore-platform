<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { AlertCircle, CheckCircle, LoaderCircle, Settings } from 'lucide-svelte';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();

	type Plan = { id: string; name: string; tier_id: string; interval: string; price_usd_cents: number; provider_plan_id: string };

	let unlinked = $state<Plan[]>([]);
	let allLinked = $state(false);
	let inputs = $state<Record<string, string>>({});
	let logLines = $state<Array<{ text: string; ok: boolean; loading: boolean }>>([]);
	let autoLinking = $state(false);
	let autoProgress = $state({ done: 0, total: 0 });
	let mode = $state<'manual' | 'automatic'>('manual');
	let planLinkState = $state<Record<string, 'pending' | 'linking' | 'success' | 'failed'>>({});
	let failedIds = $state<string[]>([]);

	let sortKey = $state<string | null>(null);
	let sortDir = $state<'asc' | 'desc'>('asc');

	let validationError = $state('');

	let sortedUnlinked = $derived(sortPlans(unlinked, sortKey, sortDir));

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}

	function planDisplayName(p: Plan): string {
		if (p.name === 'Free') return p.name;
		const cap = p.interval.charAt(0).toUpperCase() + p.interval.slice(1);
		return `${p.name} - ${cap}`;
	}

	function formatPrice(cents: number): string {
		return `$${(cents / 100).toFixed(2)}`;
	}

	function toggleSort(key: string) {
		if (sortKey === key) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortKey = key;
			sortDir = 'asc';
		}
	}

	function sortPlans(plans: Plan[], key: string | null, dir: 'asc' | 'desc'): Plan[] {
		if (!key) return plans;
		const sorted = [...plans].sort((a, b) => {
			let cmp = 0;
			if (key === 'name') cmp = planDisplayName(a).localeCompare(planDisplayName(b));
			else if (key === 'interval') cmp = a.interval.localeCompare(b.interval);
			else if (key === 'price') cmp = a.price_usd_cents - b.price_usd_cents;
			return dir === 'asc' ? cmp : -cmp;
		});
		return sorted;
	}

	function sortArrow(key: string): string {
		if (sortKey !== key) return '';
		return sortDir === 'asc' ? ' ↑' : ' ↓';
	}

	async function loadPlans() {
		try {
			const res = await encore.tiers.ListPlans();
			const plans = res.plans || [];
			const missing = plans.filter(p => !p.provider_plan_id && p.price_usd_cents > 0);
			unlinked = missing;
			allLinked = missing.length === 0;
			inputs = {};
			logLines = [];
			validationError = '';
			planLinkState = {};
			failedIds = [];
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

		validationError = '';
		logLines = [...logLines, { text: `Linking ${planDisplayName(plan)}...`, ok: false, loading: true }];
		try {
			await encore.tiers.UpdatePlanProviderID(planId, { provider_plan_id: providerPlanId });
			logLines = logLines.map(l =>
				l.text.startsWith(`Linking ${planDisplayName(plan)}`) ? { text: `✓ ${planDisplayName(plan)} linked`, ok: true, loading: false } : l
			);
			await loadPlans();
		} catch (err: any) {
			logLines = logLines.map(l =>
				l.text.startsWith(`Linking ${planDisplayName(plan)}`) ? { text: `✗ ${planDisplayName(plan)} failed: ${err?.message || 'unknown error'}`, ok: false, loading: false } : l
			);
		}
	}

	async function linkAllManual() {
		validationError = '';
		const emptyPlans = unlinked.filter(p => !inputs[p.id]?.trim());
		if (emptyPlans.length > 0) {
			validationError = `${emptyPlans.length} plan${emptyPlans.length === 1 ? '' : 's'} missing NowPayments ID.`;
			return;
		}

		logLines = [];
		for (const plan of unlinked) {
			const val = inputs[plan.id];
			if (val?.trim()) await linkPlan(plan.id, val.trim());
		}
	}

	async function linkAllAutomatic() {
		autoLinking = true;
		validationError = '';
		logLines = [];
		autoProgress = { done: 0, total: unlinked.length };
		planLinkState = {};
		failedIds = [];
		unlinked.forEach(p => { planLinkState[p.id] = 'pending'; });

		for (const plan of unlinked) {
			planLinkState[plan.id] = 'linking';
			logLines = [...logLines, { text: `Creating NowPayments plan for ${planDisplayName(plan)}...`, ok: false, loading: true }];
			try {
				const res = await encore.tiers.AutoLinkPlan(plan.id, { Host: window.location.host });
				planLinkState[plan.id] = 'success';
				logLines = logLines.map(l =>
					l.text.startsWith(`Creating NowPayments plan for ${planDisplayName(plan)}`)
						? { text: `✓ ${planDisplayName(plan)} auto-linked (ID ${res.provider_plan_id})`, ok: true, loading: false }
						: l
				);
			} catch (err: any) {
				planLinkState[plan.id] = 'failed';
				failedIds.push(plan.id);
				logLines = logLines.map(l =>
					l.text.startsWith(`Creating NowPayments plan for ${planDisplayName(plan)}`)
						? { text: `✗ ${planDisplayName(plan)} failed: ${err?.message || 'unknown error'}`, ok: false, loading: false }
						: l
				);
			}
			autoProgress = { ...autoProgress, done: autoProgress.done + 1 };
		}

		autoLinking = false;
		await loadPlans();
	}

	async function retryFailedPlans(planIds: string[]) {
		const plans = unlinked.filter(p => planIds.includes(p.id));
		autoLinking = true;
		logLines = [];
		autoProgress = { done: 0, total: plans.length };
		planLinkState = {};
		failedIds = [];
		plans.forEach(p => { planLinkState[p.id] = 'pending'; });

		for (const plan of plans) {
			planLinkState[plan.id] = 'linking';
			logLines = [...logLines, { text: `Retrying ${planDisplayName(plan)}...`, ok: false, loading: true }];
			try {
				const res = await encore.tiers.AutoLinkPlan(plan.id, { Host: window.location.host });
				planLinkState[plan.id] = 'success';
				logLines = logLines.map(l =>
					l.text.startsWith(`Retrying ${planDisplayName(plan)}`)
						? { text: `✓ ${planDisplayName(plan)} auto-linked (ID ${res.provider_plan_id})`, ok: true, loading: false }
						: l
				);
			} catch (err: any) {
				planLinkState[plan.id] = 'failed';
				failedIds.push(plan.id);
				logLines = logLines.map(l =>
					l.text.startsWith(`Retrying ${planDisplayName(plan)}`)
						? { text: `✗ ${planDisplayName(plan)} failed: ${err?.message || 'unknown error'}`, ok: false, loading: false }
						: l
				);
			}
			autoProgress = { ...autoProgress, done: autoProgress.done + 1 };
		}

		autoLinking = false;
		await loadPlans();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-60 bg-black/50 flex items-center justify-center p-4" onclick={onClose} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-4xl max-h-[85vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()} role="presentation">
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

					<div class="flex items-center gap-2">
						<span class="text-sm text-muted-foreground">Mode:</span>
						<label class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors
							{mode === 'manual' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
						>
							<input type="radio" bind:group={mode} value="manual" class="sr-only" />
							Manual
						</label>
						<label class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors
							{mode === 'automatic' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
						>
							<input type="radio" bind:group={mode} value="automatic" class="sr-only" />
							Automatic
						</label>
					</div>

					{#if validationError}
						<div class="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800">
							<AlertCircle class="size-4 text-red-600 mt-0.5 flex-shrink-0" />
							<p class="text-sm text-red-800 dark:text-red-200">{validationError}</p>
						</div>
					{/if}

					{#if mode === 'automatic' && autoLinking}
						<div class="space-y-1.5">
							<div class="flex items-center justify-between text-xs text-muted-foreground">
								<span>{autoProgress.done} of {autoProgress.total} done</span>
								<span>{Math.round((autoProgress.done / autoProgress.total) * 100)}%</span>
							</div>
							<div class="w-full h-2 bg-muted rounded-full overflow-hidden">
								<div
									class="h-full bg-primary rounded-full transition-all duration-300"
									style="width: {(autoProgress.done / autoProgress.total) * 100}%"
								></div>
							</div>
						</div>
					{/if}

					<div class="rounded-xl border border-border overflow-hidden">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b bg-muted/50 text-left">
									<th class="px-4 py-2.5 font-medium cursor-pointer hover:text-primary select-none" onclick={() => toggleSort('name')}>
										Plan{sortArrow('name')}
									</th>
									{#if mode === 'manual'}
										<th class="px-4 py-2.5 font-medium cursor-pointer hover:text-primary select-none" onclick={() => toggleSort('interval')}>
											Interval{sortArrow('interval')}
										</th>
									{/if}
									<th class="px-4 py-2.5 font-medium cursor-pointer hover:text-primary select-none" onclick={() => toggleSort('price')}>
										Price{sortArrow('price')}
									</th>
									{#if mode === 'manual'}
										<th class="px-4 py-2.5 font-medium">NowPayments ID</th>
									{/if}
									<th class="px-4 py-2.5 font-medium w-24"></th>
								</tr>
							</thead>
							<tbody class="divide-y divide-border">
								{#each sortedUnlinked as plan}
									<tr>
										<td class="px-4 py-2.5 font-medium">{planDisplayName(plan)}</td>
										{#if mode === 'manual'}
											<td class="px-4 py-2.5 text-muted-foreground capitalize">{plan.interval}</td>
										{/if}
										<td class="px-4 py-2.5 text-muted-foreground">{formatPrice(plan.price_usd_cents)}</td>
										{#if mode === 'manual'}
											<td class="px-4 py-2.5">
												<div class="flex items-center gap-2">
													<input
														type="text"
														placeholder="e.g. 12345"
														bind:value={inputs[plan.id]}
														class="w-28 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm"
														class:border-red-500={!inputs[plan.id]?.trim() && !!validationError}
													/>
													<Button size="sm" variant="outline" disabled={!inputs[plan.id]?.trim() || autoLinking} onclick={() => linkPlan(plan.id, inputs[plan.id])}>
														Link
													</Button>
												</div>
											</td>
										{:else}
											<td class="px-4 py-2.5">
												{#if planLinkState[plan.id] === 'linking'}
													<div class="flex items-center gap-1.5">
														<LoaderCircle class="size-3 animate-spin text-primary" />
														<div class="w-10 h-1 bg-muted rounded-full overflow-hidden">
															<div class="h-full bg-primary rounded-full animate-pulse"></div>
														</div>
													</div>
												{:else if planLinkState[plan.id] === 'success'}
													<CheckCircle class="size-4 text-green-500" />
												{:else if planLinkState[plan.id] === 'failed'}
													<AlertCircle class="size-4 text-red-500" />
												{:else}
													<span class="text-xs text-muted-foreground">ready</span>
												{/if}
											</td>
										{/if}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					{#if mode === 'manual' && unlinked.length > 0}
						<Button onclick={linkAllManual} disabled={autoLinking} class="w-full">
							Link All
						</Button>
					{:else if mode === 'automatic' && unlinked.length > 0}
						{#if failedIds.length > 0 && !autoLinking}
							<Button onclick={() => retryFailedPlans(failedIds)} disabled={autoLinking} class="w-full">
								Retry Failed ({failedIds.length})
							</Button>
						{:else}
							<Button onclick={linkAllAutomatic} disabled={autoLinking} class="w-full">
								{autoLinking ? 'Linking...' : 'Link All'}
							</Button>
						{/if}
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
