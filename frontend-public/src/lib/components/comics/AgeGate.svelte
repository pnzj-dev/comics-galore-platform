<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { modal } from '$lib/stores/modal.svelte';

	interface Props {
		title: string;
		author: string;
		ageRating: string;
		onConfirm: () => void;
		onClose: () => void;
	}

	let { title, author, ageRating, onConfirm, onClose }: Props = $props();

	const open = $derived(modal.isOpen('agegate'));

	function close() {
		modal.close('agegate');
		onClose();
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4">
		<div class="bg-background rounded-2xl shadow-xl w-full max-w-md p-6 text-center space-y-5">
			<div class="space-y-2">
				<h2 class="text-xl font-bold">Age-Restricted Content</h2>
				<p class="text-sm text-muted-foreground">
					This comic contains content intended for mature audiences.
				</p>
			</div>

			<div class="rounded-lg border border-border p-4 space-y-2 text-left">
				<h3 class="font-semibold text-lg">{title}</h3>
				<p class="text-sm text-muted-foreground">by {author}</p>
				<span class="inline-block text-xs font-bold uppercase px-2 py-0.5 rounded-full bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
					{ageRating.replace('_', ' ')}
				</span>
			</div>

			<div class="space-y-2">
				<Button class="w-full" onclick={onConfirm}>
					I'm 18+ years old, Continue
				</Button>
				<Button variant="outline" class="w-full" onclick={close}>
					Go back
				</Button>
			</div>

			<p class="text-xs text-muted-foreground">
				By continuing, you confirm that you are 18 years of age or older and wish to view mature content.
			</p>
		</div>
	</div>
{/if}
