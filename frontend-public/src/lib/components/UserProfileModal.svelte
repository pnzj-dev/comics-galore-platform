<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import CheckoutModal from '$lib/components/CheckoutModal.svelte';
	import { onMount } from 'svelte';
	import { Camera, LoaderCircle } from 'lucide-svelte';

	let { open, onClose }: { open: boolean; onClose: () => void } = $props();
	let checkoutOpen = $state(false);
	let plansReady = $state(false);
	let uploadingAvatar = $state(false);
	let avatarKey = $state('');

	const user = $derived($currentUser);

	const MEDIA_BASE = import.meta.env.VITE_API_URL || 'http://localhost:4000';
	const avatarSrc = $derived(avatarKey ? `${MEDIA_BASE}/media/${avatarKey}` : '');

	onMount(async () => {
		try {
			const [ready, profile] = await Promise.all([
				api.get<{ complete: boolean }>('/plans/ready'),
				api.get<{ avatar_key: string }>('/me/avatar')
			]);
			plansReady = ready.complete;
			avatarKey = profile.avatar_key || '';
		} catch {}
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}

	function formatDate(d: string): string {
		if (!d) return '';
		return new Date(d).toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
	}

	async function handleAvatarChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		if (file.size > 500 * 1024) { alert('Image too large (max 500KB)'); return; }

		const reader = new FileReader();
		reader.onload = async () => {
			uploadingAvatar = true;
			try {
				const res = await api.post<{ avatar_key: string }>('/me/avatar', { avatar_data: reader.result });
				avatarKey = res.avatar_key;
			} catch (err) {
				alert((err as Error).message || 'Upload failed');
			}
			uploadingAvatar = false;
		};
		reader.readAsDataURL(file);
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={onClose} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-sm max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>
			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Profile</h2>
				<button onclick={onClose} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-6 flex flex-col items-center space-y-4">
				<!-- Avatar -->
				<label class="relative cursor-pointer group">
					<div class="size-20 rounded-full {avatarSrc ? '' : 'bg-purple-100 dark:bg-purple-900'} flex items-center justify-center overflow-hidden border-2 border-border hover:border-primary/50 transition-colors">
						{#if avatarSrc}
							<img src={avatarSrc} alt="Avatar" class="w-full h-full object-cover" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
						{/if}
						<span class="text-2xl font-semibold text-purple-600 dark:text-purple-300 {avatarSrc ? 'hidden' : ''}">{user?.email?.charAt(0)?.toUpperCase() || '?'}</span>
					</div>
					<div class="absolute inset-0 rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
						{#if uploadingAvatar}
							<LoaderCircle class="size-5 text-white animate-spin" />
						{:else}
							<Camera class="size-5 text-white" />
						{/if}
					</div>
					<input type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="hidden" onchange={handleAvatarChange} />
				</label>

				<!-- User info -->
				<div class="text-center">
					<p class="text-sm font-semibold">{user?.email || '—'}</p>
					<p class="text-xs text-muted-foreground">Member since {formatDate(user?.created_at || '')}</p>
				</div>

				<!-- Badges -->
				<div class="flex gap-2">
					{#if user?.role !== 'admin'}
						<span class="text-xs rounded-full px-2 py-0.5 capitalize bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Tier: {user?.tier || 'free'}</span>
					{/if}
					<span class="text-xs rounded-full px-2 py-0.5 bg-primary/10 text-primary">Role: {user?.role || 'user'}</span>
				</div>

				<!-- Subscription card -->
				<div class="w-full rounded-xl border border-border p-3 space-y-2">
					<p class="text-xs font-medium">Subscription</p>
					<p class="text-xs text-muted-foreground">{user?.tier === 'free' ? 'Free plan · No active subscription' : `Active ${user?.tier} plan`}</p>
					{#if plansReady}
						<Button size="sm" class="w-full mt-1" onclick={() => checkoutOpen = true}>
							{user?.tier === 'free' ? 'Upgrade Plan' : 'Manage Subscription'}
						</Button>
					{:else}
						<p class="text-xs text-muted-foreground italic">Subscriptions being configured — check back soon.</p>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<CheckoutModal open={checkoutOpen} onClose={() => checkoutOpen = false} />
