<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { onMount } from 'svelte';
	import { X, Megaphone } from 'lucide-svelte';

	type Announcement = { id: string; title: string; body: string; target: string; tier: string };

	const STORAGE_KEY = 'comics_galore_dismissed_announcements';

	let announcements = $state<Announcement[]>([]);
	let dismissed = $state<string[]>([]);

	onMount(async () => {
		try {
			const raw = localStorage.getItem(STORAGE_KEY);
			if (raw) dismissed = JSON.parse(raw);
		} catch {
			/* ignore */
		}
		try {
			const res = await encore.social.GetAnnouncements();
			announcements = (res.broadcasts || []) as Announcement[];
		} catch {
			/* ignore — banner is best-effort */
		}
	});

	const visible = $derived(announcements.find((a) => !dismissed.includes(a.id)) ?? null);

	function dismiss(id: string) {
		dismissed = [...dismissed, id];
		try {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(dismissed));
		} catch {
			/* ignore */
		}
	}
</script>

{#if visible}
	<div class="border-b bg-primary/5">
		<div class="max-w-7xl mx-auto px-4 py-2.5 flex items-start gap-3">
			<Megaphone class="size-4 text-primary mt-0.5 shrink-0" />
			<div class="flex-1 min-w-0">
				<p class="text-sm font-semibold">{visible.title}</p>
				{#if visible.body}
					<p class="text-xs text-muted-foreground mt-0.5">{visible.body}</p>
				{/if}
			</div>
			<button
				type="button"
				onclick={() => dismiss(visible.id)}
				class="text-muted-foreground hover:text-foreground shrink-0"
				aria-label="Dismiss announcement"
			>
				<X class="size-4" />
			</button>
		</div>
	</div>
{/if}
