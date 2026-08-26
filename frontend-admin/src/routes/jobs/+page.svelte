<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { formatDate } from '$lib/utils/format';
	import { RefreshCw, AlertCircle, CheckCircle2, LoaderCircle } from 'lucide-svelte';

	let { data } = $props();

	// svelte-ignore state_referenced_locally
	let runs = $state(data.runs);

	const jobNames = ['ai-moderate-comics', 'waiting-pay-expiry', 'archive-extract', 'ai-moderation'];
	const statuses = ['', 'running', 'success', 'failed'];

	let runningNow = $state<string | null>(null);

	function filter(key: string, value: string) {
		const url = new URL(page.url);
		if (value) url.searchParams.set(key, value);
		else url.searchParams.delete(key);
		goto(url.pathname + url.search, { keepFocus: true });
	}

	async function reload() {
		try {
			const res = await encore.jobs.ListJobRuns({
				JobName: data.jobName,
				Status: data.status,
				Limit: 200,
			});
			runs = res.runs || [];
		} catch {}
	}

	async function runNow(name: string, fn: () => Promise<void>) {
		runningNow = name;
		try {
			await fn();
			await reload();
		} catch (e) {
			console.error('[jobs] run failed:', e);
		} finally {
			runningNow = null;
		}
	}

	function statusBadge(status: string): string {
		switch (status) {
			case 'success': return 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300';
			case 'failed': return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300';
			case 'running': return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300';
			default: return 'bg-muted text-muted-foreground';
		}
	}
</script>

<svelte:head><title>Background Jobs - Comics Galore Admin</title></svelte:head>

<section>
	<h1 class="text-3xl font-bold mb-6">Background Jobs</h1>

	<div class="flex items-center gap-3 mb-4 flex-wrap">
		<select
			value={data.jobName}
			onchange={(e) => filter('job_name', (e.target as HTMLSelectElement).value)}
			class="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
		>
			<option value="">All jobs</option>
			{#each jobNames as n}
				<option value={n}>{n}</option>
			{/each}
		</select>
		<select
			value={data.status}
			onchange={(e) => filter('status', (e.target as HTMLSelectElement).value)}
			class="rounded-md border border-input bg-background px-3 py-1.5 text-sm"
		>
			<option value="">All statuses</option>
			{#each statuses.slice(1) as s}
				<option value={s}>{s}</option>
			{/each}
		</select>
		<div class="flex gap-2 ml-auto">
			<Button
				size="sm"
				variant="outline"
				disabled={runningNow !== null}
				onclick={() => runNow('ai-moderate-comics', () => encore.comics.RunAIModerationSweep())}
			>
				{#if runningNow === 'ai-moderate-comics'}<LoaderCircle class="size-4 animate-spin" />{:else}<RefreshCw class="size-4" />{/if}
				Run AI sweep
			</Button>
			<Button
				size="sm"
				variant="outline"
				disabled={runningNow !== null}
				onclick={() => runNow('waiting-pay-expiry', () => encore.billing.RunWaitingPayExpiry())}
			>
				{#if runningNow === 'waiting-pay-expiry'}<LoaderCircle class="size-4 animate-spin" />{:else}<RefreshCw class="size-4" />{/if}
				Run payment expiry
			</Button>
		</div>
	</div>

	<Card>
		<CardHeader class="pb-3">
			<CardTitle>Recent runs</CardTitle>
		</CardHeader>
		<CardContent>
			{#if runs.length === 0}
				<p class="text-sm text-muted-foreground text-center py-10">No job runs recorded yet.</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b text-left text-xs text-muted-foreground">
								<th class="py-2 pr-4 font-medium">Job</th>
								<th class="py-2 pr-4 font-medium">Ref</th>
								<th class="py-2 pr-4 font-medium">Status</th>
								<th class="py-2 pr-4 font-medium">Started</th>
								<th class="py-2 pr-4 font-medium">Finished</th>
								<th class="py-2 font-medium">Error</th>
							</tr>
						</thead>
						<tbody>
							{#each runs as run (run.id)}
								<tr class="border-b last:border-0">
									<td class="py-2 pr-4 font-mono text-xs">{run.job_name}</td>
									<td class="py-2 pr-4 text-xs text-muted-foreground truncate max-w-[160px]">{run.ref || '—'}</td>
									<td class="py-2 pr-4">
										<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(run.status)}">
											{#if run.status === 'failed'}<AlertCircle class="size-3" />{:else if run.status === 'success'}<CheckCircle2 class="size-3" />{/if}
											{run.status}
										</span>
									</td>
									<td class="py-2 pr-4 text-xs text-muted-foreground">{formatDate(run.started_at, 'datetime')}</td>
									<td class="py-2 pr-4 text-xs text-muted-foreground">{run.finished_at ? formatDate(run.finished_at, 'datetime') : '—'}</td>
									<td class="py-2 text-xs text-destructive truncate max-w-[260px]" title={run.error}>{run.error}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</CardContent>
	</Card>
</section>
