<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let tickets = $state(data.tickets);
	let activeId = $state<string | null>(null);
	let thread = $state<any[]>([]);
	let activeSubject = $state('');
	let reply = $state('');

	function filterStatus(s: string) {
		const url = new URL(page.url);
		if (s) url.searchParams.set('status', s);
		else url.searchParams.delete('status');
		goto(url.pathname + url.search, { keepFocus: true });
	}

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

	async function sendReply() {
		if (!reply.trim() || !activeId) return;
		try {
			const m = await encore.social.ReplyTicket(activeId, { body: reply, turnstile_token: '' });
			thread = [...thread, m];
			reply = '';
		} catch (e) {
			console.error('[support] reply failed:', e);
		}
	}

	async function resolve() {
		if (!activeId) return;
		await encore.social.ResolveTicket(activeId);
		const list = await encore.social.AdminListTickets({ Status: data.status });
		tickets = list.tickets || [];
		activeId = null;
		thread = [];
	}
</script>

<svelte:head><title>Support - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Support Tickets</h1>

	<div class="flex items-center gap-2 mb-4">
		<Button variant={data.status === '' ? 'default' : 'outline'} size="sm" onclick={() => filterStatus('')}>All</Button>
		<Button variant={data.status === 'open' ? 'default' : 'outline'} size="sm" onclick={() => filterStatus('open')}>Open</Button>
		<Button variant={data.status === 'resolved' ? 'default' : 'outline'} size="sm" onclick={() => filterStatus('resolved')}>Resolved</Button>
	</div>

	<div class="grid md:grid-cols-[280px_1fr] gap-4 items-start">
		<div class="space-y-2">
			{#each tickets as t}
				<button onclick={() => openTicket(t.id, t.subject)} class="w-full text-left p-3 rounded-lg border {activeId === t.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted'} transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium truncate">{t.subject}</span>
						<span class="text-[10px] uppercase bg-muted rounded-full px-1.5 py-0.5">{t.status}</span>
					</div>
					<p class="text-xs text-muted-foreground mt-0.5">User {t.user_id.slice(0, 8)}… · {t.priority}</p>
				</button>
			{/each}
		</div>

		<div class="rounded-lg border border-border flex flex-col h-[60vh]">
			{#if activeId}
				<div class="p-3 border-b border-border flex items-center justify-between">
					<span class="text-sm font-medium">{activeSubject}</span>
					<Button size="sm" variant="outline" onclick={resolve}>Resolve</Button>
				</div>
				<div class="flex-1 overflow-y-auto p-4 space-y-3">
					{#each thread as m}
						<div class="flex {m.is_staff ? 'justify-end' : 'justify-start'}">
							<div class="max-w-[75%] rounded-lg px-3 py-2 text-sm {m.is_staff ? 'bg-primary text-primary-foreground' : 'bg-muted'}">
								{m.body}
							</div>
						</div>
					{/each}
				</div>
				<div class="border-t border-border p-3 flex gap-2">
					<Input bind:value={reply} placeholder="Reply as staff…" />
					<Button onclick={sendReply} disabled={!reply.trim()}>Reply</Button>
				</div>
			{:else}
				<div class="flex-1 flex items-center justify-center text-muted-foreground text-sm">
					Select a ticket
				</div>
			{/if}
		</div>
	</div>
</section>
