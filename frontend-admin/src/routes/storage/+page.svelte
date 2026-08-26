<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { formatBytes, formatNumber } from '$lib/utils/format';
	import { HardDrive, Image as ImageIcon, Boxes, AlertTriangle } from 'lucide-svelte';

	let { data } = $props();

	const usage = $derived(data.usage);
	const totalBytes = $derived(usage?.s3_total_bytes ?? 0);
	const breakdown = $derived(usage?.breakdown ?? []);
</script>

<svelte:head><title>Storage - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Storage</h1>

	{#if !usage}
		<div class="flex items-center gap-2 p-4 rounded-lg border border-yellow-200 bg-yellow-50 dark:bg-yellow-900/20 dark:border-yellow-800 text-sm">
			<AlertTriangle class="size-4 text-yellow-600 dark:text-yellow-400 shrink-0" />
			Storage usage is unavailable right now. Please try again.
		</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
			<Card>
				<CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground flex items-center gap-2"><HardDrive class="size-4" /> S3 object storage</CardTitle></CardHeader>
				<CardContent>
					<p class="text-3xl font-bold">{formatBytes(usage.s3_total_bytes)}</p>
					<p class="text-xs text-muted-foreground mt-1">{formatNumber(usage.s3_object_count)} objects</p>
				</CardContent>
			</Card>
			<Card>
				<CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground flex items-center gap-2"><ImageIcon class="size-4" /> Cloudflare Images</CardTitle></CardHeader>
				<CardContent>
					<p class="text-3xl font-bold">{usage.cf_configured ? formatNumber(usage.cf_images_count) : 'N/A'}</p>
					<p class="text-xs text-muted-foreground mt-1">{usage.cf_configured ? 'images' : 'not configured'}</p>
				</CardContent>
			</Card>
			<Card>
				<CardHeader class="pb-2"><CardTitle class="text-sm font-medium text-muted-foreground flex items-center gap-2"><Boxes class="size-4" /> Bucket breakdown</CardTitle></CardHeader>
				<CardContent>
					<p class="text-3xl font-bold">{formatNumber(breakdown.length)}</p>
					<p class="text-xs text-muted-foreground mt-1">key-prefix categories</p>
				</CardContent>
			</Card>
		</div>

		<Card>
			<CardHeader class="pb-3"><CardTitle>Breakdown by category</CardTitle></CardHeader>
			<CardContent>
				{#if breakdown.length === 0}
					<p class="text-sm text-muted-foreground text-center py-8">No objects in the bucket.</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b text-left text-xs text-muted-foreground">
									<th class="py-2 pr-4 font-medium">Category</th>
									<th class="py-2 pr-4 font-medium text-right">Objects</th>
									<th class="py-2 font-medium text-right">Size</th>
								</tr>
							</thead>
							<tbody>
								{#each breakdown as b (b.prefix)}
									<tr class="border-b last:border-0">
										<td class="py-2 pr-4 font-mono text-xs">{b.prefix}</td>
										<td class="py-2 pr-4 text-right text-muted-foreground">{formatNumber(b.object_count)}</td>
										<td class="py-2 text-right">{formatBytes(b.total_bytes)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</CardContent>
		</Card>
	{/if}
</section>
