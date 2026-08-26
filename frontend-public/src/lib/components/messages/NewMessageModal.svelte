<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { goto } from '$app/navigation';
	import { modal } from '$lib/stores/modal.svelte';
	import { newMessageTarget, clearNewMessage } from '$lib/stores/new-message.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	const open = $derived(modal.isOpen('new-message'));
	const target = $derived(newMessageTarget);

	let body = $state('');
	let sending = $state(false);
	let error = $state('');

	function close() {
		modal.close('new-message');
		clearNewMessage();
		body = '';
		error = '';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	async function send() {
		const userId = target.userId;
		if (!userId || !body.trim()) return;
		sending = true;
		error = '';
		try {
			await encore.social.StartConversation(userId, { body: body.trim() });
			close();
			await goto('/messages');
		} catch (e) {
			error = (e as Error).message || 'Could not send message';
		} finally {
			sending = false;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
	<div class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4" onclick={close} onkeydown={handleKeydown} role="dialog" tabindex="-1">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-md" onclick={(e) => e.stopPropagation()} role="presentation">
			<div class="flex items-center justify-between p-4 border-b">
				<h2 class="text-lg font-semibold">Message {target.name}</h2>
				<button onclick={close} class="hover:bg-muted rounded-lg p-1" aria-label="Close">✕</button>
			</div>
			<div class="p-4 space-y-3">
				<Textarea bind:value={body} placeholder={`Say hi to ${target.name}…`} rows={4} />
				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
				<Button class="w-full" onclick={send} disabled={sending || !body.trim()}>
					{sending ? 'Sending…' : 'Send'}
				</Button>
			</div>
		</div>
	</div>
{/if}
