<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { currentUser } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let conversations = $state(data.conversations);
	let activeId = $state<string | null>(null);
	let messages = $state<any[]>([]);
	let otherUserId = $state('');
	let draft = $state('');
	let loadingThread = $state(false);

	const user = $derived($currentUser);

	function isMine(senderId: string): boolean {
		return senderId === user?.id;
	}

	async function openConversation(id: string, other: string) {
		activeId = id;
		otherUserId = other;
		loadingThread = true;
		try {
			const res = await encore.social.GetConversation(id);
			messages = res.messages || [];
		} catch {
			messages = [];
		} finally {
			loadingThread = false;
		}
	}

	async function send() {
		if (!draft.trim() || !activeId) return;
		try {
			const msg = await encore.social.SendMessage(activeId, { body: draft });
			messages = [...messages, msg];
			draft = '';
			// refresh list ordering/unread
			const list = await encore.social.ListConversations();
			conversations = list.conversations || [];
		} catch (e) {
			console.error('[messages] send failed:', e);
		}
	}
</script>

<svelte:head><title>Messages - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Messages</h1>

	{#if conversations.length === 0}
		<div class="text-center py-16">
			<p class="text-lg text-muted-foreground">No conversations yet.</p>
			<p class="text-sm text-muted-foreground mt-1">Message an uploader from their comic page to start a chat.</p>
		</div>
	{:else}
		<div class="grid md:grid-cols-[280px_1fr] gap-4 items-start">
			<div class="space-y-2">
				{#each conversations as convo}
					<button
						onclick={() => openConversation(convo.id, convo.other_user_id)}
						class="w-full text-left p-3 rounded-lg border {activeId === convo.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted'} transition-colors"
					>
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium truncate">{convo.other_user_id.slice(0, 8)}…</span>
							{#if convo.unread_count > 0}
								<span class="text-[10px] bg-primary text-primary-foreground rounded-full px-1.5 py-0.5">{convo.unread_count}</span>
							{/if}
						</div>
						<p class="text-xs text-muted-foreground truncate mt-0.5">{convo.last_message || 'No messages yet'}</p>
					</button>
				{/each}
			</div>

			<div class="rounded-lg border border-border flex flex-col h-[60vh]">
				{#if activeId}
					<div class="flex-1 overflow-y-auto p-4 space-y-3">
						{#if loadingThread}
							<p class="text-sm text-muted-foreground">Loading…</p>
						{:else if messages.length === 0}
							<p class="text-sm text-muted-foreground text-center py-8">Say hello!</p>
						{:else}
							{#each messages as m}
								<div class="flex {isMine(m.sender_id) ? 'justify-end' : 'justify-start'}">
									<div class="max-w-[75%] rounded-lg px-3 py-2 text-sm {isMine(m.sender_id) ? 'bg-primary text-primary-foreground' : 'bg-muted'}">
										{m.body}
									</div>
								</div>
							{/each}
						{/if}
					</div>
					<div class="border-t border-border p-3 flex gap-2">
						<Input bind:value={draft} placeholder="Type a message…" />
						<Button onclick={send} disabled={!draft.trim()}>Send</Button>
					</div>
				{:else}
					<div class="flex-1 flex items-center justify-center text-muted-foreground text-sm">
						Select a conversation
					</div>
				{/if}
			</div>
		</div>
	{/if}
</section>
