<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { LoaderCircle, CheckCircle } from 'lucide-svelte';

	interface Props {
		text?: string;
		loadingText?: string;
		loading?: boolean;
		success?: boolean;
		disabled?: boolean;
		class?: string;
		onclick?: () => void;
	}

	let {
		text = 'Submit',
		loadingText = '...',
		loading = false,
		success = false,
		disabled = false,
		class: className = '',
		onclick
	}: Props = $props();

	let shown = $state(false);

	$effect(() => {
		if (success) {
			shown = true;
			const timer = setTimeout(() => shown = false, 2000);
			return () => clearTimeout(timer);
		} else {
			shown = false;
		}
	});
</script>

<Button onclick={onclick} class={className} disabled={disabled || loading}>
	{#if loading}
		<LoaderCircle class="size-4 animate-spin" />
		{loadingText}
	{:else if shown}
		<CheckCircle class="size-4 text-green-500" />
		{text}
	{:else}
		{text}
	{/if}
</Button>
