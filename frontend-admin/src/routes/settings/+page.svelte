<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { AlertCircle, CheckCircle, AlertTriangle } from 'lucide-svelte';

	let { data } = $props();

	let settings = $state({
		default_language: 'en', default_content_language: 'en',
		items_per_page: 12, popular_tags_limit: 20, site_name: 'Comics Galore',
		maintenance_mode: false, registrations_open: true, max_upload_size_mb: 50,
		image_serving_mode: 'direct', require_email_verify: false,
		rate_limit: 60, s3_presigned_ttl_min: 15, cf_presigned_ttl_min: 15,
		quota_free_gb: 1, quota_bronze_gb: 10, quota_silver_gb: 50,
		quota_gold_gb: 200, quota_platinum_gb: 1000,
		boost_1_gb: 5, boost_1_price: 5,
		boost_2_gb: 10, boost_2_price: 8,
		boost_3_gb: 20, boost_3_price: 12,
		contact_email: '', hide_mature_default: false, enable_comments: true,
		default_meta_description: '',
		// svelte-ignore state_referenced_locally
		...data.settings
	});

	let mode = $state<'form' | 'json'>('form');
	// svelte-ignore state_referenced_locally
	let jsonText = $state(JSON.stringify(settings, null, 2));
	let submitting = $state(false);
	let saved = $state(false);
	let error = $state('');

	let showUnlinkConfirm = $state(false);
	let unlinking = $state(false);
	let unlinkResult = $state('');

	async function unlinkAllPlans() {
		unlinking = true;
		unlinkResult = '';
		try {
			const res = await encore.tiers.UnlinkAllPlans();
			unlinkResult = `Unlinked ${res.count} plan(s).`;
		} catch (e: any) {
			unlinkResult = `Error: ${e?.message || 'unknown'}`;
		}
		unlinking = false;
		showUnlinkConfirm = false;
	}

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

