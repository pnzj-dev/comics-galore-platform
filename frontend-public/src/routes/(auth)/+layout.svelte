<script lang="ts">
	let { children } = $props();
</script>

<div class="flex min-h-screen items-center justify-center p-4">
	{@render children()}
</div>
