<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();

	let settings = $state({
		language: 'en',
		contentLanguage: 'en',
		itemsPerPage: 12,
		popularTagsLimit: 20,
	});
	let saved = $state(false);

	function handleKeydown(e: KeyboardEvent) { if (e.key === 'Escape') onClose(); }

	async function save() {
		localStorage.setItem('cg-settings', JSON.stringify(settings));
		saved = true;
		setTimeout(() => saved = false, 2000);
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={onClose} role="dialog" tabindex="-1">
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
							<label class="text-xs text-muted-foreground">Default Language</label>
							<select bind:value={settings.language} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value="en">English</option>
								<option value="ja">Japanese</option>
								<option value="es">Spanish</option>
								<option value="ko">Korean</option>
								<option value="fr">French</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Content Language</label>
							<select bind:value={settings.contentLanguage} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
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
							<label class="text-xs text-muted-foreground">Items Per Page</label>
							<select bind:value={settings.itemsPerPage} class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm">
								<option value={12}>12</option>
								<option value={24}>24</option>
								<option value={48}>48</option>
							</select>
						</div>
						<div class="space-y-1">
							<label class="text-xs text-muted-foreground">Popular Tags Limit</label>
							<input type="number" bind:value={settings.popularTagsLimit} min="5" max="50" class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm" />
						</div>
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
