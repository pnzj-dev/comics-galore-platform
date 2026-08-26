<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Avatar, AvatarImage, AvatarFallback } from '$lib/components/ui/avatar/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { modal } from '$lib/stores/modal.svelte';
	import { formatDate } from '$lib/utils/format';
	import { Camera, LoaderCircle, CheckCircle } from 'lucide-svelte';

	let { onClose }: { onClose?: () => void } = $props();
	let plansReady = $state(false);
	let uploadingAvatar = $state(false);
	let avatarKey = $state('');

	let originalUsername = $state('');
	let username = $state('');
	let availability = $state<'idle' | 'checking' | 'available' | 'taken' | 'invalid'>('idle');
	let savingUsername = $state(false);
	let usernameSaved = $state(false);

	let subscription = $state<{ id: string; tier: string; status: string } | null>(null);
	let subscriptionLoading = $state(true);
	let cancelling = $state(false);
	let confirmCancel = $state(false);

	const open = $derived(modal.isOpen('profile'));
	const user = $derived($currentUser);

	const MEDIA_BASE = import.meta.env.VITE_API_URL || 'http://localhost:4000';
	const avatarSrc = $derived(avatarKey ? `${MEDIA_BASE}/media/${avatarKey}` : '');

	const usernameRe = /^[a-z0-9](?:[_-]?[a-z0-9])*$/;
	const usernameChanged = $derived(username.trim().toLowerCase() !== originalUsername);
	const canSaveUsername = $derived(
		usernameChanged &&
		(availability === 'available' || (username.trim().toLowerCase() === originalUsername)) &&
		!savingUsername
	);

	$effect(() => {
		if (open) loadProfile();
	});

	$effect(() => {
		const u = username.trim().toLowerCase();
		if (u.length === 0 || u === originalUsername) {
			availability = u === originalUsername ? 'idle' : 'idle';
			return;
		}
		if (u.length < 3 || u.length > 20 || !usernameRe.test(u)) {
			availability = 'invalid';
			return;
		}
		availability = 'checking';
		const timer = setTimeout(async () => {
			try {
				const res = await encore.auth.UsernameAvailable({ Username: u });
				availability = res.available ? 'available' : (res.valid ? 'taken' : 'invalid');
			} catch {
				availability = 'idle';
			}
		}, 300);
		return () => clearTimeout(timer);
	});

	async function loadProfile() {
		try {
			const [ready, avatar, profile, sub] = await Promise.all([
				encore.tiers.PlansReady(),
				encore.auth.GetAvatar(),
				encore.auth.GetProfile(),
				encore.billing.GetMySubscription(),
			]);
			plansReady = ready.complete;
			avatarKey = avatar.avatar_key || '';
			originalUsername = profile.username || '';
			username = profile.username || '';
			subscription = sub;
		} catch {}
		subscriptionLoading = false;
	}

	function close() {
		modal.close('profile');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	async function handleAvatarChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		if (file.size > 500 * 1024) { alert('Image too large (max 500KB)'); return; }

		const reader = new FileReader();
		reader.onload = async () => {
			uploadingAvatar = true;
			try {
				const res = await encore.auth.UpdateAvatar({ avatar_data: reader.result as string });
				avatarKey = res.avatar_key;
			} catch (err) {
				alert((err as Error).message || 'Upload failed');
			}
			uploadingAvatar = false;
		};
		reader.readAsDataURL(file);
	}

	async function saveUsername() {
		if (!canSaveUsername) return;
		savingUsername = true;
		usernameSaved = false;
		try {
			const profile = await encore.auth.UpdateUsername({ username: username.trim().toLowerCase() });
			originalUsername = profile.username || '';
			username = profile.username || '';
			usernameSaved = true;
			setTimeout(() => usernameSaved = false, 2000);
		} catch (err) {
			alert((err as Error).message || 'Could not save username');
		}
		savingUsername = false;
	}

	async function cancelSubscription() {
		if (!confirmCancel) { confirmCancel = true; return; }
		cancelling = true;
		try {
			await encore.billing.CancelMySubscription();
			subscription = null;
			confirmCancel = false;
			await loadProfile();
		} catch (err) {
			alert((err as Error).message || 'Could not cancel subscription');
		}
		cancelling = false;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-sm max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} role="presentation" onkeydown={(e) => e.stopPropagation()}>

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Profile</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-6 flex flex-col items-center space-y-4">
				<label class="relative cursor-pointer group">
					<Avatar class="size-20 border-2 border-border hover:border-primary/50 transition-colors">
						<AvatarImage src={avatarSrc} alt="Avatar" />
						<AvatarFallback class="text-2xl font-semibold text-purple-600 dark:text-purple-300 bg-purple-100 dark:bg-purple-900">{user?.email?.charAt(0)?.toUpperCase() || '?'}</AvatarFallback>
					</Avatar>
					<div class="absolute inset-0 rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
						{#if uploadingAvatar}
							<LoaderCircle class="size-5 text-white animate-spin" />
						{:else}
							<Camera class="size-5 text-white" />
						{/if}
					</div>
					<input type="file" accept="image/png,image/jpeg,image/webp,image/gif" class="hidden" onchange={handleAvatarChange} />
				</label>

				<div class="text-center">
					<p class="text-sm font-semibold">{user?.email || '—'}</p>
					<p class="text-xs text-muted-foreground">Member since {formatDate(user?.created_at || '', 'long')}</p>
				</div>

				<div class="flex gap-2">
					{#if user?.role !== 'admin'}
						<Badge class="capitalize bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 border-transparent">Tier: {user?.tier || 'free'}</Badge>
					{/if}
					<Badge class="bg-primary/10 text-primary border-transparent">Role: {user?.role || 'user'}</Badge>
				</div>

				<div class="w-full rounded-xl border border-border p-3 space-y-2">
					<p class="text-xs font-medium">Username</p>
					<div class="relative">
						<input
							type="text"
							bind:value={username}
							placeholder="your_handle"
							class="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm pr-16"
							autocomplete="off"
						/>
						{#if availability === 'checking'}
							<LoaderCircle class="size-4 animate-spin absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
						{:else if availability === 'available'}
							<CheckCircle class="size-4 absolute right-2 top-1/2 -translate-y-1/2 text-green-500" />
						{/if}
					</div>
					{#if availability === 'invalid'}
						<p class="text-xs text-destructive">3-20 characters: lowercase letters, numbers, single - or _ in between</p>
					{:else if availability === 'taken'}
						<p class="text-xs text-destructive">This username is already taken</p>
					{/if}
					{#if usernameChanged && username.trim().toLowerCase() !== originalUsername && availability !== 'invalid' && availability !== 'taken' && availability !== 'checking'}
						<Button size="sm" class="w-full" onclick={saveUsername} disabled={!canSaveUsername}>
							{savingUsername ? 'Saving…' : usernameSaved ? 'Saved!' : 'Save username'}
						</Button>
					{/if}
				</div>

				<div class="w-full rounded-xl border border-border p-3 space-y-2">
					<p class="text-xs font-medium">Subscription</p>
					<p class="text-xs text-muted-foreground">{user?.tier === 'free' ? 'Free plan · No active subscription' : `Active ${user?.tier} plan`}</p>
					{#if plansReady}
						<Button size="sm" class="w-full mt-1" onclick={() => modal.open('checkout')}>
							{user?.tier === 'free' ? 'Upgrade Plan' : 'Manage Subscription'}
						</Button>
					{:else}
						<p class="text-xs text-muted-foreground italic">Subscriptions being configured — check back soon.</p>
					{/if}

					{#if !subscriptionLoading && subscription}
						{#if confirmCancel}
							<div class="space-y-2">
								<p class="text-xs text-muted-foreground">Cancel your active subscription? You'll keep access until the end of the billing period.</p>
								<div class="flex gap-2">
									<Button size="sm" variant="outline" class="flex-1" onclick={() => confirmCancel = false} disabled={cancelling}>Keep</Button>
									<Button size="sm" variant="destructive" class="flex-1" onclick={cancelSubscription} disabled={cancelling}>
										{cancelling ? 'Cancelling…' : 'Confirm cancel'}
									</Button>
								</div>
							</div>
						{:else}
							<Button size="sm" variant="destructive" class="w-full mt-1" onclick={cancelSubscription}>
								Cancel Subscription
							</Button>
						{/if}
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}
