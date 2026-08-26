<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Eye, EyeOff } from 'lucide-svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props {
		id?: string;
		value?: string;
		placeholder?: string;
		required?: boolean;
		autocomplete?: HTMLInputAttributes['autocomplete'];
		class?: string;
	}

	let {
		id,
		value = $bindable(''),
		placeholder,
		required = false,
		autocomplete,
		class: className = '',
	}: Props = $props();

	let visible = $state(false);
</script>

<div class="relative {className}">
	<Input
		{id}
		bind:value
		type={visible ? 'text' : 'password'}
		{placeholder}
		{required}
		{autocomplete}
		class="pr-10"
	/>
	<button
		type="button"
		onclick={() => (visible = !visible)}
		class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground transition-colors"
		aria-label={visible ? 'Hide password' : 'Show password'}
	>
		{#if visible}
			<EyeOff class="size-4" />
		{:else}
			<Eye class="size-4" />
		{/if}
	</button>
</div>
