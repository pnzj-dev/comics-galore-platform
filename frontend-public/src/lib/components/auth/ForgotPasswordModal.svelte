<script lang="ts">
	import { modal } from '$lib/stores/modal.svelte';
	import ForgotPasswordForm from './ForgotPasswordForm.svelte';

	const open = $derived(modal.isOpen('forgot-password'));

	function close() {
		modal.close('forgot-password');
	}

	function switchToLogin() {
		modal.close('forgot-password');
		modal.open('login');
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4 overflow-y-auto" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="w-full max-w-md" onclick={(e) => e.stopPropagation()} role="presentation">
			<ForgotPasswordForm onSwitchToLogin={switchToLogin} />
		</div>
	</div>
{/if}
