<script lang="ts">
	import SeriesPicker from './SeriesPicker.svelte';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';

	export interface SeriesValue {
		series_id?: string;
		series_title?: string;
		series_genre?: string;
		series_category?: string;
		series_schedule_day?: string;
	}

	interface Props {
		value?: SeriesValue;
		onChange: (v: SeriesValue) => void;
		genre?: string;
		category?: string;
	}

	let { value = {}, onChange, genre = '', category = '' }: Props = $props();

	let mode = $state<'existing' | 'new'>('existing');
	let newTitle = $state('');
	let newSchedule = $state('');

	const scheduleDays = [
		{ value: '', label: 'No schedule' },
		{ value: 'mon', label: 'Monday' },
		{ value: 'tue', label: 'Tuesday' },
		{ value: 'wed', label: 'Wednesday' },
		{ value: 'thu', label: 'Thursday' },
		{ value: 'fri', label: 'Friday' },
		{ value: 'sat', label: 'Saturday' },
		{ value: 'sun', label: 'Sunday' },
		{ value: 'completed', label: 'Completed' },
	];

	function selectSeries(s: { id: string; title: string }) {
		mode = 'existing';
		onChange({ series_id: s.id });
	}

	function startNew() {
		mode = 'new';
		onChange({ series_title: newTitle, series_genre: genre, series_category: category, series_schedule_day: newSchedule });
	}

	function updateNew() {
		onChange({ series_title: newTitle.trim(), series_genre: genre, series_category: category, series_schedule_day: newSchedule });
	}
</script>

<div class="space-y-2">
	<div class="flex gap-2">
		<button
			type="button"
			onclick={() => (mode = 'existing')}
			class="rounded-full px-3 py-1.5 text-xs font-medium transition {mode === 'existing' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
		>
			Existing series
		</button>
		<button
			type="button"
			onclick={startNew}
			class="rounded-full px-3 py-1.5 text-xs font-medium transition {mode === 'new' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-muted/80'}"
		>
			Create new series
		</button>
	</div>

	{#if mode === 'existing'}
		<SeriesPicker value={value.series_id} onSelect={selectSeries} onCreateNew={startNew} />
	{:else}
		<div class="space-y-2">
			<div class="space-y-1.5">
				<Label for="series-title">Series title *</Label>
				<Input id="series-title" bind:value={newTitle} oninput={updateNew} placeholder="Series name" />
			</div>
			<div class="space-y-1.5">
				<Label for="series-schedule">Schedule (optional)</Label>
				<Select type="single" bind:value={newSchedule} onValueChange={() => updateNew()}>
					<SelectTrigger id="series-schedule" class="w-full">
						{scheduleDays.find((d) => d.value === newSchedule)?.label || 'No schedule'}
					</SelectTrigger>
					<SelectContent>
						{#each scheduleDays as day (day.value)}
							<SelectItem value={day.value}>{day.label}</SelectItem>
						{/each}
					</SelectContent>
				</Select>
			</div>
		</div>
	{/if}
</div>
