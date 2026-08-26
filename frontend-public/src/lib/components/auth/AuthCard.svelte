<script lang="ts">
	import { Card } from '$lib/components/ui/card/index.js';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<Card class="w-full max-w-md">
	<div class="flex flex-col items-center gap-2 text-center">
		<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-primary">
			<path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20" />
			<path d="M9 2v15" />
		</svg>
		<span class="text-xl font-bold tracking-tight">Comics Galore</span>
	</div>
	{@render children()}
</Card>
