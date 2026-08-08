<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { currentUser, isAuthenticated, fetchMe } from '$lib/stores/auth';
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import LogoutConfirmationModal from '$lib/components/LogoutConfirmationModal.svelte';
	import NowPaymentsLinkWizard from '$lib/components/NowPaymentsLinkWizard.svelte';
	import { Settings, LogOut, AlertTriangle } from 'lucide-svelte';

	let { children } = $props();
	let profileOpen = $state(false);
	let logoutOpen = $state(false);
	let wizardOpen = $state(false);
	let planMatrixComplete = $state(true);
	let wizardBlocked = $state(false);

	onMount(() => { fetchMe(); });

	const user = $derived($currentUser);
	const authed = $derived($isAuthenticated);

	async function checkPlanMatrix() {
		try {
			const res = await api.get<{ complete: boolean }>('/admin/plans/matrix-status');
			planMatrixComplete = res.complete;
			if (!res.complete && !wizardBlocked) {
				wizardOpen = true;
			}
		} catch {
			planMatrixComplete = true;
		}
	}

	function onWizardClose() {
		wizardOpen = false;
		wizardBlocked = !planMatrixComplete;
	}

	$effect(() => {
		if (authed) checkPlanMatrix();
	});
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
				<button onclick={() => logoutOpen = true} class="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" aria-label="Logout">
					<LogOut class="size-4" />
				</button>
			</div>
		</div>
	</nav>

	{#if !planMatrixComplete}
		<button onclick={() => { wizardOpen = true; wizardBlocked = true; }} class="w-full flex items-center justify-center gap-1.5 bg-red-600 hover:bg-red-700 text-white text-xs text-center py-1.5 font-medium cursor-pointer border-0 transition-colors">
			<AlertTriangle class="size-3" />
			Unlinked NowPayments plans — Click here to configure
		</button>
	{/if}

	<main class="max-w-7xl mx-auto p-4 flex-1 w-full">
		{@render children()}
	</main>

	<LogoutConfirmationModal open={logoutOpen} onClose={() => logoutOpen = false} />
	<NowPaymentsLinkWizard open={wizardOpen} onClose={onWizardClose} />
{:else}
	{@render children()}
{/if}
