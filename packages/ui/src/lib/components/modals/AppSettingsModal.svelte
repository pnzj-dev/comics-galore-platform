<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { LoaderCircle, CheckCircle } from 'lucide-svelte';

	let { onClose }: { onClose?: () => void } = $props();

	const open = $derived(modal.isOpen('settings'));

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

	const languages = [
		{ value: 'en', label: 'English' },
		{ value: 'ja', label: 'Japanese' },
		{ value: 'es', label: 'Spanish' },
		{ value: 'ko', label: 'Korean' },
		{ value: 'fr', label: 'French' },
	];

	const languageLabel = $derived(languages.find((l) => l.value === settings.language)?.label ?? 'Select language');

	async function load() {
		try {
			const res = await encore.auth.GetPreferences();
			settings = { ...settings, ...res };
		} catch {}
		loading = false;
	}

	$effect(() => { if (open) { loading = true; load(); } });

	async function save() {
		saving = true;
		try {
			await encore.auth.SavePreferences(settings);
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
				<h2 class="text-lg font-semibold">Settings</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-5">
				<div>
					<h3 class="text-sm font-medium mb-3">Language</h3>
					<div class="grid grid-cols-2 gap-3">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-language">Default Language</label>
							<Select type="single" bind:value={settings.language}>
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
							<Select type="single" bind:value={settings.content_language}>
								<SelectTrigger id="settings-content-language" class="w-full">
									{settings.content_language}
								</SelectTrigger>
								<SelectContent>
									{#each languages as lang (lang.value)}
										<SelectItem value={lang.value}>{lang.value}</SelectItem>
									{/each}
								</SelectContent>
							</Select>
						</div>
					</div>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Display</h3>
					<div class="grid grid-cols-2 gap-3">
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground" for="settings-items-per-page">Items Per Page</label>
							<Select type="single" value={String(settings.items_per_page)} onValueChange={(v) => settings.items_per_page = Number(v)}>
								<SelectTrigger id="settings-items-per-page" class="w-full">
									{settings.items_per_page}
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
							<input id="settings-popular-tags-limit" type="number" bind:value={settings.popular_tags_limit} min="5" max="50" class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
					</div>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Content</h3>
					<label class="flex items-center gap-2 text-sm cursor-pointer">
						<Checkbox bind:checked={settings.hide_mature} />
						Hide mature content
					</label>
				</div>

				<div>
					<h3 class="text-sm font-medium mb-3">Notifications</h3>
					<div class="space-y-2">
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={settings.email_from_following} />
							New comics from creators you follow
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={settings.email_support_replies} />
							Support ticket replies
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={settings.email_marketing} />
							Marketing emails and promotions
						</label>
						<label class="flex items-center gap-2 text-sm cursor-pointer">
							<Checkbox bind:checked={settings.in_app_enabled} />
							In-app notifications
						</label>
					</div>
				</div>

				<div class="flex items-center gap-3 pt-2">
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
	</div>
{/if}
