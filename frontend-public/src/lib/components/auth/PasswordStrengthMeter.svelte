<script lang="ts">
	import zxcvbn from 'zxcvbn';

	interface Props {
		password: string;
	}

	let { password }: Props = $props();

	const result = $derived(password ? zxcvbn(password) : null);
	const score = $derived(result ? result.score : 0);
	const labels = ['Too weak', 'Weak', 'Fair', 'Good', 'Strong'];
	const colors = ['bg-destructive', 'bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-green-500'];
	const width = $derived(result ? ((score + 1) / 5) * 100 : 0);
	const label = $derived(result ? labels[score] : '');
</script>

{#if result}
	<div class="space-y-1">
		<div class="h-1.5 w-full rounded-full bg-muted overflow-hidden">
			<div class="h-full rounded-full transition-all duration-300 {colors[score]}" style="width: {width}%"></div>
		</div>
		<p class="text-xs text-muted-foreground">Password strength: {label}</p>
	</div>
{/if}
