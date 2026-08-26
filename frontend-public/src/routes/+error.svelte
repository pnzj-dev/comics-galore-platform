<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
</script>

<svelte:head><title>{page.status} — Comics Galore</title></svelte:head>

<div class="flex min-h-screen flex-col items-center justify-center p-6 text-center">
	<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="text-primary mb-4">
		<path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20" />
		<path d="M9 2v15" />
	</svg>
	<h1 class="text-6xl font-extrabold tracking-tight">{page.status}</h1>
	<p class="mt-3 text-lg text-muted-foreground max-w-md">
		{page.error?.message || 'Something went wrong.'}
	</p>
	<div class="mt-8 flex gap-3">
		<Button href="/">Back to home</Button>
		<Button variant="outline" onclick={() => location.reload()}>Reload</Button>
	</div>
</div>
