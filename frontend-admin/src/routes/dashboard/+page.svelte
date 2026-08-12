<script lang="ts">
	import { barY, defineChart } from '@tanstack/charts';
	import { scaleLinear } from '@tanstack/charts/scales/linear';
	import { scalePoint } from '@tanstack/charts/scales/point';
	import { Chart } from '@tanstack/svelte-charts';
	import KpiCard from '$lib/components/KpiCard.svelte';
	import { formatNum, formatCurrency, formatBytes } from '$lib/utils/format';
	import {
		Users, UserPlus, DollarSign, TrendingUp, CreditCard,
		BookOpen, Download, Eye, HardDrive, Image, AlertTriangle,
	} from 'lucide-svelte';

	let { data } = $props();

	const revenueData = $derived(
		(data.billing?.revenue_by_tier ?? []).map((r) => ({ tier: r.tier, revenue: r.revenue / 100 }))
	);

	const revenueChart = $derived(
		defineChart({
			marks: [barY(revenueData, { x: 'tier', y: 'revenue' })],
			x: { scale: () => scalePoint<string>().padding(0.4) },
			y: { scale: scaleLinear, nice: true, grid: true, axis: { label: 'Revenue (USD)' } },
			svgAnimation: true,
		})
	);

	const topLiked = $derived(data.comics?.top_liked ?? []);
</script>

<svelte:head><title>Dashboard — Admin</title></svelte:head>

<section class="space-y-6">
	<h1 class="text-3xl font-bold">Dashboard</h1>

	{#if !data.users && !data.comics && !data.billing && !data.reading && !data.storage}
		<p class="text-muted-foreground text-center py-12">Unable to load dashboard data.</p>
	{:else}
		<!-- KPI Cards -->
		<div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-4">
			<KpiCard title="Total Users" value={formatNum(data.users?.total_users)} icon={Users} />
			<KpiCard title="New Users (MTD)" value={formatNum(data.users?.new_users_this_month)} icon={UserPlus} />
			<KpiCard title="Active Subscriptions" value={formatNum(data.billing?.active_subscriptions)} icon={CreditCard} />
			<KpiCard title="MRR" value={formatCurrency(data.billing?.active_revenue)} icon={DollarSign} accent />
			<KpiCard title="Total Revenue" value={formatCurrency(data.billing?.total_revenue)} icon={TrendingUp} />
			<KpiCard title="Revenue (30d)" value={formatCurrency(data.billing?.recent_revenue)} icon={TrendingUp} />
			<KpiCard title="Total Comics" value={formatNum(data.comics?.total_comics)} icon={BookOpen} />
			<KpiCard title="Published" value={formatNum(data.comics?.published_comics)} icon={BookOpen} />
			<KpiCard title="Pending Review" value={formatNum(data.comics?.pending_comics)} icon={AlertTriangle} />
			<KpiCard title="Total Views" value={formatNum(data.comics?.total_views)} icon={Eye} />
			<KpiCard title="Total Downloads" value={formatNum(data.reading?.total_downloads)} icon={Download} />
			<KpiCard title="Total Payments" value={formatNum(data.billing?.total_payments)} icon={CreditCard} />
			<KpiCard title="Total Deposits" value={formatNum(data.billing?.total_deposits)} icon={DollarSign} />
			<KpiCard title="Storage Used" value={formatBytes(data.comics?.storage_bytes)} icon={HardDrive} />
			<KpiCard title="Cloudflare Images" value={data.storage?.cf_configured ? formatNum(data.storage?.cf_images_count) : 'N/A'} icon={Image} />
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
								<span class="text-xs font-medium text-primary">{formatNum(comic.like_count)} likes</span>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-sm text-muted-foreground text-center py-12">No likes yet.</p>
				{/if}
			</div>
		</div>
	{/if}
</section>
