<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { currentUser, isAuthenticated, fetchMe, logout } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let { children } = $props();

	onMount(() => {
		fetchMe();
	});

	const user = $derived($currentUser);
	const authed = $derived($isAuthenticated);
</script>

<nav class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
	<div class="flex h-14 items-center justify-between px-4 max-w-7xl mx-auto">
		<div class="flex items-center gap-4">
			<a href="/" class="font-semibold text-lg">Comics Galore</a>
			<a href="/comics" class="text-sm text-muted-foreground hover:text-foreground">Browse</a>
			<a href="/pricing" class="text-sm text-muted-foreground hover:text-foreground">Pricing</a>
			{#if authed && user}
				{#if user.role === 'uploader' || user.role === 'admin'}
					<a href="/upload" class="text-sm text-muted-foreground hover:text-foreground">Upload</a>
				{/if}
				{#if user.role === 'moderator' || user.role === 'admin'}
					<a href="/moderation" class="text-sm text-muted-foreground hover:text-foreground">Moderation</a>
				{/if}
				{#if user.role === 'admin'}
					<a href="/subscriptions" class="text-sm text-muted-foreground hover:text-foreground">Subscriptions</a>
				{/if}
			{/if}
		</div>
		<div class="flex items-center gap-2">
			<ThemeToggle />
			{#if authed && user}
				<span class="text-sm text-muted-foreground mr-2">{user.email}</span>
				<span class="text-xs rounded-full bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 px-2 py-0.5 capitalize">{user.tier}</span>
				<span class="text-xs rounded-full bg-primary/10 text-primary px-2 py-0.5">{user.role}</span>
				<Button variant="ghost" size="sm" onclick={logout}>Logout</Button>
			{:else}
				<Button variant="ghost" size="sm" href="/login">Login</Button>
				<Button size="sm" href="/register">Register</Button>
			{/if}
		</div>
	</div>
</nav>

{#if authed && user?.role === 'admin'}
	<div class="bg-red-600 text-white text-xs text-center py-1.5 font-medium">
		Action required: Configure the complete plan matrix in admin settings.
	</div>
{/if}

<main class="max-w-7xl mx-auto p-4">
	{@render children()}
</main>
