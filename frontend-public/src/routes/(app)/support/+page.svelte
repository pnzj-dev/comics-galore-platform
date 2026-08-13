<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let tickets = $state(data.tickets);
	let activeId = $state<string | null>(null);
	let thread = $state<any[]>([]);
	let subject = $state('');
	let activeSubject = $state('');
	let body = $state('');
	let reply = $state('');
	let creating = $state(false);
	let error = $state('');

	async function openTicket(id: string, subj: string) {
		activeId = id;
		activeSubject = subj;
		try {
			const res = await encore.social.GetTicket(id);
			thread = res.messages || [];
		} catch {
			thread = [];
		}
	}

	async function createTicket() {
		if (!subject.trim() || !body.trim()) {
			error = 'Subject and message are required';
			return;
		}
		creating = true;
		error = '';
		try {
			const t = await encore.social.CreateTicket({ subject, body, priority: 'normal' });
			subject = '';
			body = '';
			const list = await encore.social.ListMyTickets();
			tickets = list.tickets || [];
			openTicket(t.ticket.id, t.ticket.subject);
		} catch (e) {
			error = (e as Error).message;
		} finally {
			creating = false;
		}
	}

	async function sendReply() {
		if (!reply.trim() || !activeId) return;
		try {
			const m = await encore.social.ReplyTicket(activeId, { body: reply });
			thread = [...thread, m];
			reply = '';
		} catch (e) {
			console.error('[support] reply failed:', e);
		}
	}
</script>

<svelte:head><title>Support - Comics Galore</title></svelte:head>

<section class="py-8">
	<h1 class="text-3xl font-bold mb-6">Support</h1>

	<div class="grid md:grid-cols-[280px_1fr] gap-4 items-start">
		<div class="space-y-4">
			<Card>
				<CardHeader><CardTitle>New ticket</CardTitle></CardHeader>
				<CardContent class="space-y-3">
					<div class="space-y-1.5">
						<Label for="subject">Subject</Label>
						<Input id="subject" bind:value={subject} placeholder="Brief summary" />
					</div>
					<div class="space-y-1.5">
						<Label for="body">Message</Label>
						<textarea id="body" bind:value={body} rows="3" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Describe your issue…"></textarea>
					</div>
					{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
					<Button class="w-full" onclick={createTicket} disabled={creating}>{creating ? 'Submitting…' : 'Submit ticket'}</Button>
				</CardContent>
			</Card>

			<div class="space-y-2">
				{#each tickets as t}
					<button onclick={() => openTicket(t.id, t.subject)} class="w-full text-left p-3 rounded-lg border {activeId === t.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted'} transition-colors">
						<div class="flex items-center justify-between">
							<span class="text-sm font-medium truncate">{t.subject}</span>
							<span class="text-[10px] uppercase bg-muted rounded-full px-1.5 py-0.5">{t.status}</span>
						</div>
					</button>
				{/each}
			</div>
		</div>

		<div class="rounded-lg border border-border flex flex-col h-[60vh]">
			{#if activeId}
				<div class="p-3 border-b border-border text-sm font-medium">{activeSubject}</div>
				<div class="flex-1 overflow-y-auto p-4 space-y-3">
					{#each thread as m}
						<div class="flex {m.is_staff ? 'justify-start' : 'justify-end'}">
							<div class="max-w-[75%] rounded-lg px-3 py-2 text-sm {m.is_staff ? 'bg-primary text-primary-foreground' : 'bg-muted'}">
								{m.body}
							</div>
						</div>
					{/each}
				</div>
				<div class="border-t border-border p-3 flex gap-2">
					<Input bind:value={reply} placeholder="Reply…" />
					<Button onclick={sendReply} disabled={!reply.trim()}>Reply</Button>
				</div>
			{:else}
				<div class="flex-1 flex items-center justify-center text-muted-foreground text-sm">
					Select a ticket or create a new one
				</div>
			{/if}
		</div>
	</div>
</section>
