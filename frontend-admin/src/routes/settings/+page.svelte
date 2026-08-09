<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { AlertCircle, CheckCircle } from 'lucide-svelte';

	let { data } = $props();

	let settings = $state({
		default_language: 'en', default_content_language: 'en',
		items_per_page: 12, popular_tags_limit: 20, site_name: 'Comics Galore',
		maintenance_mode: false, registrations_open: true, max_upload_size_mb: 50,
		image_serving_mode: 'direct', require_email_verify: false,
		rate_limit: 60, s3_presigned_ttl_min: 15, cf_presigned_ttl_min: 15,
		quota_free_gb: 1, quota_bronze_gb: 10, quota_silver_gb: 50,
		quota_gold_gb: 200, quota_platinum_gb: 1000,
		boost_5gb_price: 5, boost_10gb_price: 8, boost_20gb_price: 12,
		contact_email: '', hide_mature_default: false, enable_comments: true,
		default_meta_description: '',
		...data.settings
	});

	let mode = $state<'form' | 'json'>('form');
	let jsonText = $state(JSON.stringify(settings, null, 2));
	let submitting = $state(false);
	let saved = $state(false);
	let error = $state('');

	function syncJson() {
		jsonText = JSON.stringify(settings, null, 2);
	}

	function switchMode(m: 'form' | 'json') {
		if (m === 'json') {
			syncJson();
			mode = 'json';
		} else {
			try {
				const parsed = JSON.parse(jsonText);
				settings = { ...settings, ...parsed };
				error = '';
				mode = 'form';
			} catch {
				error = 'Invalid JSON — fix the syntax before switching to Form mode.';
			}
		}
	}

	async function save() {
		submitting = true; saved = false; error = '';

		let payload = settings;
		if (mode === 'json') {
			try {
				payload = JSON.parse(jsonText);
			} catch {
				error = 'Invalid JSON. Please fix the syntax.';
				submitting = false;
				return;
			}
		}

		try {
			await encore.auth.SaveAdminSettings(payload);
			settings = payload;
			saved = true;
			setTimeout(() => saved = false, 2000);
		} catch (e: any) {
			error = e?.message || 'Save failed';
		}
		submitting = false;
	}
</script>

<svelte:head><title>Settings — Admin</title></svelte:head>