<section>
	<h1 class="text-3xl font-bold mb-6">Settings</h1>

	<div class="flex items-center gap-2 mb-6">
		<span class="text-sm text-muted-foreground mr-1">View:</span>
		<button class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors border-0 bg-transparent
			{mode === 'form' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
			onclick={() => switchMode('form')}
		>
			Form
		</button>
		<button class="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors border-0 bg-transparent
			{mode === 'json' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'}"
			onclick={() => switchMode('json')}
		>
			JSON
		</button>
	</div>

	{#if mode === 'json'}
		<div class="space-y-4">
			<textarea
				bind:value={jsonText}
				class="w-full min-h-[60vh] rounded-md border border-input bg-muted/50 px-4 py-3 text-sm font-mono resize-y"
				spellcheck="false"
			></textarea>

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
			<div class="grid grid-cols-2 gap-4">
				<Card>
					<CardHeader class="pb-2"><CardTitle>Site</CardTitle></CardHeader>
					<CardContent class="space-y-2">
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Name</span>
							<input bind:value={settings.site_name} class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm" />
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Contact</span>
							<input type="email" bind:value={settings.contact_email} placeholder="admin@site.com" class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm" />
						</div>
						<div class="flex gap-4">
							<label class="flex items-center gap-1.5 text-xs"><input type="checkbox" bind:checked={settings.maintenance_mode} class="rounded" /> Maintenance</label>
							<label class="flex items-center gap-1.5 text-xs"><input type="checkbox" bind:checked={settings.registrations_open} class="rounded" /> Registrations</label>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-2"><CardTitle>Content</CardTitle></CardHeader>
					<CardContent class="space-y-2">
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Lang</span>
							<select bind:value={settings.default_language} class="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs">
								<option value="en">English</option><option value="ja">Japanese</option><option value="es">Spanish</option><option value="ko">Korean</option><option value="fr">French</option>
							</select>
							<select bind:value={settings.default_content_language} class="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs">
								<option value="en">en</option><option value="ja">ja</option><option value="es">es</option><option value="ko">ko</option><option value="fr">fr</option>
							</select>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Upload</span>
							<input type="number" bind:value={settings.max_upload_size_mb} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">MB</span>
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">Images</span>
							<select bind:value={settings.image_serving_mode} class="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs">
								<option value="direct">direct</option><option value="imgproxy">imgproxy</option><option value="cloudflare_images">CF images</option>
							</select>
						</div>
						<div class="flex items-start gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap pt-1.5">Meta</span>
							<textarea bind:value={settings.default_meta_description} placeholder="SEO meta description..." class="flex-1 rounded-md border border-input bg-background px-2.5 py-1 text-xs resize-y" rows="2"></textarea>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-2"><CardTitle>Display</CardTitle></CardHeader>
					<CardContent class="space-y-2">
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Per Page</span>
							<select bind:value={settings.items_per_page} class="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs">
								<option value={12}>12</option><option value={24}>24</option><option value={48}>48</option>
							</select>
							<span class="text-xs text-muted-foreground whitespace-nowrap">Tags</span>
							<input type="number" bind:value={settings.popular_tags_limit} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
						</div>
						<div class="flex gap-4">
							<label class="flex items-center gap-1.5 text-xs"><input type="checkbox" bind:checked={settings.hide_mature_default} class="rounded" /> Hide Mature</label>
							<label class="flex items-center gap-1.5 text-xs"><input type="checkbox" bind:checked={settings.enable_comments} class="rounded" /> Comments</label>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-2"><CardTitle>Security</CardTitle></CardHeader>
					<CardContent class="space-y-2">
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">Rate</span>
							<input type="number" bind:value={settings.rate_limit} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">req/min</span>
							<label class="flex items-center gap-1.5 text-xs ml-2"><input type="checkbox" bind:checked={settings.require_email_verify} class="rounded" /> Email Verify</label>
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">S3 TTL</span>
							<input type="number" bind:value={settings.s3_presigned_ttl_min} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">min</span>
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">CF TTL</span>
							<input type="number" bind:value={settings.cf_presigned_ttl_min} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">min</span>
						</div>
					</CardContent>
				</Card>
			</div>

			<Card>
				<CardHeader class="pb-2"><CardTitle>Quotas</CardTitle></CardHeader>
				<CardContent class="grid grid-cols-5 gap-2">
					<div class="text-center"><span class="text-[10px] text-muted-foreground">Free</span><input type="number" bind:value={settings.quota_free_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /><span class="text-[10px] text-muted-foreground">GB</span></div>
					<div class="text-center"><span class="text-[10px] text-muted-foreground">Bronze</span><input type="number" bind:value={settings.quota_bronze_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /><span class="text-[10px] text-muted-foreground">GB</span></div>
					<div class="text-center"><span class="text-[10px] text-muted-foreground">Silver</span><input type="number" bind:value={settings.quota_silver_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /><span class="text-[10px] text-muted-foreground">GB</span></div>
					<div class="text-center"><span class="text-[10px] text-muted-foreground">Gold</span><input type="number" bind:value={settings.quota_gold_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /><span class="text-[10px] text-muted-foreground">GB</span></div>
					<div class="text-center"><span class="text-[10px] text-muted-foreground">Platinum</span><input type="number" bind:value={settings.quota_platinum_gb} class="w-full rounded-md border border-input bg-background px-2 py-1 text-sm text-center" /><span class="text-[10px] text-muted-foreground">GB</span></div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="pb-2"><CardTitle>Quota Boosts</CardTitle></CardHeader>
				<CardContent class="space-y-2">
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 1</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_1_gb} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">GB</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_1_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 2</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_2_gb} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">GB</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_2_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 3</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_3_gb} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">GB</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_3_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
				</CardContent>
			</Card>

			<Card class="border-red-200 dark:border-red-800">
				<CardHeader class="pb-2"><CardTitle class="text-red-600 dark:text-red-400">NowPayments Integration</CardTitle></CardHeader>
				<CardContent class="space-y-2">
					<p class="text-sm text-muted-foreground">Remove all linked NowPayments plan IDs. Subscriptions stop working until plans are re-linked.</p>

					{#if showUnlinkConfirm}
						<div class="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800">
							<AlertTriangle class="size-4 text-red-600 mt-0.5 flex-shrink-0" />
							<div class="space-y-2 flex-1">
								<p class="text-xs text-red-800 dark:text-red-200">This will remove all NowPayments plan IDs from all plans.</p>
								<div class="flex gap-2">
									<Button variant="outline" size="sm" onclick={() => showUnlinkConfirm = false} disabled={unlinking}>Cancel</Button>
									<Button variant="destructive" size="sm" onclick={unlinkAllPlans} disabled={unlinking}>
										{unlinking ? 'Unlinking...' : 'Yes, Unlink All'}
									</Button>
								</div>
							</div>
						</div>
					{:else}
						<Button variant="destructive" onclick={() => showUnlinkConfirm = true}>
							Unlink All NowPayments Plans
						</Button>
					{/if}

					{#if unlinkResult}
						<p class="text-sm {unlinkResult.startsWith('Error') ? 'text-red-600' : 'text-green-600'}">{unlinkResult}</p>
					{/if}
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
