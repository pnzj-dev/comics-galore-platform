<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { currentUser, isAuthenticated } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import UserProfileModal from '$lib/components/UserProfileModal.svelte';
	import AppSettingsModal from '$lib/components/AppSettingsModal.svelte';
	import LogoutConfirmationModal from '$lib/components/LogoutConfirmationModal.svelte';
	import { initializeLocale, t } from '$lib/i18n';
	import { Settings, LogOut } from 'lucide-svelte';

	let { data, children } = $props();
	let profileOpen = $state(false);
	let settingsOpen = $state(false);
	let logoutOpen = $state(false);

	// Intentionally initial-value only — locale is resolved once server-side.
	// svelte-ignore state_referenced_locally
	initializeLocale(data.locale);

	onMount(() => {
		if (data.user) {
			currentUser.set(data.user);
			isAuthenticated.set(true);
		}
	});

	const user = $derived(data.user || $currentUser);
	const authed = $derived(!!(data.user) || $isAuthenticated);
</script>

<svelte:head>
	<html lang={data.locale}></html>
</svelte:head>

<nav class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
	<div class="flex h-14 items-center justify-between px-4 max-w-7xl mx-auto">
		<div class="flex items-center gap-4">
			<a href="/" class="font-semibold text-lg">Comics Galore</a>
			<a href="/comics" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.browse')}</a>
			<a href="/pricing" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.pricing')}</a>
			{#if authed && user}
				{#if user.role === 'uploader' || user.role === 'admin'}
					<a href="/upload" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.upload')}</a>
				{/if}
			{/if}
		</div>
		<div class="flex items-center gap-1">
			{#if authed && user}
				<button onclick={() => settingsOpen = true} class="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" aria-label={t('nav.settings')}>
					<Settings class="size-4" />
				</button>
			{/if}
			<ThemeToggle />
			{#if authed && user}
				<button onclick={() => profileOpen = true} class="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-muted transition-colors text-left">
					<span class="text-sm text-muted-foreground">{user.email}</span>
				</button>
				{#if user.role !== 'admin'}
					<span class="text-xs rounded-full bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 px-2 py-0.5 capitalize">{user.tier}</span>
				{/if}
				<span class="text-xs rounded-full bg-primary/10 text-primary px-2 py-0.5">{user.role}</span>
				<button onclick={() => logoutOpen = true} class="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" aria-label={t('nav.logout')}>
					<LogOut class="size-4" />
				</button>
			{:else}
				<Button variant="ghost" size="sm" href="/login">{t('nav.login')}</Button>
				<Button size="sm" href="/register">{t('nav.register')}</Button>
			{/if}
		</div>
	</div>
</nav>

<main class="max-w-7xl mx-auto p-4 flex-1 w-full">
	{@render children()}
</main>

<footer class="border-t py-4 text-center text-xs text-muted-foreground">
	<a href="/legal/terms" class="hover:underline mx-2">{t('footer.terms')}</a>
	<a href="/legal/privacy" class="hover:underline mx-2">{t('footer.privacy')}</a>
	<a href="/legal/dmca" class="hover:underline mx-2">{t('footer.dmca')}</a>
</footer>

<UserProfileModal open={profileOpen} onClose={() => profileOpen = false} />
<AppSettingsModal open={settingsOpen} onClose={() => settingsOpen = false} />
<LogoutConfirmationModal open={logoutOpen} onClose={() => logoutOpen = false} />
