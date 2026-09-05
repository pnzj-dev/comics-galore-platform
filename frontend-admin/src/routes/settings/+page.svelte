<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { AlertCircle, CheckCircle, AlertTriangle } from 'lucide-svelte';
	import type { auth } from '$lib/api/encore-client';

	let { data } = $props();

	let settings = $state<auth.AppSettings>({
		default_language: 'en', default_content_language: 'en',
		items_per_page: 12, popular_tags_limit: 20, site_name: 'Comics Galore',
		maintenance_mode: false, registrations_open: true, max_upload_size_mb: 3000,
		image_serving_mode: 'direct', upload_mode: 'backend', require_email_verify: false,
		imgproxy_base_url: '', imgproxy_key: '', imgproxy_salt: '',
		rate_limit: 60, s3_presigned_ttl_min: 15, cf_presigned_ttl_min: 15,
		boost_1_downloads: 10, boost_1_price: 5,
		boost_2_downloads: 25, boost_2_price: 10,
		boost_3_downloads: 60, boost_3_price: 20,
		contact_email: '', hide_mature_default: false, enable_comments: false,
		forbid_mature_for_free: false,
		default_meta_description: '',
		ai_moderation_enabled: false,
		ai_model: 'gpt-4o-mini',
		ai_endpoint: 'https://api.openai.com/v1/chat/completions',
		ai_prompt: 'You moderate user-generated content on a comics platform. Reply with only JSON: {"decision":"approved|rejected|uncertain","confidence":0.0,"reason":"..."}.',
		ai_auto_approve_threshold: 0.85,
		ai_auto_reject_threshold: 0.15,
		waiting_pay_job_enabled: true,
		waiting_pay_expiry_hours: 24,
		download_stream_threshold_mb: 10,
		page_preview_threshold: 20,
		upload_part_size_mb: 100,
		upload_concurrency: 4,
		crypto_currencies: 'btc,xrp,eth,usdttrc20,usdtsol,usdc,ltc,sol,xlm,pyusd',
		// svelte-ignore state_referenced_locally
		...(data.settings ?? {}),
	});

	let mode = $state<'form' | 'json'>('form');
	// svelte-ignore state_referenced_locally
	let jsonText = $state(JSON.stringify(settings, null, 2));
	let submitting = $state(false);
	let saved = $state(false);
	let error = $state('');

	// Tier download quotas (stored on the tiers table, not AppSettings).
	let tiers = $state<{ id: string; name: string; quota_downloads: number }[]>([]);
	let tierQuotas = $state<Record<string, number>>({});
	let quotaSaving = $state(false);
	let quotaSaved = $state(false);
	let quotaError = $state('');

	async function loadTiers() {
		try {
			const res = await encore.tiers.ListTiers();
			tiers = res.tiers || [];
			const q: Record<string, number> = {};
			for (const t of tiers) q[t.id] = t.quota_downloads;
			tierQuotas = q;
		} catch {}
	}

	async function saveTierQuotas() {
		quotaSaving = true; quotaSaved = false; quotaError = '';
		try {
			await encore.tiers.AdminUpdateTierQuotas({
				quotas: tiers.map((t) => ({ tier_id: t.id, quota_downloads: tierQuotas[t.id] ?? t.quota_downloads })),
			});
			quotaSaved = true;
			setTimeout(() => quotaSaved = false, 2000);
			await loadTiers();
		} catch (e: any) {
			quotaError = e?.message || 'Failed to save tier quotas';
		}
		quotaSaving = false;
	}

	onMount(() => { loadTiers(); });

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
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">Upload</span>
							<select bind:value={settings.upload_mode} class="flex-1 rounded-md border border-input bg-background px-2 py-1.5 text-xs">
								<option value="backend">backend</option><option value="direct">direct</option>
							</select>
						</div>
						<div class="flex items-start gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap pt-1.5">Meta</span>
							<textarea bind:value={settings.default_meta_description} placeholder="SEO meta description..." class="flex-1 rounded-md border border-input bg-background px-2.5 py-1 text-xs resize-y" rows="2"></textarea>
						</div>
						<div class="flex items-start gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap pt-1.5">Currencies</span>
							<input bind:value={settings.crypto_currencies} placeholder="btc,eth,usdttrc20,…" class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-xs" />
						</div>
						<div class="flex items-center gap-2">
							<span class="text-xs text-muted-foreground whitespace-nowrap">DL stream</span>
							<input type="number" bind:value={settings.download_stream_threshold_mb} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">MB</span>
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">Page previews</span>
							<input type="number" bind:value={settings.page_preview_threshold} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">Part size</span>
							<input type="number" bind:value={settings.upload_part_size_mb} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
							<span class="text-xs text-muted-foreground">MB</span>
							<span class="text-xs text-muted-foreground whitespace-nowrap ml-2">Concurrency</span>
							<input type="number" bind:value={settings.upload_concurrency} class="w-16 rounded-md border border-input bg-background px-2 py-1.5 text-sm" />
						</div>
						{#if settings.image_serving_mode === 'imgproxy'}
							<div class="space-y-2 pt-2 border-t border-border">
								<div class="flex items-center gap-2">
									<span class="text-xs text-muted-foreground whitespace-nowrap">imgproxy URL</span>
									<input bind:value={settings.imgproxy_base_url} placeholder="https://imgproxy.example.com" class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm" />
								</div>
								<div class="flex items-center gap-2">
									<span class="text-xs text-muted-foreground whitespace-nowrap">Key</span>
									<input bind:value={settings.imgproxy_key} placeholder="hex-encoded key" class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm" />
								</div>
								<div class="flex items-center gap-2">
									<span class="text-xs text-muted-foreground whitespace-nowrap">Salt</span>
									<input bind:value={settings.imgproxy_salt} placeholder="hex-encoded salt" class="flex-1 rounded-md border border-input bg-background px-2.5 py-1.5 text-sm" />
								</div>
							</div>
						{/if}
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
							<label class="flex items-center gap-1.5 text-xs"><input type="checkbox" bind:checked={settings.forbid_mature_for_free} class="rounded" /> Forbid mature for free users</label>
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
				<CardHeader class="pb-2"><CardTitle>Tier download quotas (downloads/month)</CardTitle></CardHeader>
				<CardContent class="space-y-3">
					{#each tiers as tier (tier.id)}
						<div class="flex items-center gap-3">
							<span class="text-xs text-muted-foreground w-20">{tier.name}</span>
							<input
								type="number"
								min="0"
								bind:value={tierQuotas[tier.id]}
								class="w-32 rounded-md border border-input bg-background px-2 py-1 text-sm text-center"
							/>
							{#if tierQuotas[tier.id] >= 999999}
								<span class="text-[10px] text-muted-foreground">unlimited</span>
							{/if}
						</div>
					{/each}
					<div class="flex items-center gap-3 pt-1">
						<Button size="sm" onclick={saveTierQuotas} disabled={quotaSaving}>
							{quotaSaving ? 'Saving…' : 'Save tier quotas'}
						</Button>
						{#if quotaSaved}
							<span class="flex items-center gap-1 text-sm text-green-500"><CheckCircle class="size-3.5" /> Saved!</span>
						{/if}
						{#if quotaError}
							<span class="flex items-center gap-1 text-sm text-red-500"><AlertCircle class="size-3.5" /> {quotaError}</span>
						{/if}
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="pb-2"><CardTitle>Quota Boosts</CardTitle></CardHeader>
				<CardContent class="space-y-2">
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 1</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_1_downloads} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">dl</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_1_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 2</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_2_downloads} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">dl</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_2_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-14">Boost 3</span>
						<span class="text-xs text-muted-foreground">+</span>
						<input type="number" bind:value={settings.boost_3_downloads} class="w-14 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
						<span class="text-xs text-muted-foreground">dl</span>
						<span class="text-xs text-muted-foreground ml-4">$</span>
						<input type="number" bind:value={settings.boost_3_price} class="w-16 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="pb-2"><CardTitle>Billing hygiene</CardTitle></CardHeader>
				<CardContent class="space-y-2">
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<input type="checkbox" bind:checked={settings.waiting_pay_job_enabled} class="rounded" />
						Downgrade WAITING_PAY subscriptions to free
					</label>
					<div class="flex items-center gap-2">
						<span class="text-xs text-muted-foreground w-40">Expire after (hours)</span>
						<input type="number" bind:value={settings.waiting_pay_expiry_hours} class="w-20 rounded-md border border-input bg-background px-2 py-1 text-sm text-center" />
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
