<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import EnvironmentBadge from '$lib/components/common/EnvironmentBadge.svelte';

	let { data } = $props();
</script>

<svelte:head><title>Admin Login — Comics Galore</title></svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<div class="flex items-center gap-2">
				<CardTitle>Admin Login</CardTitle>
				<EnvironmentBadge variant="pill" />
			</div>
			<CardDescription>Sign in to the admin panel</CardDescription>
		</CardHeader>
		{#if data.accessDenied}
			<CardContent class="space-y-4">
				<p class="text-sm text-muted-foreground">
					You're signed in as
					<span class="font-medium text-foreground">{data.user?.email ?? 'this account'}</span>, which doesn't
					have access to the admin panel.
				</p>
				<Button href="/logout" class="w-full">Sign out</Button>
			</CardContent>
		{:else}
			<CardContent>
				<form method="POST" action="?/signIn" class="space-y-4">
					<Button type="submit" class="w-full">Sign in</Button>
				</form>
				<form method="POST" action="?/forgotPassword" class="mt-3 text-center">
					<button type="submit" class="text-sm text-muted-foreground hover:text-foreground hover:underline">Forgot password?</button>
				</form>
			</CardContent>
		{/if}
	</Card>
</div>
