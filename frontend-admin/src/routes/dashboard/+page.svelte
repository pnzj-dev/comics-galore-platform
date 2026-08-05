<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Users, BookOpen, Clock, CreditCard, Download, Eye } from 'lucide-svelte';

	let stats = $state<any>(null);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/'); return; }
		try { stats = await api.get('/admin/stats'); } catch {}
		loading = false;
	});
</script>

<svelte:head><title>Dashboard — Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Dashboard</h1>

	{#if loading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _}
				<div class="rounded-xl border border-border p-6 animate-pulse space-y-2">
					<div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
					<div class="h-8 bg-gray-200 dark:bg-gray-700 rounded w-1/3"></div>
				</div>
			{/each}
		</div>
	{:else if stats}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<Card>
				<CardHeader>
					<div class="flex items-center gap-2">
						<Users class="size-4 text-muted-foreground" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Total Users</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold">{stats.total_users}</p></CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center gap-2">
						<BookOpen class="size-4 text-muted-foreground" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Total Comics</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold">{stats.total_comics}</p></CardContent>
			</Card>

			<Card class="border-yellow-200 dark:border-yellow-800">
				<CardHeader>
					<div class="flex items-center gap-2">
						<Clock class="size-4 text-yellow-500" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Pending Review</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold text-yellow-600">{stats.pending_comics}</p></CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center gap-2">
						<CreditCard class="size-4 text-muted-foreground" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Active Subscriptions</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold">{stats.active_subscriptions}</p></CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center gap-2">
						<Download class="size-4 text-muted-foreground" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Total Downloads</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold">{stats.total_downloads}</p></CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center gap-2">
						<Eye class="size-4 text-muted-foreground" />
						<CardTitle class="text-sm font-medium text-muted-foreground">Total Views</CardTitle>
					</div>
				</CardHeader>
				<CardContent><p class="text-3xl font-bold">{stats.total_views}</p></CardContent>
			</Card>
		</div>
	{:else}
		<p class="text-muted-foreground text-center py-12">Unable to load dashboard data.</p>
	{/if}
</section>
