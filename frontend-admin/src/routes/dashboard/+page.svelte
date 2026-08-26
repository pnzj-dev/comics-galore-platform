<script lang="ts">
	import { barY, defineChart } from '@tanstack/charts';
	import { scaleLinear } from '@tanstack/charts/scales/linear';
	import { scalePoint } from '@tanstack/charts/scales/point';
	import { Chart } from '@tanstack/svelte-charts';
	import KpiCard from '$lib/components/KpiCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { formatNumber, formatCurrency, formatBytes } from '$lib/utils/format';
	import type { dashboard } from '$lib/server/client';
	import {
		Users, UserPlus, DollarSign, TrendingUp, CreditCard,
		BookOpen, Download, Eye, HardDrive, Image, AlertTriangle,
	} from 'lucide-svelte';

	let { data } = $props();

	// Live stats: seeded from SSR, updated by the SSE stream when realtime is on.
	// svelte-ignore state_referenced_locally
	let live = $state<dashboard.DashboardResponse | null>(data.dashboard);
	let realtime = $state(true);

	$effect(() => {
		if (!realtime) return;
		const es = new EventSource('/api/admin/dashboard-stream');
		es.onmessage = (e) => {
			try {
				live = JSON.parse(e.data);
			} catch {
				/* ignore malformed frames */
			}
		};
		es.onerror = () => {
			// EventSource auto-reconnects; nothing to do here.
		};
		return () => es.close();
	});

	const revenueData = $derived(
		(live?.billing?.revenue_by_tier ?? []).map((r) => ({ tier: r.tier, revenue: r.revenue / 100 }))
	);

	const revenueChart = $derived(
		defineChart({
			marks: [barY(revenueData, { x: 'tier', y: 'revenue' })],
			x: { scale: () => scalePoint<string>().padding(0.4) },
			y: { scale: scaleLinear, nice: true, grid: true, axis: { label: 'Revenue (USD)' } },
			svgAnimation: true,
		})
	);

	const topLiked = $derived(live?.comics?.top_liked ?? []);
	const topViewed = $derived(live?.comics?.top_viewed ?? []);

	const statusData = $derived.by(() => {
		const published = live?.comics?.published_comics ?? 0;
		const pending = live?.comics?.pending_comics ?? 0;
		const rejected = live?.comics?.rejected_comics ?? 0;
		return [
			{ label: 'Published', value: published, color: 'bg-emerald-500' },
			{ label: 'Pending', value: pending, color: 'bg-yellow-500' },
			{ label: 'Rejected', value: rejected, color: 'bg-red-500' },
		];
	});
	const statusTotal = $derived(statusData.reduce((s, x) => s + x.value, 0));

	const downloadsData = $derived((live?.download_trend ?? []).map((p) => ({ day: p.day.slice(5), count: p.count })));
	const downloadsChart = $derived(
		defineChart({
			marks: [barY(downloadsData, { x: 'day', y: 'count' })],
			x: { scale: () => scalePoint<string>().padding(0.4) },
			y: { scale: scaleLinear, nice: true, grid: true },
			svgAnimation: true,
		})
	);

	const signupsData = $derived((live?.signup_trend ?? []).map((p) => ({ day: p.day.slice(5), count: p.count })));
	const signupsChart = $derived(
		defineChart({
			marks: [barY(signupsData, { x: 'day', y: 'count' })],
			x: { scale: () => scalePoint<string>().padding(0.4) },
			y: { scale: scaleLinear, nice: true, grid: true },
			svgAnimation: true,
		})
	);
</script>

<svelte:head><title>Dashboard — Admin</title></svelte:head>

