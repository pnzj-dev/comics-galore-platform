<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { formatDate } from '$lib/utils/format';
	import { LoaderCircle, CheckCircle2, XCircle } from 'lucide-svelte';
	import type { billing } from '$lib/api/encore-client';

	let { data } = $props();

	type WType = 'subscription' | 'deposit';

	const STATUS_OPTIONS = ['finished', 'waiting', 'partially_paid', 'expired', 'failed', 'waiting_pay', 'paid'];

	let type = $state<WType>('subscription');
	let status = $state('finished');
	let dryRun = $state(false);
	let selectedId = $state('');
	let submitting = $state(false);
	let result = $state<billing.SimulateWebhookResponse | null>(null);
	let error = $state('');

	const records = $derived(type === 'subscription' ? data.subscriptions : data.deposits);

	function select(id: string) {
		selectedId = id;
	}

	async function simulate() {
		if (!selectedId) return;
		submitting = true;
		error = '';
		result = null;
		try {
			const res = await encore.billing.SimulateWebhook({
				type,
				id: selectedId,
				status,
				dry_run: dryRun,
			});
			result = res;
		} catch (e) {
			error = (e as Error).message || 'simulation failed';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>Webhook Simulator - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Webhook Simulator</h1>
	<p class="text-sm text-muted-foreground mb-6 max-w-2xl">
		Replay a NowPayments webhook for an existing deposit or subscription. The payload is signed with your
		IPN secret and run through the same processing path (signature verification, status update, tier activation)
		as a real callback. Use it to test a subscription manually.
	</p>

	<div class="grid gap-6 lg:grid-cols-[1fr_1.2fr] items-start">
		<!-- Form -->
		<Card>
			<CardHeader class="pb-3"><CardTitle>Simulation</CardTitle></CardHeader>
			<CardContent class="space-y-4">
				<div class="flex items-center gap-2">
					<span class="text-sm text-muted-foreground w-24">Type</span>
					<button
						type="button"
						onclick={() => { type = 'subscription'; selectedId = ''; result = null; error = ''; }}
						class="px-3 py-1.5 rounded-lg text-sm transition-colors {type === 'subscription' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
					>
						Subscription
					</button>
					<button
						type="button"
						onclick={() => { type = 'deposit'; selectedId = ''; result = null; error = ''; }}
						class="px-3 py-1.5 rounded-lg text-sm transition-colors {type === 'deposit' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
					>
						Deposit
					</button>
				</div>

				<div class="flex items-center gap-2">
					<span class="text-sm text-muted-foreground w-24">Status</span>
					<select bind:value={status} class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm">
						{#each STATUS_OPTIONS as s}<option value={s}>{s}</option>{/each}
					</select>
				</div>

				<label class="flex items-center gap-2 text-sm cursor-pointer">
					<input type="checkbox" bind:checked={dryRun} class="size-4" />
					Dry run (build + sign only, do not process)
				</label>

				<div>
					<p class="text-sm font-medium mb-2">Select a {type} record</p>
					<div class="rounded-lg border border-border max-h-64 overflow-y-auto">
						{#if records.length === 0}
							<p class="text-sm text-muted-foreground text-center py-6">No {type}s found.</p>
						{:else}
							<table class="w-full text-sm">
								<thead class="sticky top-0 bg-muted/60">
									<tr class="text-left text-xs text-muted-foreground">
										<th class="px-3 py-2 font-medium">Created</th>
										<th class="px-3 py-2 font-medium">Tier / Currency</th>
										<th class="px-3 py-2 font-medium">Status</th>
										<th class="px-3 py-2 font-medium">Provider ID</th>
									</tr>
								</thead>
								<tbody class="divide-y divide-border">
									{#each records as r (r.id)}
										<tr
											onclick={() => select(r.id)}
											class="cursor-pointer transition-colors {selectedId === r.id ? 'bg-primary/10' : 'hover:bg-muted'}"
										>
											<td class="px-3 py-2 text-xs text-muted-foreground whitespace-nowrap">{formatDate(r.created_at, 'datetime')}</td>
											<td class="px-3 py-2 text-xs">{type === 'subscription' ? (r as any).tier : (r as any).currency_crypto?.toUpperCase()}</td>
											<td class="px-3 py-2 text-xs">
												<span class="px-1.5 py-0.5 rounded-full text-[10px] font-medium {r.status === 'active' || r.status === 'completed' ? 'bg-green-100 text-green-700' : 'bg-muted text-muted-foreground'}">{r.status}</span>
											</td>
											<td class="px-3 py-2 text-[10px] font-mono text-muted-foreground truncate max-w-[140px]">{type === 'subscription' ? (r as any).provider_subscription_id || '—' : (r as any).provider_deposit_id || '—'}</td>
										</tr>
									{/each}
								</tbody>
							</table>
						{/if}
					</div>
				</div>

				<Button class="w-full" onclick={simulate} disabled={!selectedId || submitting}>
					{#if submitting}<LoaderCircle class="size-4 animate-spin" />{/if}
					Simulate Webhook
				</Button>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
			</CardContent>
		</Card>

		<!-- Result -->
		<div class="space-y-4">
			{#if !result}
				<Card>
					<CardContent class="py-10 text-center text-sm text-muted-foreground">
						Run a simulation to see the request, response and outcome here.
					</CardContent>
				</Card>
			{:else}
				<Card>
					<CardHeader class="pb-3"><CardTitle>Request</CardTitle></CardHeader>
					<CardContent class="space-y-2 text-sm">
						<div class="grid grid-cols-[90px_1fr] gap-y-1">
							<span class="text-muted-foreground">Method</span><span class="font-mono">{result.request.method}</span>
							<span class="text-muted-foreground">Path</span><span class="font-mono">{result.request.path}</span>
							<span class="text-muted-foreground">Query</span><span class="font-mono break-all">{JSON.stringify(result.request.query) === '{}' ? '(none)' : JSON.stringify(result.request.query)}</span>
							<span class="text-muted-foreground">Signature</span><span class="font-mono break-all text-[11px]">{result.request.signature}</span>
						</div>
						<div>
							<p class="text-muted-foreground mb-1">Body (payload)</p>
							<pre class="text-xs bg-muted rounded-lg p-3 overflow-x-auto">{JSON.stringify(result.request.payload, null, 2)}</pre>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-3"><CardTitle>Response</CardTitle></CardHeader>
					<CardContent class="space-y-2 text-sm">
						<div class="flex items-center gap-2">
							{#if result.response.status_code >= 200 && result.response.status_code < 300}
								<CheckCircle2 class="size-4 text-green-500" />
							{:else}
								<XCircle class="size-4 text-red-500" />
							{/if}
							<span class="font-medium">Status {result.response.status_code}</span>
						</div>
						<pre class="text-xs bg-muted rounded-lg p-3 overflow-x-auto">{result.response.body}</pre>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-3"><CardTitle>Outcome</CardTitle></CardHeader>
					<CardContent class="space-y-1.5 text-sm">
						{#if type === 'subscription'}
							<div class="flex items-center gap-2">
								{#if result.outcome.subscription_active}<CheckCircle2 class="size-4 text-green-500" />{:else}<XCircle class="size-4 text-muted-foreground" />{/if}
								<span>subscription active → <strong>{result.outcome.subscription_active}</strong></span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-muted-foreground">subscription status →</span>
								<strong>{result.outcome.subscription_status || '—'}</strong>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								{#if result.outcome.deposit_completed}<CheckCircle2 class="size-4 text-green-500" />{:else}<XCircle class="size-4 text-muted-foreground" />{/if}
								<span>deposit completed → <strong>{result.outcome.deposit_completed}</strong></span>
							</div>
							<div class="flex items-center gap-2">
								{#if result.outcome.boost_granted}<CheckCircle2 class="size-4 text-green-500" />{:else}<XCircle class="size-4 text-muted-foreground" />{/if}
								<span>boost granted → <strong>{result.outcome.boost_granted}</strong></span>
							</div>
						{/if}
					</CardContent>
				</Card>
			{/if}
		</div>
	</div>
</section>
