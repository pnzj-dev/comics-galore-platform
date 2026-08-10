<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';

	let { data } = $props();

	let prefs = $derived(data.prefs);

	async function save() {
		await encore.auth.UpdateNotificationPrefs(prefs);
	}
</script>

<svelte:head><title>Settings — Comics Galore</title></svelte:head>

<section class="py-8 max-w-xl mx-auto">
	<h1 class="text-3xl font-bold mb-6">Notification Preferences</h1>

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

	<Button onclick={save}>Save Preferences</Button>

</section>