<section class="max-w-3xl">
	<h1 class="text-3xl font-bold mb-6">Settings</h1>

	<div class="flex items-center gap-2 mb-6">
		<span class="text-sm text-muted-foreground mr-1">View:</span>
		<label class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors
			{mode === 'form' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
			onclick={() => switchMode('form')}
		>
			Form
		</label>
		<label class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors
			{mode === 'json' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
			onclick={() => switchMode('json')}
		>
			JSON
		</label>
	</div>

	{#if mode === 'json'}
		<div class="space-y-4">
			<div class="space-y-1">
				<label class="text-xs text-muted-foreground">Edit settings as JSON</label>
				<textarea
					bind:value={jsonText}
					class="w-full min-h-[60vh] rounded-md border border-input bg-muted/50 px-4 py-3 text-sm font-mono resize-y"
					spellcheck="false"
				></textarea>
			</div>

			{#if error}
				<div class="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800">
					<AlertCircle class="size-4 text-red-600 mt-0.5 flex-shrink-0" />
					<p class="text-sm text-red-800 dark:text-red-200">{error}</p>
				</div>
			{/if}

			<div class="flex items-center gap-3">
				<Button onclick={save} disabled={submitting} class="min-w-[150px]">
					{submitting ? 'Saving...' : 'Save Settings'}
				</Button>
				{#if saved}
					<span class="flex items-center gap-1 text-sm text-green-500"><CheckCircle class="size-3.5" /> Saved!</span>
				{/if}
			</div>
		</div>
	{:else}
		<div class="space-y-4">
			<Card>
				<CardHeader><CardTitle>Site</CardTitle></CardHeader>
				<CardContent class="space-y-4">
					<div class="space-y-1">
						<label class="text-xs text-muted-foreground">Site Name</label>
						<input bind:value={settings.site_name} class="w-full max-w-md rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
					</div>
					<div class="space-y-1">
						<label class="text-xs text-muted-foreground">Contact Email</label>
						<input type="email" bind:value={settings.contact_email} placeholder="admin@comics-galore.com" class="w-full max-w-md rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
					</div>
					<div class="flex gap-6">
						<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.maintenance_mode} class="rounded" /> Maintenance Mode</label>
						<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.registrations_open} class="rounded" /> Registrations Open</label>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader><CardTitle>Content</CardTitle></CardHeader>
				<CardContent class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Default Language</label>
							<select bind:value={settings.default_language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="en">English</option><option value="ja">Japanese</option><option value="es">Spanish</option><option value="ko">Korean</option><option value="fr">French</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Default Content Language</label>
							<select bind:value={settings.default_content_language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="en">en</option><option value="ja">ja</option><option value="es">es</option><option value="ko">ko</option><option value="fr">fr</option>
							</select>
						</div>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Max Upload Size (MB)</label>
							<input type="number" bind:value={settings.max_upload_size_mb} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Image Serving Mode</label>
							<select bind:value={settings.image_serving_mode} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="direct">direct</option><option value="imgproxy">imgproxy</option><option value="cloudflare_images">cloudflare_images</option>
							</select>
						</div>
					</div>
					<div class="space-y-1">
						<label class="text-xs text-muted-foreground">Default Meta Description</label>
						<textarea bind:value={settings.default_meta_description} placeholder="SEO meta description for public pages..." class="w-full max-w-lg rounded-md border border-input bg-background px-3 py-1.5 text-sm resize-y" rows="2"></textarea>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader><CardTitle>Display</CardTitle></CardHeader>
				<CardContent class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Items Per Page</label>
							<select bind:value={settings.items_per_page} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value={12}>12</option><option value={24}>24</option><option value={48}>48</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Popular Tags Limit</label>
							<input type="number" bind:value={settings.popular_tags_limit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
					</div>
					<div class="flex gap-6">
						<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.hide_mature_default} class="rounded" /> Hide Mature Content by Default</label>
						<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.enable_comments} class="rounded" /> Enable Comments</label>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader><CardTitle>Quotas (GB)</CardTitle></CardHeader>
				<CardContent class="grid grid-cols-5 gap-3">
					<div class="text-center"><label class="text-[10px] text-muted-foreground">Free</label><input type="number" bind:value={settings.quota_free_gb} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div>
					<div class="text-center"><label class="text-[10px] text-muted-foreground">Bronze</label><input type="number" bind:value={settings.quota_bronze_gb} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div>
					<div class="text-center"><label class="text-[10px] text-muted-foreground">Silver</label><input type="number" bind:value={settings.quota_silver_gb} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div>
					<div class="text-center"><label class="text-[10px] text-muted-foreground">Gold</label><input type="number" bind:value={settings.quota_gold_gb} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div>
					<div class="text-center"><label class="text-[10px] text-muted-foreground">Platinum</label><input type="number" bind:value={settings.quota_platinum_gb} class="w-full rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader><CardTitle>Quota Boosts</CardTitle></CardHeader>
				<CardContent class="grid grid-cols-3 gap-4">
					<div class="text-center space-y-1"><label class="text-xs font-medium">+5 GB</label><div class="flex items-center gap-1 justify-center"><span class="text-xs text-muted-foreground">$</span><input type="number" bind:value={settings.boost_5gb_price} class="w-20 rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div></div>
					<div class="text-center space-y-1"><label class="text-xs font-medium">+10 GB</label><div class="flex items-center gap-1 justify-center"><span class="text-xs text-muted-foreground">$</span><input type="number" bind:value={settings.boost_10gb_price} class="w-20 rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div></div>
					<div class="text-center space-y-1"><label class="text-xs font-medium">+20 GB</label><div class="flex items-center gap-1 justify-center"><span class="text-xs text-muted-foreground">$</span><input type="number" bind:value={settings.boost_20gb_price} class="w-20 rounded-md border border-input bg-background px-2 py-1.5 text-sm text-center" /></div></div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader><CardTitle>Security</CardTitle></CardHeader>
				<CardContent class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Rate Limit (req/min)</label>
							<input type="number" bind:value={settings.rate_limit} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
					</div>
					<label class="flex items-center gap-2 text-sm"><input type="checkbox" bind:checked={settings.require_email_verify} class="rounded" /> Require Email Verification</label>
					<div class="grid grid-cols-2 gap-4">
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

			{#if error}
				<div class="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800">
					<AlertCircle class="size-4 text-red-600 mt-0.5 flex-shrink-0" />
					<p class="text-sm text-red-800 dark:text-red-200">{error}</p>
				</div>
			{/if}

			<div class="flex items-center gap-3 pt-2">
				<Button onclick={save} disabled={submitting} class="min-w-[150px]">
					{submitting ? 'Saving...' : 'Save Settings'}
				</Button>
				{#if saved}
					<span class="flex items-center gap-1 text-sm text-green-500"><CheckCircle class="size-3.5" /> Saved!</span>
				{/if}
			</div>
		</div>
	{/if}
</section>
