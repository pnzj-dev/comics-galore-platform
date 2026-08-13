<script lang="ts">
	import { Circle, Medal, Shield, Crown, Gem } from 'lucide-svelte';

	let { tier }: { tier: string } = $props();

	const config = $derived.by(() => {
		switch (tier?.toLowerCase()) {
			case 'bronze':
				return { icon: Medal, cls: 'bg-amber-700/20 text-amber-700 dark:text-amber-500', label: 'Bronze' };
			case 'silver':
				return { icon: Shield, cls: 'bg-slate-400/20 text-slate-500 dark:text-slate-300', label: 'Silver' };
			case 'gold':
				return { icon: Crown, cls: 'bg-yellow-500/20 text-yellow-600 dark:text-yellow-400', label: 'Gold' };
			case 'platinum':
				return { icon: Gem, cls: 'bg-violet-500/20 text-violet-600 dark:text-violet-400', label: 'Platinum' };
			case 'free':
			default:
				return { icon: Circle, cls: 'bg-slate-200/60 text-slate-500 dark:text-slate-400', label: 'Free' };
		}
	});

	const Icon = $derived(config.icon);
</script>

<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium {config.cls}">
	<Icon class="size-3" />
	{config.label}
</span>
