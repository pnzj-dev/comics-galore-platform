<script lang="ts">
	import { ENV_LABEL, ENV_TOOLTIP, IS_NON_PROD } from '$lib/utils/env';

	let {
		variant = 'pill',
		onDark = false,
	}: {
		variant?: 'pill' | 'bar';
		onDark?: boolean;
	} = $props();
</script>

{#if IS_NON_PROD && variant === 'bar'}
	<div class="env-bar" aria-hidden="true"></div>
{:else if IS_NON_PROD && variant === 'pill'}
	<span class="env-pill {onDark ? 'env-pill--on-dark' : ''}" title={ENV_TOOLTIP}>{ENV_LABEL}</span>
{/if}
