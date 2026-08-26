<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { LoaderCircle, CheckCircle } from 'lucide-svelte';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('settings'));

	let prefs = $state({
		language: 'en',
		content_language: 'en',
		items_per_page: 12,
		popular_tags_limit: 20,
		hide_mature: false,
	});
	let notifications = $state({
		email_new_from_following: true,
		email_support_replies: true,
		email_marketing: false,
		in_app_enabled: true,
	});
	let saved = $state(false);
	let saving = $state(false);
	let loading = $state(true);

	const languages = [
		{ value: 'en', label: 'English' },
		{ value: 'ja', label: 'Japanese' },
		{ value: 'es', label: 'Spanish' },
		{ value: 'ko', label: 'Korean' },
		{ value: 'fr', label: 'French' },
	];

	const languageLabel = $derived(languages.find((l) => l.value === prefs.language)?.label ?? 'Select language');

	async function load() {
		try {
			const [p, n] = await Promise.all([
				encore.auth.GetPreferences(),
				encore.auth.GetNotificationPrefs(),
			]);
			prefs = { ...prefs, ...p };
			notifications = { ...notifications, ...n };
		} catch {}
		loading = false;
	}

	$effect(() => { if (open) { loading = true; load(); } });

	async function save() {
		saving = true;
		try {
			await Promise.all([
				encore.auth.SavePreferences(prefs),
				encore.auth.UpdateNotificationPrefs(notifications),
			]);
			saved = true;
			setTimeout(() => saved = false, 2000);
		} catch {}
		saving = false;
	}

	function close() {
		modal.close('settings');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) { if (e.key === 'Escape') close(); }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Preferences</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-4">
				<Card>
					<CardHeader class="pb-2"><CardTitle>General</CardTitle></CardHeader>
					<CardContent class="space-y-4">
						<div class="grid grid-cols-2 gap-3">
							<div class="space-y-1">
								<label class="text-xs text-muted-foreground" for="settings-language">Default Language</label>
								<Select type="single" bind:value={prefs.language}>
									<SelectTrigger id="settings-language" class="w-full">
										{languageLabel}
									</SelectTrigger>
									<SelectContent>
										{#each languages as lang (lang.value)}
											<SelectItem value={lang.value}>{lang.label}</SelectItem>
										{/each}
									</SelectContent>
								</Select>
							</div>
							<div class="space-y-1">
								<label class="text-xs text-muted-foreground" for="settings-content-language">Content Language</label>
								<Select type="single" bind:value={prefs.content_language}>
									<SelectTrigger id="settings-content-language" class="w-full">
										{prefs.content_language}
									</SelectTrigger>
									<SelectContent>
										{#each languages as lang (lang.value)}
											<SelectItem value={lang.value}>{lang.value}</SelectItem>
										{/each}
									</SelectContent>
								</Select>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-3">
							<div class="space-y-1">
								<label class="text-xs text-muted-foreground" for="settings-items-per-page">Items Per Page</label>
								<Select type="single" value={String(prefs.items_per_page)} onValueChange={(v) => prefs.items_per_page = Number(v)}>
									<SelectTrigger id="settings-items-per-page" class="w-full">
										{prefs.items_per_page}
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="12">12</SelectItem>
										<SelectItem value="24">24</SelectItem>
										<SelectItem value="48">48</SelectItem>
									</SelectContent>
								</Select>
							</div>
							<div class="space-y-1">
								<label class="text-xs text-muted-foreground" for="settings-popular-tags-limit">Popular Tags Limit</label>
								<input id="settings-popular-tags-limit" type="number" bind:value={prefs.popular_tags_limit} min="5" max="50" class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
							</div>
						</div>

						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={prefs.hide_mature} />
							Hide mature content
						</label>
					</CardContent>
				</Card>

				<Card>
					<CardHeader class="pb-2"><CardTitle>Notifications</CardTitle></CardHeader>
					<CardContent class="space-y-2">
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={notifications.email_new_from_following} />
							New comics from creators you follow
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={notifications.email_support_replies} />
							Support ticket replies
						</label>
						<label class="flex items-center gap-2 text-sm text-muted-foreground cursor-not-allowed opacity-60">
							<Checkbox bind:checked={notifications.email_marketing} disabled />
							Marketing emails and promotions
							<span class="text-[10px]">(coming soon)</span>
						</label>
						<label class="flex items-center gap-2 text-sm text-muted-foreground cursor-not-allowed opacity-60">
							<Checkbox bind:checked={notifications.in_app_enabled} disabled />
							In-app notifications
							<span class="text-[10px]">(coming soon)</span>
						</label>
					</CardContent>
				</Card>

				<div class="flex items-center gap-3 pt-1">
					<Button onclick={save} class="w-full" disabled={saving || loading}>
						{#if saving}
							<LoaderCircle class="size-4 animate-spin" />
							Saving...
						{:else if saved}
							<CheckCircle class="size-4 text-green-500" />
							Save Changes
						{:else}
							Save Changes
						{/if}
					</Button>
				</div>
			</div>
		</div>
	</div>
{/if}
