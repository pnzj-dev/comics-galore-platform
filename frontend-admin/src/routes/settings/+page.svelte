<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let settings = $state({
		default_language: 'en',
		default_content_language: 'en',
		items_per_page: 12,
		popular_tags_limit: 20,
		site_name: 'Comics Galore',
		maintenance_mode: false,
		registrations_open: true,
		max_upload_size_mb: 50,
		image_serving_mode: 'direct',
		require_email_verify: false,
		rate_limit: 60,
		s3_presigned_ttl_min: 15,
		cf_presigned_ttl_min: 15,
		quota_free_gb: 1,
		quota_bronze_gb: 10,
		quota_silver_gb: 50,
		quota_gold_gb: 200,
		quota_platinum_gb: 1000,
		boost_5gb_price: 5,
		boost_10gb_price: 8,
		boost_20gb_price: 12,
	});
	let saved = $state(false);
	let loading = $state(true);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		try {
			const res = await api.get<typeof settings>('/admin/settings');
			settings = { ...settings, ...res };
		} catch {}
		loading = false;
	});

	async function save() {
		await api.patch('/admin/settings', settings);
		saved = true;
		setTimeout(() => saved = false, 2000);
	}
</script>

<svelte:head><title>Settings — Admin</title></svelte:head>

<section class="py-8 max-w-2xl mx-auto">
	<h1 class="text-3xl font-bold mb-6">Settings</h1>

	<Card>
		<CardHeader><CardTitle>Site</CardTitle></CardHeader>
		<CardContent class="space-y-3">
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Site Name</label>
				<input bind:value={settings.site_name} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<div class="flex gap-6">
				<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.maintenance_mode} class="rounded" /> Maintenance Mode</label>
				<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.registrations_open} class="rounded" /> Registrations Open</label>
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Content</CardTitle></CardHeader>
		<CardContent class="space-y-3">
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">Default Language</label>
					<select bind:value={settings.default_language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
						<option value="en">English</option>
						<option value="ja">Japanese</option>
						<option value="es">Spanish</option>
						<option value="ko">Korean</option>
						<option value="fr">French</option>
					</select>
				</div>
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">Default Content Language</label>
					<select bind:value={settings.default_content_language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
						<option value="en">en</option>
						<option value="ja">ja</option>
						<option value="es">es</option>
						<option value="ko">ko</option>
						<option value="fr">fr</option>
					</select>
				</div>
			</div>
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Max Upload Size (MB)</label>
				<input type="number" bind:value={settings.max_upload_size_mb} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Image Serving Mode</label>
				<select bind:value={settings.image_serving_mode} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
					<option value="direct">direct</option>
					<option value="imgproxy">imgproxy</option>
					<option value="cloudflare_images">cloudflare_images</option>
				</select>
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Display</CardTitle></CardHeader>
		<CardContent class="grid grid-cols-2 gap-3">
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Items Per Page</label>
				<select bind:value={settings.items_per_page} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
					<option value={12}>12</option>
					<option value={24}>24</option>
					<option value={48}>48</option>
				</select>
			</div>
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Popular Tags Limit</label>
				<input type="number" bind:value={settings.popular_tags_limit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Quotas (GB)</CardTitle></CardHeader>
		<CardContent class="grid grid-cols-5 gap-2">
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Free</label><input type="number" bind:value={settings.quota_free_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Bronze</label><input type="number" bind:value={settings.quota_bronze_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Silver</label><input type="number" bind:value={settings.quota_silver_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Gold</label><input type="number" bind:value={settings.quota_gold_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Platinum</label><input type="number" bind:value={settings.quota_platinum_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Quota Boosts</CardTitle></CardHeader>
		<CardContent class="grid grid-cols-3 gap-3">
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+5 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost_5gb_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+10 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost_10gb_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+20 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost_20gb_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Security</CardTitle></CardHeader>
		<CardContent class="space-y-3">
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Rate Limit (req/min)</label>
				<input type="number" bind:value={settings.rate_limit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.require_email_verify} class="rounded" /> Require Email Verification</label>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">S3 Presigned TTL (min)</label>
					<input type="number" bind:value={settings.s3_presigned_ttl_min} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
				</div>
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">CF Presigned TTL (min)</label>
					<input type="number" bind:value={settings.cf_presigned_ttl_min} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
				</div>
			</div>
		</CardContent>
	</Card>

	<div class="flex items-center gap-3 mt-6">
		<Button onclick={save} size="lg">Save Settings</Button>
		{#if saved}<span class="text-sm text-green-500">Saved!</span>{/if}
	</div>
</section>
