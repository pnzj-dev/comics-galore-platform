<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import Turnstile from '$lib/components/common/Turnstile.svelte';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault(); error = ''; loading = true;
		try {
			const user = await login(email, password, turnstileToken || '');
			await goto(user.role === 'moderator' ? '/moderation' : '/dashboard');
		}
		catch (err) { error = (err as Error).message; }
		finally {
			loading = false;
			turnstileToken = null;
			turnstileReset++;
		}
	}
</script>

<svelte:head><title>Admin Login — Comics Galore</title></svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Admin Login</CardTitle>
			<CardDescription>Sign in to the admin panel</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2"><Label for="email">Email</Label><Input id="email" type="email" bind:value={email} required placeholder="admin@example.com" /></div>
				<div class="space-y-2"><Label for="password">Password</Label><Input id="password" type="password" bind:value={password} required /></div>
				<Turnstile action="login" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />
				{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
				<Button type="submit" class="w-full" disabled={loading}>{loading ? 'Signing in...' : 'Sign in'}</Button>
			</form>
		</CardContent>
	</Card>
</div>
