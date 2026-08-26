<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { AlertCircle, CheckCircle, LoaderCircle, Settings } from 'lucide-svelte';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('wizard'));

	type Plan = { id: string; name: string; tier_id: string; interval: string; price_usd_cents: number; provider_plan_id: string };

	let plans = $state<Plan[]>([]);
	let inputs = $state<Record<string, string>>({});
	let logLines = $state<Array<{ text: string; ok: boolean; loading: boolean }>>([]);
	let autoLinking = $state(false);
	let autoProgress = $state({ done: 0, total: 0 });
	let mode = $state<'manual' | 'automatic'>('automatic');
	let planLinkState = $state<Record<string, 'pending' | 'linking' | 'success' | 'failed'>>({});
	let failedIds = $state<string[]>([]);
	let confirmUnlinkId = $state<string | null>(null);
	let unlinkBusyId = $state<string | null>(null);
	let unlinkTimer: ReturnType<typeof setTimeout> | undefined;

	let sortKey = $state<string | null>(null);
	let sortDir = $state<'asc' | 'desc'>('asc');

	let validationError = $state('');

	let sortedPlans = $derived(sortPlans(plans, sortKey, sortDir));
	let unlinkedPlans = $derived(plans.filter((p) => !p.provider_plan_id));
	let allLinked = $derived(unlinkedPlans.length === 0);

	function close() {
		modal.close('wizard');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
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
			const paid = (res.plans || []).filter((p) => p.price_usd_cents > 0);
			plans = paid;
			inputs = {};
			logLines = [];
			validationError = '';
			planLinkState = {};
			failedIds = [];
			paid.forEach((p) => {
				inputs[p.id] = '';
			});
		} catch {
			plans = [];
		}
	}

	$effect(() => {
		if (open) loadPlans();
		else cancelUnlink();
	});

	async function linkPlan(planId: string, providerPlanId: string) {
		const plan = plans.find((p) => p.id === planId);
		if (!plan || !providerPlanId.trim()) return;

		validationError = '';
		logLines = [...logLines, { text: `Linking ${planDisplayName(plan)}...`, ok: false, loading: true }];
		try {
			await encore.tiers.UpdatePlanProviderID(planId, { provider_plan_id: providerPlanId });
			logLines = logLines.map((l) =>
				l.text.startsWith(`Linking ${planDisplayName(plan)}`) ? { text: `✓ ${planDisplayName(plan)} linked`, ok: true, loading: false } : l
			);
			await loadPlans();
		} catch (err: any) {
			logLines = logLines.map((l) =>
				l.text.startsWith(`Linking ${planDisplayName(plan)}`) ? { text: `✗ ${planDisplayName(plan)} failed: ${err?.message || 'unknown error'}`, ok: false, loading: false } : l
			);
		}
	}

	async function linkAllAutomatic() {
		autoLinking = true;
		validationError = '';
		logLines = [];
		autoProgress = { done: 0, total: unlinkedPlans.length };
		planLinkState = {};
		failedIds = [];
		unlinkedPlans.forEach((p) => {
			planLinkState[p.id] = 'pending';
		});

		for (const plan of unlinkedPlans) {
			planLinkState[plan.id] = 'linking';
			logLines = [...logLines, { text: `Creating NowPayments plan for ${planDisplayName(plan)}...`, ok: false, loading: true }];
			try {
				const res = await encore.tiers.AutoLinkPlan(plan.id);
				planLinkState[plan.id] = 'success';
				logLines = logLines.map((l) =>
					l.text.startsWith(`Creating NowPayments plan for ${planDisplayName(plan)}`)
						? { text: `✓ ${planDisplayName(plan)} auto-linked (ID ${res.provider_plan_id})`, ok: true, loading: false }
						: l
				);
			} catch (err: any) {
				planLinkState[plan.id] = 'failed';
				failedIds.push(plan.id);
				logLines = logLines.map((l) =>
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
		const targets = unlinkedPlans.filter((p) => planIds.includes(p.id));
		autoLinking = true;
		logLines = [];
		autoProgress = { done: 0, total: targets.length };
		planLinkState = {};
		failedIds = [];
		targets.forEach((p) => {
			planLinkState[p.id] = 'pending';
		});

		for (const plan of targets) {
			planLinkState[plan.id] = 'linking';
			logLines = [...logLines, { text: `Retrying ${planDisplayName(plan)}...`, ok: false, loading: true }];
			try {
				const res = await encore.tiers.AutoLinkPlan(plan.id);
				planLinkState[plan.id] = 'success';
				logLines = logLines.map((l) =>
					l.text.startsWith(`Retrying ${planDisplayName(plan)}`)
						? { text: `✓ ${planDisplayName(plan)} auto-linked (ID ${res.provider_plan_id})`, ok: true, loading: false }
						: l
				);
			} catch (err: any) {
				planLinkState[plan.id] = 'failed';
				failedIds.push(plan.id);
				logLines = logLines.map((l) =>
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

	function startUnlink(planId: string) {
		cancelUnlink();
		confirmUnlinkId = planId;
		unlinkTimer = setTimeout(() => {
			confirmUnlinkId = null;
		}, 4000);
	}

	function cancelUnlink() {
		if (unlinkTimer) clearTimeout(unlinkTimer);
		unlinkTimer = undefined;
		confirmUnlinkId = null;
	}

	async function confirmUnlink(planId: string) {
		cancelUnlink();
		unlinkBusyId = planId;
		try {
			await encore.tiers.UnlinkPlan(planId);
			await loadPlans();
		} catch (err: any) {
			logLines = [...logLines, { text: `✗ Unlink failed: ${err?.message || 'unknown error'}`, ok: false, loading: false }];
		} finally {
			unlinkBusyId = null;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-4xl max-h-[85vh] overflow-hidden flex flex-col" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b flex-shrink-0">
				<div class="flex items-center gap-2">
					<Settings class="size-5 text-primary" />
					<h2 class="text-lg font-semibold">Link Plans to NowPayments</h2>
				</div>
				<button onclick={close} class="p-1 hover:bg-muted rounded-lg" aria-label="Close">
					<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" x2="6" y1="6" y2="18"/><line x1="6" x2="18" y1="6" y2="18"/></svg>
				</button>
			</div>

			<div class="p-6 overflow-y-auto flex-1 space-y-4">
				{#if allLinked}
					<div class="flex items-start gap-2 p-3 rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800">
						<CheckCircle class="size-4 text-green-600 mt-0.5 flex-shrink-0" />
						<p class="text-sm text-green-800 dark:text-green-200">All paid plans are linked to NowPayments.</p>
					</div>
				{:else}
					<div class="flex items-start gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800">
						<AlertCircle class="size-4 text-amber-600 mt-0.5 flex-shrink-0" />
						<p class="text-sm text-amber-800 dark:text-amber-200">
							{unlinkedPlans.length} {unlinkedPlans.length === 1 ? 'plan needs' : 'plans need'} a NowPayments plan ID.
						</p>
					</div>
				{/if}

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
								<th class="px-4 py-2.5 font-medium cursor-pointer hover:text-primary select-none" onclick={() => toggleSort('interval')}>
									Interval{sortArrow('interval')}
								</th>
								<th class="px-4 py-2.5 font-medium cursor-pointer hover:text-primary select-none" onclick={() => toggleSort('price')}>
									Price{sortArrow('price')}
								</th>
								<th class="px-4 py-2.5 font-medium text-right">NowPayments</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-border">
							{#each sortedPlans as plan (plan.id)}
								<tr>
									<td class="px-4 py-2.5 font-medium">{planDisplayName(plan)}</td>
									<td class="px-4 py-2.5 text-muted-foreground capitalize">{plan.interval}</td>
									<td class="px-4 py-2.5 text-muted-foreground">{formatPrice(plan.price_usd_cents)}</td>
									<td class="px-4 py-2.5">
										{#if plan.provider_plan_id}
											<div class="flex items-center justify-end gap-2">
												<CheckCircle class="size-4 text-green-500 flex-shrink-0" />
												<span class="text-xs text-muted-foreground font-mono">{plan.provider_plan_id}</span>
												{#if confirmUnlinkId === plan.id}
													<Button size="sm" variant="destructive" onclick={() => confirmUnlink(plan.id)} disabled={unlinkBusyId === plan.id}>Confirm</Button>
													<Button size="sm" variant="outline" onclick={cancelUnlink}>Cancel</Button>
												{:else}
													<Button size="sm" variant="ghost" onclick={() => startUnlink(plan.id)} disabled={unlinkBusyId === plan.id}>Unlink</Button>
												{/if}
											</div>
										{:else if mode === 'manual'}
											<div class="flex items-center justify-end gap-2">
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
										{:else}
											<div class="flex items-center justify-end">
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
											</div>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if mode === 'automatic'}
					{#if failedIds.length > 0 && !autoLinking}
						<Button onclick={() => retryFailedPlans(failedIds)} disabled={autoLinking} class="w-full">
							Retry Failed ({failedIds.length})
						</Button>
					{:else}
						<Button onclick={linkAllAutomatic} disabled={autoLinking || unlinkedPlans.length === 0} class="w-full">
							{autoLinking ? 'Linking...' : 'Link All'}
						</Button>
					{/if}
				{/if}

				{#if logLines.length > 0}
					<div class="rounded-xl border border-border p-3 space-y-1 max-h-40 overflow-y-auto bg-muted/30">
						{#each logLines as line, i (i)}
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
			</div>
		</div>
	</div>
{/if}
