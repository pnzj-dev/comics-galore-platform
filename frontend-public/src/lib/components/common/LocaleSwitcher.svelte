<script lang="ts">
	import { state as i18n, setLocale, ENABLED_LOCALES, LOCALE_META } from '$lib/i18n';
	import { Languages } from 'lucide-svelte';

	let { compact = false } = $props<{ compact?: boolean }>();
	let open = $state(false);

	const currentLabel = $derived(LOCALE_META[i18n.locale]?.label ?? 'English');

	function choose(code: (typeof ENABLED_LOCALES)[number]) {
		setLocale(code);
		open = false;
	}
</script>

<div class="relative">
	<button
		onclick={(e) => {
			e.stopPropagation();
			open = !open;
		}}
		class="inline-flex items-center gap-1.5 {compact ? 'p-2' : 'px-2 py-1'} rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
		aria-label="Change language"
	>
		<Languages class="size-4" />
		{#if !compact}
			<span class="text-sm">{currentLabel}</span>
		{/if}
	</button>

	{#if open}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="absolute right-0 mt-1 z-50 w-44 rounded-lg border border-border bg-background shadow-lg py-1"
			role="menu"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
		>
			{#each ENABLED_LOCALES as code}
				<button
					class="w-full text-left px-3 py-1.5 text-sm hover:bg-muted transition-colors {code === i18n.locale ? 'font-medium text-primary' : ''}"
					onclick={() => choose(code)}
				>
					{LOCALE_META[code].label}
				</button>
			{/each}
		</div>
	{/if}
</div>

<svelte:window
	onclick={() => (open = false)}
/>