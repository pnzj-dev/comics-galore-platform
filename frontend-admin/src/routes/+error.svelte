<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
</script>

<svelte:head><title>{page.status} — Comics Galore Admin</title></svelte:head>

<div class="flex min-h-[60vh] flex-col items-center justify-center p-6 text-center">
	<h1 class="text-6xl font-extrabold tracking-tight">{page.status}</h1>
	<p class="mt-3 text-lg text-muted-foreground max-w-md">
		{page.error?.message || 'Something went wrong.'}
	</p>
	<div class="mt-8 flex gap-3">
		<Button href="/dashboard">Back to dashboard</Button>
		<Button variant="outline" onclick={() => location.reload()}>Reload</Button>
	</div>
</div>
