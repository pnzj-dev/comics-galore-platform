<script lang="ts">
	import { register } from '$lib/stores/auth';
	import { encore } from '$lib/api/encore';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client as zodClient } from 'sveltekit-superforms/adapters';
	import { registerSchema } from '$lib/schemas/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import OAuthButtons from '$lib/components/auth/OAuthButtons.svelte';
	import AuthCard from './AuthCard.svelte';
	import PasswordInput from './PasswordInput.svelte';
	import PasswordStrengthMeter from './PasswordStrengthMeter.svelte';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { LoaderCircle, Check } from 'lucide-svelte';
	import { CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	interface Props {
		onSwitchToLogin?: () => void;
		onSuccess?: () => void;
	}

	let { onSwitchToLogin, onSuccess }: Props = $props();

	let error = $state('');
	let loading = $state(false);
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;

	type Availability = 'idle' | 'checking' | 'available' | 'taken' | 'invalid';
	let availability = $state<Availability>('idle');

	const { form, errors, enhance } = superForm(
		{ username: '', email: '', password: '', confirm: '', terms: false },
		{
			SPA: true,
			validationMethod: 'submit-only',
			validators: zodClient(registerSchema),
			onUpdate: async ({ form: f }) => {
				if (!f.valid) return;
				error = '';
				loading = true;
				try {
					await register(f.data.username.trim().toLowerCase(), f.data.email, f.data.password, turnstileToken || '');
					onSuccess?.();
				} catch (err) {
					error = (err as Error).message;
				} finally {
					loading = false;
					turnstileToken = null;
					turnstileReset++;
				}
			},
		},
	);

	const canSubmit = $derived(availability === 'available');

	// Live username validation: format check is synchronous, availability is
	// debounced against the backend so the user gets feedback without a round-trip
	// on every keystroke.
	$effect(() => {
		const u = $form.username.trim().toLowerCase();
		if (u.length === 0) {
			availability = 'idle';
			return;
		}
		if (u.length < 3 || u.length > 20 || !/^[a-z0-9](?:[_-]?[a-z0-9])*$/.test(u)) {
			availability = 'invalid';
			return;
		}
		availability = 'checking';
		const timer = setTimeout(async () => {
			try {
				const res = await encore.auth.UsernameAvailable({ Username: u });
				availability = res.available ? 'available' : (res.valid ? 'taken' : 'invalid');
			} catch {
				availability = 'idle';
			}
		}, 300);
		return () => clearTimeout(timer);
	});
</script>

<AuthCard>
	<CardHeader class="text-center">
		<CardTitle>Register</CardTitle>
		<CardDescription>Create your Comics Galore account</CardDescription>
	</CardHeader>
	<CardContent class="space-y-4">
		<form method="POST" use:enhance class="space-y-4">
			<div class="space-y-2">
				<Label for="username">Username</Label>
				<div class="relative">
					<Input id="username" bind:value={$form.username} placeholder="your_handle" autocomplete="username" />
					{#if availability === 'checking'}
						<LoaderCircle class="absolute right-3 top-1/2 -translate-y-1/2 size-4 animate-spin text-muted-foreground" />
					{:else if availability === 'available'}
						<Check class="absolute right-3 top-1/2 -translate-y-1/2 size-4 text-green-600 dark:text-green-400" />
					{/if}
				</div>
				{#if $errors.username}
					<p class="text-xs text-destructive">{$errors.username}</p>
				{:else if availability === 'invalid'}
					<p class="text-xs text-destructive">3–20 characters, lowercase letters, numbers, and single - or _ in between</p>
				{:else if availability === 'taken'}
					<p class="text-xs text-destructive">This username is already taken</p>
				{:else if availability === 'available'}
					<p class="text-xs text-green-600 dark:text-green-400">Available</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="email">Email</Label>
				<Input id="email" type="email" bind:value={$form.email} placeholder="you@example.com" />
				{#if $errors.email}<p class="text-xs text-destructive">{$errors.email}</p>{/if}
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<PasswordInput id="password" bind:value={$form.password} placeholder="Min 8 characters" autocomplete="new-password" />
				<PasswordStrengthMeter password={$form.password} />
				{#if $errors.password}<p class="text-xs text-destructive">{$errors.password}</p>{/if}
			</div>
			<div class="space-y-2">
				<Label for="confirm">Confirm Password</Label>
				<PasswordInput id="confirm" bind:value={$form.confirm} placeholder="Repeat password" autocomplete="new-password" />
				{#if $errors.confirm}<p class="text-xs text-destructive">{$errors.confirm}</p>{/if}
			</div>

			<label class="flex items-center gap-2 text-sm cursor-pointer">
				<Checkbox bind:checked={$form.terms} />
				I agree to the <a href="/legal/terms" class="text-primary hover:underline">Terms of Service</a> and <a href="/legal/privacy" class="text-primary hover:underline">Privacy Policy</a>
			</label>
			{#if $errors.terms}<p class="text-xs text-destructive">{$errors.terms}</p>{/if}

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Turnstile action="register" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />

			<Button type="submit" class="w-full" disabled={loading || !canSubmit || (turnstileRequired && !turnstileToken)}>
				{loading ? 'Creating account...' : 'Create account'}
			</Button>
		</form>

		<div class="flex items-center gap-3 text-xs text-muted-foreground">
			<Separator class="flex-1" />
			or sign up with
			<Separator class="flex-1" />
		</div>

		<OAuthButtons disabled={!$form.terms} onError={(m) => (error = m)} />
	</CardContent>
	<CardFooter class="flex justify-center">
		<p class="text-sm text-muted-foreground">
			Already have an account?
			<button type="button" class="text-primary underline underline-offset-4 hover:no-underline" onclick={() => onSwitchToLogin?.()}>Login</button>
		</p>
	</CardFooter>
</AuthCard>
