<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { modal } from '$lib/stores/modal.svelte';
	import LoginForm from './LoginForm.svelte';

	const open = $derived(modal.isOpen('login'));

	function close() {
		modal.close('login');
	}

	function switchToRegister() {
		modal.close('login');
		modal.open('register');
	}

	function switchToForgotPassword() {
		modal.close('login');
		modal.open('forgot-password');
	}

	async function handleSuccess() {
		modal.close('login');
		await invalidateAll();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4 overflow-y-auto" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="w-full max-w-md" onclick={(e) => e.stopPropagation()} role="presentation">
			<LoginForm onSwitchToRegister={switchToRegister} onForgotPassword={switchToForgotPassword} onSuccess={handleSuccess} />
		</div>
	</div>
{/if}
