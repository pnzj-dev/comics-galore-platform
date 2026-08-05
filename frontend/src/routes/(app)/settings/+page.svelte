<script lang="ts">
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { currentUser } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';

	let prefs = $state({
		email_new_from_following: true,
		email_support_replies: true,
		email_marketing: false,
		in_app_enabled: true
	});
	let loading = $state(true);
	let saved = $state(false);

	onMount(async () => {
		if (!$currentUser) { await goto('/login'); return; }
		try {
			const res = await api.get<typeof prefs>('/me/notification-preferences');
			prefs = res;
		} catch {}
		loading = false;
	});

	async function save() {
		await api.patch('/me/notification-preferences', prefs);
		saved = true;
		setTimeout(() => saved = false, 2000);
	}
</script>

<svelte:head><title>Settings — Comics Galore</title></svelte:head>

<section class="py-8 max-w-xl mx-auto">
	<h1 class="text-3xl font-bold mb-6">Notification Preferences</h1>

	{#if loading}
		<div class="animate-pulse space-y-4">
			<div class="h-12 bg-gray-200 dark:bg-gray-700 rounded"></div>
			<div class="h-12 bg-gray-200 dark:bg-gray-700 rounded"></div>
		</div>
	{:else}
		<Card>
			<CardHeader><CardTitle>Email Notifications</CardTitle></CardHeader>
			<CardContent class="space-y-4">
				<label class="flex items-center gap-3 cursor-pointer">
					<input type="checkbox" bind:checked={prefs.email_new_from_following} class="rounded" />
					<span class="text-sm">New comics from creators you follow</span>
				</label>
				<label class="flex items-center gap-3 cursor-pointer">
					<input type="checkbox" bind:checked={prefs.email_support_replies} class="rounded" />
					<span class="text-sm">Support ticket replies</span>
				</label>
				<label class="flex items-center gap-3 cursor-pointer">
					<input type="checkbox" bind:checked={prefs.email_marketing} class="rounded" />
					<span class="text-sm">Marketing emails and promotions</span>
				</label>
			</CardContent>
		</Card>

		<Card class="mt-4">
			<CardHeader><CardTitle>In-App</CardTitle></CardHeader>
			<CardContent>
				<label class="flex items-center gap-3 cursor-pointer">
					<input type="checkbox" bind:checked={prefs.in_app_enabled} class="rounded" />
					<span class="text-sm">Enable in-app notifications</span>
				</label>
			</CardContent>
		</Card>

		<div class="flex items-center gap-3 mt-6">
			<Button onclick={save}>Save Preferences</Button>
			{#if saved}<span class="text-sm text-green-500">Saved!</span>{/if}
		</div>
	{/if}
</section>
