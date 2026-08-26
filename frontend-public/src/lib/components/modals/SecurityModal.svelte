<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { modal } from '$lib/stores/modal.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import {
		webauthnSupported,
		prepareCreationOptions,
		convertCreation,
		mapWebauthnError,
	} from '$lib/webauthn';
	import type { auth } from '$lib/api/encore-client';

	let { onClose }: { onClose?: () => void } = $props();

	let passkeys = $state<auth.PasskeyInfo[]>([]);
	let accounts = $state<auth.AccountInfo[]>([]);
	let sessions = $state<auth.SessionInfo[]>([]);
	let siteName = $state('Comics Galore');
	let error = $state('');
	let adding = $state(false);

	// TOTP (authenticator-app) two-factor authentication.
	let totpEnabled = $state(false);
	let totpMode = $state<'idle' | 'setup' | 'disable'>('idle');
	let totpSetup = $state<{ secret: string; otpauth_url: string; qr_image: string } | null>(null);
	let totpCode = $state('');
	let totpBusy = $state(false);

	const open = $derived(modal.isOpen('security'));
	const supported = webauthnSupported();

	$effect(() => {
		if (open) load();
	});

	async function load() {
		try {
			const [p, a, s, site, totp] = await Promise.all([
				encore.auth.ListPasskeys(),
				encore.auth.ListAccounts(),
				encore.auth.ListSessions(),
				encore.auth.GetSiteConfig(),
				encore.auth.TOTPStatus(),
			]);
			passkeys = p.passkeys || [];
			accounts = a.accounts || [];
			sessions = s.sessions || [];
			siteName = site.site_name || 'Comics Galore';
			totpEnabled = totp.enabled;
		} catch {
			/* ignore */
		}
	}

	function close() {
		modal.close('security');
		onClose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	function defaultPasskeyName(): string {
		const host = window.location.hostname;
		return `${siteName}${host ? ` (${host})` : ''}`;
	}

	async function startTotpSetup() {
		error = '';
		totpBusy = true;
		try {
			totpSetup = await encore.auth.SetupTOTP();
			totpCode = '';
			totpMode = 'setup';
		} catch (err) {
			error = (err as Error).message;
		} finally {
			totpBusy = false;
		}
	}

	async function confirmTotp() {
		if (!totpSetup) return;
		error = '';
		totpBusy = true;
		try {
			await encore.auth.ConfirmTOTP({ secret: totpSetup.secret, code: totpCode });
			totpEnabled = true;
			totpSetup = null;
			totpCode = '';
			totpMode = 'idle';
		} catch (err) {
			error = (err as Error).message;
		} finally {
			totpBusy = false;
		}
	}

	async function disableTotp() {
		error = '';
		totpBusy = true;
		try {
			await encore.auth.DisableTOTP({ code: totpCode });
			totpEnabled = false;
			totpCode = '';
			totpMode = 'idle';
		} catch (err) {
			error = (err as Error).message;
		} finally {
			totpBusy = false;
		}
	}

	function cancelTotp() {
		totpSetup = null;
		totpCode = '';
		totpMode = 'idle';
		error = '';
	}

	async function addPasskey() {
		error = '';
		adding = true;
		try {
			const suggested = defaultPasskeyName();
			const name = (prompt('Name this passkey (e.g. MacBook Pro):', suggested) || suggested).trim() || 'Passkey';
			const optsRes = await encore.auth.PasskeyRegisterOptions({ name });
			const options = prepareCreationOptions(optsRes.options);
			const credential = (await navigator.credentials.create(options)) as PublicKeyCredential | null;
			if (!credential) {
				error = 'Passkey registration was cancelled.';
				return;
			}
			const payload = convertCreation(credential);
			const res = await encore.auth.PasskeyRegisterVerify({ name, response: payload });
			passkeys = res.passkeys || [];
		} catch (err) {
			error = mapWebauthnError(err);
		} finally {
			adding = false;
		}
	}

	async function removePasskey(id: string) {
		error = '';
		try {
			await encore.auth.DeletePasskey(id);
		} catch (err) {
			error = (err as Error).message;
			return;
		}
		await load();
	}

	async function unlinkAccount(id: string) {
		error = '';
		try {
			await encore.auth.UnlinkAccount(id);
		} catch (err) {
			error = (err as Error).message;
			return;
		}
		await load();
	}

	async function revokeSession(id: string) {
		error = '';
		try {
			await encore.auth.RevokeSession({ session_id: id });
		} catch (err) {
			error = (err as Error).message;
			return;
		}
		await load();
	}

	async function logoutAll() {
		error = '';
		try {
			await encore.auth.LogoutAll();
		} catch (err) {
			error = (err as Error).message;
			return;
		}
		await load();
	}

	function linkOAuth(provider: string) {
		const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:4000';
		window.location.href = `${BACKEND_URL}/auth/oauth/${provider}`;
	}

	function formatDate(d: string): string {
		if (!d) return '—';
		return new Date(d).toLocaleString();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-xl max-h-[90vh] overflow-y-auto" role="presentation">

			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Security</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>

			<div class="p-4 space-y-4">
				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<Card>
					<CardHeader>
						<CardTitle>Passkeys</CardTitle>
						<CardDescription>Sign in with your device's fingerprint, face, or PIN.</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3">
						{#if !supported}
							<p class="text-sm text-muted-foreground">Your browser does not support passkeys.</p>
						{:else}
							{#each passkeys as pk (pk.id)}
								<div class="flex items-center justify-between py-1.5 border-b border-border/50 last:border-0">
									<div>
										<p class="text-sm font-medium">{pk.name}</p>
										<p class="text-xs text-muted-foreground">
											Added {formatDate(pk.created_at)}{pk.last_used_at ? ` · Last used ${formatDate(pk.last_used_at)}` : ''}
										</p>
									</div>
									<Button variant="outline" size="sm" onclick={() => removePasskey(pk.id)}>Remove</Button>
								</div>
							{:else}
								<p class="text-sm text-muted-foreground">No passkeys registered yet.</p>
							{/each}
							<Button class="w-full" onclick={addPasskey} disabled={adding}>
								{adding ? 'Adding…' : 'Add a passkey'}
							</Button>
						{/if}
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Two-factor authentication</CardTitle>
						<CardDescription>Add a second step to sign in with an authenticator app (e.g. Google Authenticator, Authy).</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3">
						{#if totpMode === 'setup' && totpSetup}
							<div class="flex flex-col items-center gap-3">
								<img src={totpSetup.qr_image} alt="QR code for two-factor authentication" class="size-44 rounded-lg border border-border" />
								<p class="text-xs text-muted-foreground text-center">
									Scan with your authenticator app, or enter this key manually:
									<code class="block mt-1 select-all break-all text-foreground">{totpSetup.secret}</code>
								</p>
								<input
									type="text"
									bind:value={totpCode}
									inputmode="numeric"
									maxlength="6"
									placeholder="6-digit code"
									class="w-full rounded-md border border-input bg-background px-3 py-2 text-center text-lg tracking-widest"
								/>
								<div class="flex gap-2 w-full">
									<Button variant="outline" class="flex-1" onclick={cancelTotp}>Cancel</Button>
									<Button class="flex-1" onclick={confirmTotp} disabled={totpBusy || totpCode.length < 6}>Verify & enable</Button>
								</div>
							</div>
						{:else if totpMode === 'disable'}
							<div class="space-y-3">
								<p class="text-xs text-muted-foreground">Enter a code from your authenticator app to disable two-factor authentication.</p>
								<input
									type="text"
									bind:value={totpCode}
									inputmode="numeric"
									maxlength="6"
									placeholder="6-digit code"
									class="w-full rounded-md border border-input bg-background px-3 py-2 text-center text-lg tracking-widest"
								/>
								<div class="flex gap-2 w-full">
									<Button variant="outline" class="flex-1" onclick={cancelTotp}>Cancel</Button>
									<Button variant="destructive" class="flex-1" onclick={disableTotp} disabled={totpBusy || totpCode.length < 6}>Disable</Button>
								</div>
							</div>
						{:else if totpEnabled}
							<div class="flex items-center justify-between">
								<div>
									<p class="text-sm font-medium text-green-600 dark:text-green-400">Enabled</p>
									<p class="text-xs text-muted-foreground">You'll need a code from your authenticator app to sign in.</p>
								</div>
								<Button variant="outline" size="sm" onclick={() => { totpCode = ''; totpMode = 'disable'; }}>Disable</Button>
							</div>
						{:else}
							<div class="flex items-center justify-between">
								<p class="text-sm text-muted-foreground">Not enabled.</p>
								<Button size="sm" onclick={startTotpSetup} disabled={totpBusy}>Enable</Button>
							</div>
						{/if}
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Linked accounts</CardTitle>
						<CardDescription>Connect social sign-in providers to your account.</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3">
						{#each accounts as acc (acc.id)}
							<div class="flex items-center justify-between py-1.5 border-b border-border/50 last:border-0">
								<div>
									<p class="text-sm font-medium capitalize">{acc.provider}</p>
									{#if acc.email}<p class="text-xs text-muted-foreground">{acc.email}</p>{/if}
								</div>
								<Button variant="outline" size="sm" onclick={() => unlinkAccount(acc.id)}>Unlink</Button>
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No linked social accounts.</p>
						{/each}

						<div class="flex flex-wrap gap-2 pt-1">
							<Button variant="outline" size="sm" onclick={() => linkOAuth('google')}>Connect Google</Button>
							<Button variant="outline" size="sm" onclick={() => linkOAuth('facebook')}>Connect Facebook</Button>
							<Button variant="outline" size="sm" onclick={() => linkOAuth('twitter')}>Connect X</Button>
							<Button variant="outline" size="sm" onclick={() => linkOAuth('apple')}>Connect Apple</Button>
						</div>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Active sessions</CardTitle>
						<CardDescription>Devices currently signed in to your account.</CardDescription>
					</CardHeader>
					<CardContent class="space-y-3">
						{#each sessions as s (s.id)}
							<div class="flex items-center justify-between py-1.5 border-b border-border/50 last:border-0">
								<p class="text-sm text-muted-foreground">Started {formatDate(s.created_at)}</p>
								<Button variant="outline" size="sm" onclick={() => revokeSession(s.id)}>Revoke</Button>
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No active sessions.</p>
						{/each}
						{#if sessions.length > 0}
							<Button class="w-full" variant="destructive" onclick={logoutAll}>Sign out of all devices</Button>
						{/if}
					</CardContent>
				</Card>
			</div>
		</div>
	</div>
{/if}
