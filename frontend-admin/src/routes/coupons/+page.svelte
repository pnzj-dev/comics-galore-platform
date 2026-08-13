<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let coupons = $state(data.coupons);
	let code = $state('');
	let percentOff = $state(20);
	let tier = $state('');
	let maxUses = $state(0);
	let error = $state('');
	let creating = $state(false);

	async function createCoupon() {
		if (!code.trim()) {
			error = 'Code is required';
			return;
		}
		creating = true;
		error = '';
		try {
			await encore.billing.AdminCreateCoupon({ code, percent_off: percentOff, tier, max_uses: maxUses });
			code = '';
			const list = await encore.billing.AdminListCoupons();
			coupons = list.coupons || [];
		} catch (e) {
			error = (e as Error).message;
		} finally {
			creating = false;
		}
	}
</script>

<svelte:head><title>Coupons - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Coupons</h1>

	<Card class="mb-6">
		<CardHeader><CardTitle>New coupon</CardTitle></CardHeader>
		<CardContent class="flex flex-wrap items-end gap-3">
			<div class="space-y-1.5">
				<label class="text-xs text-muted-foreground">Code</label>
				<Input bind:value={code} placeholder="SUMMER20" />
			</div>
			<div class="space-y-1.5">
				<label class="text-xs text-muted-foreground">Percent off</label>
				<Input type="number" bind:value={percentOff} min={1} max={100} />
			</div>
			<div class="space-y-1.5">
				<label class="text-xs text-muted-foreground">Tier (optional)</label>
				<Input bind:value={tier} placeholder="gold" />
			</div>
			<div class="space-y-1.5">
				<label class="text-xs text-muted-foreground">Max uses (0=unlimited)</label>
				<Input type="number" bind:value={maxUses} min={0} />
			</div>
			<Button onclick={createCoupon} disabled={creating}>{creating ? 'Creating…' : 'Create'}</Button>
			{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
		</CardContent>
	</Card>

	{#if coupons.length === 0}
		<p class="text-muted-foreground text-sm">No coupons yet.</p>
	{:else}
		<div class="rounded-lg border border-border overflow-hidden">
			<table class="w-full text-sm">
				<thead class="bg-muted/50">
					<tr>
						<th class="text-left px-3 py-2">Code</th>
						<th class="text-left px-3 py-2">Percent</th>
						<th class="text-left px-3 py-2">Tier</th>
						<th class="text-left px-3 py-2">Uses</th>
					</tr>
				</thead>
				<tbody class="divide-y">
					{#each coupons as c}
						<tr>
							<td class="px-3 py-2 font-medium">{c.code}</td>
							<td class="px-3 py-2">{c.percent_off}%</td>
							<td class="px-3 py-2">{c.tier || 'all'}</td>
							<td class="px-3 py-2">{c.used}{c.max_uses > 0 ? `/${c.max_uses}` : ''}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
