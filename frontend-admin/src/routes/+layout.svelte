<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { currentUser, isAuthenticated, fetchMe, logout } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Settings } from 'lucide-svelte';

	let { children } = $props();
	let profileOpen = $state(false);

	onMount(() => { fetchMe(); });

	const user = $derived($currentUser);
	const authed = $derived($isAuthenticated);
</script>

{#if authed && user?.role === 'admin'}
	<nav class="sticky top-0 z-50 border-b bg-background/95 backdrop-blur">
		<div class="flex h-14 items-center justify-between px-4 max-w-7xl mx-auto">
			<div class="flex items-center gap-4">
				<a href="/dashboard" class="font-semibold text-sm">Admin</a>
				<a href="/dashboard" class="text-sm text-muted-foreground hover:text-foreground">Dashboard</a>
				<a href="/moderation" class="text-sm text-muted-foreground hover:text-foreground">Moderation</a>
				<a href="/users" class="text-sm text-muted-foreground hover:text-foreground">Users</a>
				<a href="/subscriptions" class="text-sm text-muted-foreground hover:text-foreground">Subscriptions</a>
				<a href="/comics" class="text-sm text-muted-foreground hover:text-foreground">Comics</a>
			</div>
			<div class="flex items-center gap-1">
				<a href="/settings" class="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" aria-label="Settings">
					<Settings class="size-4" />
				</a>
				<button onclick={() => profileOpen = true} class="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-muted transition-colors">
					<span class="text-xs text-muted-foreground">{user.email}</span>
				</button>
				<Button variant="ghost" size="sm" onclick={() => { logout(); }}>Logout</Button>
			</div>
		</div>
	</nav>

	<div class="bg-red-600 text-white text-xs text-center py-1.5 font-medium">
		Action required: Configure the complete plan matrix in admin settings.
	</div>

	<main class="max-w-7xl mx-auto p-4 flex-1 w-full">
		{@render children()}
	</main>
{:else}
	{@render children()}
{/if}
