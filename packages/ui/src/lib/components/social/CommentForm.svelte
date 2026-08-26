<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	interface Props {
		onSubmit: (bodyText: string, parentId?: string) => Promise<void>;
		parentId?: string;
		placeholder?: string;
	}

	let { onSubmit, parentId, placeholder = 'Write a comment...' }: Props = $props();

	let bodyText = $state('');
	let submitting = $state(false);

	async function handleSubmit() {
		if (!bodyText.trim()) return;
		submitting = true;
		try {
			await onSubmit(bodyText.trim(), parentId);
			bodyText = '';
		} catch {}
		submitting = false;
	}
</script>

<div class="flex gap-2">
	<Textarea
		bind:value={bodyText}
		placeholder={placeholder}
		rows={2}
		class="flex-1 resize-none"
		onkeydown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSubmit(); } }}
	/>
	<Button size="sm" onclick={handleSubmit} disabled={submitting || !bodyText.trim()} class="self-end">
		{#if submitting}Posting...{:else}Post{/if}
	</Button>
</div>
