<script lang="ts">
	let {
		value,
		onchange,
		debounce = 300,
		class: className = '',
		...rest
	}: {
		value: string | number;
		onchange: (value: string) => void;
		debounce?: number;
		class?: string;
		[key: string]: any;
	} = $props();

	let internalValue = $state(String(value ?? ''));

	$effect(() => {
		internalValue = String(value ?? '');
	});

	$effect(() => {
		const v = internalValue;
		const timeout = setTimeout(() => onchange(v), debounce);
		return () => clearTimeout(timeout);
	});
</script>

<input {...rest} class={className} bind:value={internalValue} />