<section class="space-y-6">
	<div class="flex items-center justify-between gap-4">
		<h1 class="text-3xl font-bold">Dashboard</h1>
		<Button size="sm" variant={realtime ? 'default' : 'outline'} onclick={() => (realtime = !realtime)}>
			<span class="inline-flex items-center gap-1.5">
				<span class="size-2 rounded-full {realtime ? 'bg-emerald-400 animate-pulse' : 'bg-muted-foreground'}"></span>
				Realtime: {realtime ? 'On' : 'Off'}
			</span>
		</Button>
	</div>

	{#if !live}
		<p class="text-muted-foreground text-center py-12">Unable to load dashboard data.</p>
	{:else}
		<!-- KPI Cards -->
		<div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-4">
			<KpiCard title="Total Users" value={formatNumber(live.users?.total_users)} icon={Users} />
			<KpiCard title="New Users (MTD)" value={formatNumber(live.users?.new_users_this_month)} icon={UserPlus} />
			<KpiCard title="Active Subscriptions" value={formatNumber(live.billing?.active_subscriptions)} icon={CreditCard} />
			<KpiCard title="MRR" value={formatCurrency(live.billing?.active_revenue)} icon={DollarSign} accent />
			<KpiCard title="Total Revenue" value={formatCurrency(live.billing?.total_revenue)} icon={TrendingUp} />
			<KpiCard title="Revenue (30d)" value={formatCurrency(live.billing?.recent_revenue)} icon={TrendingUp} />
			<KpiCard title="Total Comics" value={formatNumber(live.comics?.total_comics)} icon={BookOpen} />
			<KpiCard title="Published" value={formatNumber(live.comics?.published_comics)} icon={BookOpen} />
			<KpiCard title="Pending Review" value={formatNumber(live.comics?.pending_comics)} icon={AlertTriangle} />
			<KpiCard title="Total Views" value={formatNumber(live.comics?.total_views)} icon={Eye} />
			<KpiCard title="Total Downloads" value={formatNumber(live.reading?.total_downloads)} icon={Download} />
			<KpiCard title="Total Payments" value={formatNumber(live.billing?.total_payments)} icon={CreditCard} />
			<KpiCard title="Total Deposits" value={formatNumber(live.billing?.total_deposits)} icon={DollarSign} />
			<KpiCard title="Storage Used" value={formatBytes(live.comics?.storage_bytes)} icon={HardDrive} />
			<KpiCard title="Cloudflare Images" value={live.storage?.cf_configured ? formatNumber(live.storage?.cf_images_count) : 'N/A'} icon={Image} />
		</div>

		<!-- Charts row -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Revenue by Tier</h2>
				{#if revenueData.length > 0}
					<Chart definition={revenueChart} height={260} ariaLabel="Revenue by tier" />
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No revenue data yet.</p>
				{/if}
			</div>

			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Top Liked Comics</h2>
				{#if topLiked.length > 0}
					<div class="space-y-2">
						{#each topLiked as comic, i}
							<div class="flex items-center gap-3 py-2 border-b border-border/50 last:border-0">
								<span class="text-lg font-bold text-muted-foreground w-6">{i + 1}</span>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium truncate">{comic.title}</p>
								</div>
								<span class="text-xs font-medium text-primary">{formatNumber(comic.like_count)} likes</span>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No likes yet.</p>
				{/if}
			</div>
		</div>

		<!-- Trend charts -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Downloads (last 30 days)</h2>
				{#if downloadsData.length > 0}
					<Chart definition={downloadsChart} height={240} ariaLabel="Downloads last 30 days" />
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No downloads yet.</p>
				{/if}
			</div>

			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Signups (last 30 days)</h2>
				{#if signupsData.length > 0}
					<Chart definition={signupsChart} height={240} ariaLabel="Signups last 30 days" />
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No signups yet.</p>
				{/if}
			</div>
		</div>

		<!-- Status + top viewed -->
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Comic Status Distribution</h2>
				<div class="space-y-3">
					{#each statusData as s (s.label)}
						<div>
							<div class="flex items-center justify-between text-sm mb-1">
								<span class="text-muted-foreground">{s.label}</span>
								<span class="font-medium">{formatNumber(s.value)}</span>
							</div>
							<div class="h-2 rounded-full bg-muted overflow-hidden">
								<div class="h-full {s.color}" style="width: {statusTotal > 0 ? (s.value / statusTotal) * 100 : 0}%"></div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div class="rounded-xl border border-border p-4">
				<h2 class="text-sm font-medium mb-4">Top Viewed Comics</h2>
				{#if topViewed.length > 0}
					<div class="space-y-2">
						{#each topViewed as comic, i}
							<div class="flex items-center gap-3 py-2 border-b border-border/50 last:border-0">
								<span class="text-lg font-bold text-muted-foreground w-6">{i + 1}</span>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium truncate">{comic.title}</p>
								</div>
								<span class="text-xs font-medium text-primary">{formatNumber(comic.view_count)} views</span>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No views yet.</p>
				{/if}
			</div>
		</div>
	{/if}
</section>
