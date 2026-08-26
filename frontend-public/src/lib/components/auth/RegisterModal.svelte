<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { modal } from '$lib/stores/modal.svelte';
	import RegisterForm from './RegisterForm.svelte';

	const open = $derived(modal.isOpen('register'));

	function close() {
		modal.close('register');
	}

	function switchToLogin() {
		modal.close('register');
		modal.open('login');
	}

	async function handleSuccess() {
		modal.close('register');
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
			<RegisterForm onSwitchToLogin={switchToLogin} onSuccess={handleSuccess} />
		</div>
	</div>
{/if}
