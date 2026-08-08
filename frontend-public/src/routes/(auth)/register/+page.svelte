<script lang="ts">
	import { goto } from '$app/navigation';
	import { register } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let email = $state('');
	let password = $state('');
	let confirm = $state('');
	let agreedToTerms = $state(false);
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (password !== confirm) {
			error = 'Passwords do not match';
			return;
		}

		loading = true;

		try {
			await register(email, password);
			await goto('/');
		} catch (err) {
			error = (err as Error).message;
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Register - Comics Galore</title>
</svelte:head>

<div class="flex min-h-[80vh] items-center justify-center p-4">
	<Card class="w-full max-w-md">
		<CardHeader>
			<CardTitle>Register</CardTitle>
			<CardDescription>Create your Comics Galore account</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input id="email" type="email" bind:value={email} required placeholder="you@example.com" />
				</div>
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input id="password" type="password" bind:value={password} required placeholder="Min 8 characters" />
					{#if password.length > 0 && password.length < 8}
						<p class="text-xs text-destructive">Password must be at least 8 characters</p>
					{/if}
				</div>
				<div class="space-y-2">
					<Label for="confirm">Confirm Password</Label>
					<Input id="confirm" type="password" bind:value={confirm} required placeholder="Repeat password" />
				</div>

				<label class="flex items-center gap-2 text-sm cursor-pointer">
					<input type="checkbox" bind:checked={agreedToTerms} class="rounded" />
					I agree to the <a href="/legal/terms" class="text-primary hover:underline">Terms of Service</a> and <a href="/legal/privacy" class="text-primary hover:underline">Privacy Policy</a>
				</label>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<Button type="submit" class="w-full" disabled={loading || !agreedToTerms}>
					{loading ? 'Creating account...' : 'Create account'}
				</Button>
			</form>
		</CardContent>
		<CardFooter class="flex justify-center">
			<p class="text-sm text-muted-foreground">
				Already have an account?
				<a href="/login" class="text-primary underline underline-offset-4 hover:no-underline">Login</a>
			</p>
		</CardFooter>
	</Card>
</div>
