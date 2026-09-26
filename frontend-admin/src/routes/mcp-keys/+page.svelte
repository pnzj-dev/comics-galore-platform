<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { formatDate } from '$lib/utils/format';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let keys = $state(data.keys);
	let showCreate = $state(false);
	let newLabel = $state('');
	let newUserId = $state('');
	let createdKey = $state<string | null>(null);
	let copied = $state(false);
	let error = $state('');

	async function refresh() {
		const res = await encore.mcp.AdminListMcpKeys();
		keys = res.keys || [];
	}

	async function createKey() {
		error = '';
		try {
			const res = await encore.mcp.AdminCreateMcpKey({ label: newLabel, user_id: newUserId });
			createdKey = res.key;
			copied = false;
			newLabel = '';
			newUserId = '';
			showCreate = false;
			await refresh();
		} catch (e) {
			error = (e as Error).message || 'Failed to create key';
		}
	}

	async function revokeKey(id: string) {
		await encore.mcp.AdminRevokeMcpKey(id);
		await refresh();
	}

	async function copyKey() {
		if (!createdKey) return;
		try {
			await navigator.clipboard.writeText(createdKey);
			copied = true;
		} catch {
			/* clipboard unavailable */
		}
	}

	function dismissKey() {
		createdKey = null;
		copied = false;
	}
</script>

<svelte:head><title>MCP Keys — Admin</title></svelte:head>

<section>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-bold">MCP Keys</h1>
		<Button size="sm" onclick={() => (showCreate = true)}>Create key</Button>
	</div>

	{#if showCreate}
		<div class="mb-6 rounded-xl border bg-card p-5 space-y-4">
			<h2 class="font-semibold">Create MCP key</h2>
			<div class="space-y-2">
				<Label for="label">Label</Label>
				<Input id="label" bind:value={newLabel} placeholder="e.g. Claude Desktop" />
			</div>
			<div class="space-y-2">
				<Label for="user">Bound user</Label>
				<select id="user" bind:value={newUserId} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
					<option value="">Myself (default)</option>
					{#each data.users as u}
						<option value={u.id}>{u.email}</option>
					{/each}
				</select>
			</div>
			{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
			<div class="flex gap-2">
				<Button size="sm" onclick={createKey}>Create</Button>
				<Button size="sm" variant="outline" onclick={() => (showCreate = false)}>Cancel</Button>
			</div>
		</div>
	{/if}

	{#if createdKey}
		<div class="mb-6 rounded-xl border border-amber-500/40 bg-amber-50 dark:bg-amber-950/30 p-5 space-y-3">
			<h2 class="font-semibold">Key created — copy it now</h2>
			<p class="text-sm text-muted-foreground">This is the only time the full key is shown. Store it somewhere safe.</p>
			<code class="block w-full rounded-md bg-background px-3 py-2 text-sm break-all border">{createdKey}</code>
			<div class="flex gap-2">
				<Button size="sm" onclick={copyKey}>{copied ? 'Copied' : 'Copy'}</Button>
				<Button size="sm" variant="outline" onclick={dismissKey}>I've saved it</Button>
			</div>
		</div>
	{/if}

	{#if keys.length === 0}
		<p class="text-sm text-muted-foreground">No MCP keys yet.</p>
	{:else}
		<div class="overflow-x-auto rounded-lg border">
			<table class="w-full text-sm">
				<thead class="bg-muted/50 text-left">
					<tr>
						<th class="px-4 py-2 font-medium">Label</th>
						<th class="px-4 py-2 font-medium">Key</th>
						<th class="px-4 py-2 font-medium">Bound user</th>
						<th class="px-4 py-2 font-medium">Created</th>
						<th class="px-4 py-2 font-medium">Status</th>
						<th class="px-4 py-2 font-medium text-right">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each keys as key (key.id)}
						<tr class="border-t">
							<td class="px-4 py-2">{key.label || '—'}</td>
							<td class="px-4 py-2 font-mono">…{key.key_suffix}</td>
							<td class="px-4 py-2">{key.username || key.user_id}</td>
							<td class="px-4 py-2">{formatDate(key.created_at)}</td>
							<td class="px-4 py-2">
								{#if key.revoked_at}
									<span class="px-2 py-0.5 rounded-full text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">Revoked</span>
								{:else}
									<span class="px-2 py-0.5 rounded-full text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">Active</span>
								{/if}
							</td>
							<td class="px-4 py-2 text-right">
								{#if !key.revoked_at}
									<Button size="sm" variant="destructive" onclick={() => revokeKey(key.id)}>Revoke</Button>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
