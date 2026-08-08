<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();

	let settings = $state({
		language: 'en',
		content_language: 'en',
		items_per_page: 12,
		popular_tags_limit: 20,
		email_from_following: true,
		email_support_replies: true,
		email_marketing: false,
		in_app_enabled: true,
	});
	let saved = $state(false);
	let loading = $state(true);

	async function load() {
		try {
			const res = await api.get<typeof settings>('/me/preferences');
			settings = res;
		} catch {}
		loading = false;
	}

	onMount(() => { if (open) load(); });

	$effect(() => { if (open) { loading = true; load(); } });

	async function save() {
		await api.patch('/me/preferences', settings);
		saved = true;
		setTimeout(() => saved = false, 2000);
	}

	function handleKeydown(e: KeyboardEvent) { if (e.key === 'Escape') onClose(); }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={onClose} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Settings</h2>
				<button onclick={onClose} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-5">
				<div>
					<h3 class="text-sm font-medium mb-3">Language</h3>
					<div class="grid grid-cols-2 gap-3">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-language">Default Language</label>
							<select id="settings-language" bind:value={settings.language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="en">English</option>
								<option value="ja">Japanese</option>
								<option value="es">Spanish</option>
								<option value="ko">Korean</option>
								<option value="fr">French</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-content-language">Content Language</label>
							<select id="settings-content-language" bind:value={settings.content_language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="en">en</option>
								<option value="ja">ja</option>
								<option value="es">es</option>
								<option value="ko">ko</option>
								<option value="fr">fr</option>
							</select>
						</div>
					</div>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Display</h3>
					<div class="grid grid-cols-2 gap-3">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-items-per-page">Items Per Page</label>
							<select id="settings-items-per-page" bind:value={settings.items_per_page} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value={12}>12</option>
								<option value={24}>24</option>
								<option value={48}>48</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-popular-tags-limit">Popular Tags Limit</label>
							<input id="settings-popular-tags-limit" type="number" bind:value={settings.popular_tags_limit} min="5" max="50" class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
					</div>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Notifications</h3>
					<div class="space-y-2">
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<input type="checkbox" bind:checked={settings.email_from_following} class="rounded" />
							New comics from creators you follow
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<input type="checkbox" bind:checked={settings.email_support_replies} class="rounded" />
							Support ticket replies
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<input type="checkbox" bind:checked={settings.email_marketing} class="rounded" />
							Marketing emails and promotions
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<input type="checkbox" bind:checked={settings.in_app_enabled} class="rounded" />
							In-app notifications
						</label>
					</div>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Quota Boosts</h3>
					<div class="grid grid-cols-3 gap-2">
						<div class="rounded-lg border border-border p-2 text-center">
							<p class="text-xs font-medium">+5 GB</p>
							<p class="text-xs text-muted-foreground">$5.00</p>
						</div>
						<div class="rounded-lg border border-border p-2 text-center">
							<p class="text-xs font-medium">+10 GB</p>
							<p class="text-xs text-muted-foreground">$8.00</p>
						</div>
						<div class="rounded-lg border border-border p-2 text-center">
							<p class="text-xs font-medium">+20 GB</p>
							<p class="text-xs text-muted-foreground">$12.00</p>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-3 pt-2">
					<Button onclick={save} class="w-full">Save Settings</Button>
					{#if saved}<span class="text-sm text-green-500 flex-shrink-0">Saved!</span>{/if}
				</div>
			</div>
		</div>
	</div>
{/if}
