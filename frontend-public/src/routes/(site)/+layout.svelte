<script lang="ts">
	import { modal } from '$lib/stores/modal.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import ThemeToggle from '$lib/components/common/ThemeToggle.svelte';
	import LocaleSwitcher from '$lib/components/common/LocaleSwitcher.svelte';
	import AvatarMenu from '$lib/components/common/AvatarMenu.svelte';
	import AnnouncementBanner from '$lib/components/common/AnnouncementBanner.svelte';
	import { currentUser, isAuthenticated, hydrated } from '$lib/stores/auth';
	import { t } from '$lib/i18n';

	let { data, children } = $props();

	const user = $derived($hydrated ? $currentUser : (data.user || $currentUser));
	const authed = $derived($hydrated ? $isAuthenticated : !!(data.user || $isAuthenticated));
</script>

<nav class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
	<div class="flex h-14 items-center justify-between px-4 max-w-7xl mx-auto">
		<div class="flex items-center gap-4">
			<a href="/" class="font-semibold text-lg">Comics Galore</a>
			<a href="/comics" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.browse')}</a>
			<a href="/series" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.series')}</a>
			<a href="/pricing" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.pricing')}</a>
			{#if authed && user}
				<a href="/messages" class="text-sm text-muted-foreground hover:text-foreground">Messages</a>
				<a href="/lists" class="text-sm text-muted-foreground hover:text-foreground">Lists</a>
				<a href="/favorites" class="text-sm text-muted-foreground hover:text-foreground">Favorites</a>
				{#if user.role === 'uploader' || user.role === 'admin'}
					<a href="/upload" class="text-sm text-muted-foreground hover:text-foreground">{t('nav.upload')}</a>
				{/if}
			{/if}
		</div>
		<div class="flex items-center gap-1">
			<ThemeToggle />
			<LocaleSwitcher compact />
			{#if authed && user}
				<AvatarMenu />
			{:else}
				<Button variant="ghost" size="sm" onclick={() => modal.open('login')}>{t('nav.login')}</Button>
				<Button size="sm" onclick={() => modal.open('register')}>{t('nav.register')}</Button>
			{/if}
		</div>
	</div>
</nav>

<AnnouncementBanner />

<main class="max-w-7xl mx-auto p-4 flex-1 w-full">
	{@render children()}
</main>

<footer class="border-t py-4 text-center text-xs text-muted-foreground">
	<a href="/legal/terms" class="hover:underline mx-2">{t('footer.terms')}</a>
	<a href="/legal/privacy" class="hover:underline mx-2">{t('footer.privacy')}</a>
	<a href="/legal/dmca" class="hover:underline mx-2">{t('footer.dmca')}</a>
	<p class="mt-2">© {new Date().getFullYear()} Comics Galore. All rights reserved.</p>
</footer>
