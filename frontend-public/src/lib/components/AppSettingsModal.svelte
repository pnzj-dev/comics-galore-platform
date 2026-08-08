<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { LoaderCircle, CheckCircle } from 'lucide-svelte';

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
		hide_mature: false,
	});
	let saved = $state(false);
	let saving = $state(false);
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
		saving = true;
		try {
			await api.patch('/me/preferences', settings);
			saved = true;
			setTimeout(() => saved = false, 2000);
		} catch {}
		saving = false;
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

			<div class="p-4 space-y-4">
				<!-- General -->
				<div class="rounded-xl border border-border p-3 space-y-3">
					<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">General</h3>
					<div class="space-y-2.5">
						<label class="flex items-center justify-between gap-3 text-sm">
							<span class="text-muted-foreground">Language</span>
							<select bind:value={settings.language} class="rounded-md border border-input bg-background px-2 py-1 text-sm w-32">
								<option value="en">English</option>
								<option value="ja">Japanese</option>
								<option value="es">Spanish</option>
								<option value="ko">Korean</option>
								<option value="fr">French</option>
							</select>
						</label>
						<label class="flex items-center justify-between gap-3 text-sm">
							<span class="text-muted-foreground">Content Language</span>
							<select bind:value={settings.content_language} class="rounded-md border border-input bg-background px-2 py-1 text-sm w-32">
								<option value="en">en</option>
								<option value="ja">ja</option>
								<option value="es">es</option>
								<option value="ko">ko</option>
								<option value="fr">fr</option>
							</select>
						</label>
						<label class="flex items-center justify-between gap-3 text-sm">
							<span class="text-muted-foreground">Items Per Page</span>
							<select bind:value={settings.items_per_page} class="rounded-md border border-input bg-background px-2 py-1 text-sm w-32">
								<option value={12}>12</option>
								<option value={24}>24</option>
								<option value={48}>48</option>
							</select>
						</label>
						<label class="flex items-center justify-between gap-3 text-sm">
							<span class="text-muted-foreground">Tags Limit</span>
							<input type="number" bind:value={settings.popular_tags_limit} min="5" max="50" class="w-20 rounded-md border border-input bg-background px-2 py-1 text-sm" />
						</label>
					</div>
				</div>

				<!-- Content -->
				<div class="rounded-xl border border-border p-3">
					<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-3">Content</h3>
					<label class="flex items-center justify-between gap-2 text-sm cursor-pointer">
						<span class="text-muted-foreground">Hide mature content</span>
						<input type="checkbox" bind:checked={settings.hide_mature} class="rounded" />
					</label>
				</div>

				<!-- Notifications -->
				<div class="rounded-xl border border-border p-3">
					<h3 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-3">Notifications</h3>
					<div class="space-y-2.5">
						<label class="flex items-center justify-between gap-2 text-sm cursor-pointer">
							<span class="text-muted-foreground">New comics from creators you follow</span>
							<input type="checkbox" bind:checked={settings.email_from_following} class="rounded" />
						</label>
						<label class="flex items-center justify-between gap-2 text-sm cursor-pointer">
							<span class="text-muted-foreground">Support ticket replies</span>
							<input type="checkbox" bind:checked={settings.email_support_replies} class="rounded" />
						</label>
						<label class="flex items-center justify-between gap-2 text-sm cursor-pointer">
							<span class="text-muted-foreground">Marketing emails and promotions</span>
							<input type="checkbox" bind:checked={settings.email_marketing} class="rounded" />
						</label>
						<label class="flex items-center justify-between gap-2 text-sm cursor-pointer">
							<span class="text-muted-foreground">In-app notifications</span>
							<input type="checkbox" bind:checked={settings.in_app_enabled} class="rounded" />
						</label>
					</div>
				</div>

				<Button onclick={save} class="w-full" disabled={saving}>
						{#if saving}
							<LoaderCircle class="size-4 animate-spin" />
							Saving...
						{:else if saved}
							<CheckCircle class="size-4 text-green-500" />
							Save Settings
						{:else}
							Save Settings
						{/if}
					</Button>
				</div>
			</div>
		</div>
{/if}
