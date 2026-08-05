<script lang="ts">
	interface Props {
		id: string;
		title: string;
		slug: string;
		status?: string;
		cover_key?: string;
		cover_url?: string;
		is_premium?: boolean;
		view_count?: number;
		uploader_name?: string;
	}

	let {
		id,
		title,
		slug,
		status = '',
		cover_key = '',
		cover_url = '',
		is_premium = false,
		view_count = 0,
		uploader_name = ''
	}: Props = $props();

	function compactNum(n: number): string {
		if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
		if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k';
		return String(n);
	}

	function statusColor(s: string): string {
		switch (s) {
			case 'published': return 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400';
			case 'pending_review': return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400';
			case 'rejected': return 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400';
			default: return 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400';
		}
	}

	function coverSrc(): string {
		if (cover_url) return cover_url;
		return '';
	}
</script>

<div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden hover:border-purple-300 dark:hover:border-purple-700 transition-all flex flex-col">
	<a href="/comics/{id}" class="block relative">
		<div class="aspect-[2/3] bg-gray-100 dark:bg-gray-700 relative overflow-hidden">
			{#if coverSrc()}
				<img
					src={coverSrc()}
					alt={title}
					onerror={(e) => {
						const img = e.target as HTMLImageElement;
						img.style.display = 'none';
						const fallback = img.nextElementSibling as HTMLElement;
						if (fallback) fallback.style.display = 'flex';
					}}
					class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
					loading="lazy"
				/>
			{/if}
			<div class="w-full h-full items-center justify-center text-gray-400 dark:text-gray-500 {coverSrc() ? 'hidden' : 'flex'}">
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="size-8"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
			</div>
			{#if is_premium}
				<div class="absolute top-1.5 right-1.5 bg-yellow-400 text-black rounded-full p-0.5">
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="size-3"><path d="M11.562 3.266a.506.506 0 0 1 .876 0L15.39 8.87a1 1 0 0 0 .798.58l5.264.562a.506.506 0 0 1 .287.892l-3.913 3.52a1 1 0 0 0-.313.907l1.053 5.653a.506.506 0 0 1-.747.536L13 18.31a1 1 0 0 0-.976 0l-4.81 2.71a.506.506 0 0 1-.747-.536l1.053-5.653a1 1 0 0 0-.313-.907l-3.913-3.52a.506.506 0 0 1 .287-.892l5.264-.562a1 1 0 0 0 .798-.58z"/></svg>
				</div>
			{/if}
		</div>
	</a>
	<div class="p-2 flex-1 flex flex-col">
		<a href="/comics/{id}" class="block">
			<h3 class="text-xs font-medium leading-tight line-clamp-1 hover:text-purple-500 transition-colors">{title}</h3>
		</a>
		<div class="mt-auto pt-1.5 flex items-center justify-between">
			<div class="flex items-center gap-1.5">
				<span class="text-[10px] px-1.5 py-0.5 rounded-full font-medium {statusColor(status)}">{status.replace('_', ' ')}</span>
				{#if uploader_name}
					<span class="text-[10px] text-gray-400 truncate max-w-[80px]" title={uploader_name}>{uploader_name}</span>
				{/if}
			</div>
			<span class="flex items-center gap-0.5 text-[10px] text-gray-400">
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0"/><circle cx="12" cy="12" r="3"/></svg>
				<span>{compactNum(view_count)}</span>
			</span>
		</div>
	</div>
</div>
