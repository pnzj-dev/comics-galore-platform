<script lang="ts">
	import { currentUser } from '$lib/stores/auth';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';

	const user = $derived($currentUser);
</script>

<svelte:head>
	<title>Comics Galore</title>
	<meta name="description" content="Discover, read, and share digital comics." />
</svelte:head>

<section class="flex flex-col items-center justify-center py-20 text-center">
	<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">Welcome to Comics Galore</h1>
	<p class="mt-4 text-lg text-muted-foreground max-w-2xl">
		Discover and read amazing comics from creators around the world.
	</p>

	<div class="mt-8 flex gap-4">
		{#if !user}
			<Button size="lg" href="/register">Get Started</Button>
			<Button size="lg" variant="outline" href="/login">Sign In</Button>
		{:else}
			<Button size="lg">Browse Comics</Button>
		{/if}
	</div>
</section>

<section class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 py-8">
	<Card>
		<CardHeader>
			<CardTitle>Upload Comics</CardTitle>
			<CardDescription>Share your creations with the world as an uploader.</CardDescription>
		</CardHeader>
	</Card>
	<Card>
		<CardHeader>
			<CardTitle>Read Anywhere</CardTitle>
			<CardDescription>Web reader with progress tracking and Continue Reading.</CardDescription>
		</CardHeader>
	</Card>
	<Card>
		<CardHeader>
			<CardTitle>Premium Tiers</CardTitle>
			<CardDescription>Support creators with crypto subscriptions via NowPayments.</CardDescription>
		</CardHeader>
	</Card>
</section>
