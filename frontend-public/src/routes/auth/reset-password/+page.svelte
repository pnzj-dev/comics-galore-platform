<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let email = $state('');
	let sent = $state(false);
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault(); error = ''; loading = true;
		try {
			await api.post('/auth/password-reset/request', { email });
			sent = true;
		} catch (err) {
			error = (err as Error).message;
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>Reset Password — Comics Galore</title></svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Reset Password</CardTitle>
			<CardDescription>{sent ? 'Check your email' : 'Enter your email to receive a reset link'}</CardDescription>
		</CardHeader>
		<CardContent>
			{#if sent}
				<p class="text-sm text-muted-foreground text-center py-4">If an account with that email exists, we've sent a password reset link. Check your inbox.</p>
				<Button class="w-full" variant="outline" href="/login">Back to Login</Button>
			{:else}
				<form onsubmit={handleSubmit} class="space-y-4">
					<div class="space-y-2">
						<Label for="email">Email</Label>
						<Input id="email" type="email" bind:value={email} required placeholder="you@example.com" />
					</div>
					{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
					<Button type="submit" class="w-full" disabled={loading}>{loading ? 'Sending...' : 'Send Reset Link'}</Button>
				</form>
			{/if}
		</CardContent>
	</Card>
</div>
