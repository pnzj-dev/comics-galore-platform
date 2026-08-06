<script lang="ts">
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let settings = $state({
		siteName: 'Comics Galore',
		maintenanceMode: false,
		registrationsOpen: true,
		defaultLanguage: 'en',
		defaultContentLanguage: 'en',
		itemsPerPage: 12,
		popularTagsLimit: 20,
		maxUploadSizeMB: 50,
		imageServingMode: 'direct',
		requireEmailVerify: false,
		rateLimit: 60,
		s3PresignedTTL: 15,
		cfPresignedTTL: 15,
		quotaFreeGB: 1,
		quotaBronzeGB: 10,
		quotaSilverGB: 50,
		quotaGoldGB: 200,
		quotaPlatinumGB: 1000,
		boost5Price: 5,
		boost10Price: 8,
		boost20Price: 12,
	});
	let saved = $state(false);

	onMount(async () => {
		if (!$currentUser || $currentUser.role !== 'admin') { await goto('/login'); return; }
		const stored = localStorage.getItem('cg-admin-settings');
		if (stored) settings = { ...settings, ...JSON.parse(stored) };
	});

	async function save() {
		localStorage.setItem('cg-admin-settings', JSON.stringify(settings));
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
				<input bind:value={settings.siteName} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<div class="flex gap-6">
				<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.maintenanceMode} class="rounded" /> Maintenance Mode</label>
				<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.registrationsOpen} class="rounded" /> Registrations Open</label>
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Content</CardTitle></CardHeader>
		<CardContent class="space-y-3">
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">Default Language</label>
					<select bind:value={settings.defaultLanguage} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
						<option value="en">English</option>
						<option value="ja">Japanese</option>
						<option value="es">Spanish</option>
						<option value="ko">Korean</option>
						<option value="fr">French</option>
					</select>
				</div>
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">Default Content Language</label>
					<select bind:value={settings.defaultContentLanguage} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
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
				<input type="number" bind:value={settings.maxUploadSizeMB} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Image Serving Mode</label>
				<select bind:value={settings.imageServingMode} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
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
				<select bind:value={settings.itemsPerPage} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
					<option value={12}>12</option>
					<option value={24}>24</option>
					<option value={48}>48</option>
				</select>
			</div>
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Popular Tags Limit</label>
				<input type="number" bind:value={settings.popularTagsLimit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Quotas (GB)</CardTitle></CardHeader>
		<CardContent class="grid grid-cols-5 gap-2">
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Free</label><input type="number" bind:value={settings.quotaFreeGB} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Bronze</label><input type="number" bind:value={settings.quotaBronzeGB} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Silver</label><input type="number" bind:value={settings.quotaSilverGB} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Gold</label><input type="number" bind:value={settings.quotaGoldGB} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
			<div class="text-center"><label class="text-[10px] text-muted-foreground">Platinum</label><input type="number" bind:value={settings.quotaPlatinumGB} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /></div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Quota Boosts</CardTitle></CardHeader>
		<CardContent class="grid grid-cols-3 gap-3">
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+5 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost5Price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+10 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost10Price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
			<div class="text-center space-y-1">
				<label class="text-xs font-medium">+20 GB</label>
				<div class="flex items-center gap-1 justify-center">
					<span class="text-xs text-muted-foreground">$</span>
					<input type="number" bind:value={settings.boost20Price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
				</div>
			</div>
		</CardContent>
	</Card>

	<Card class="mt-4">
		<CardHeader><CardTitle>Security</CardTitle></CardHeader>
		<CardContent class="space-y-3">
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Rate Limit (req/min)</label>
				<input type="number" bind:value={settings.rateLimit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
			</div>
			<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.requireEmailVerify} class="rounded" /> Require Email Verification</label>
			<div class="grid grid-cols-2 gap-3">
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">S3 Presigned TTL (min)</label>
					<input type="number" bind:value={settings.s3PresignedTTL} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
				</div>
				<div class="space-y-1">
					<label class="text-xs text-muted-foreground">CF Presigned TTL (min)</label>
					<input type="number" bind:value={settings.cfPresignedTTL} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
				</div>
			</div>
		</CardContent>
	</Card>

	<div class="flex items-center gap-3 mt-6">
		<Button onclick={save} size="lg">Save Settings</Button>
		{#if saved}<span class="text-sm text-green-500">Saved!</span>{/if}
	</div>
</section>
