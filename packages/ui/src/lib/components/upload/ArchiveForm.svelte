<script lang="ts">
	import { extractComicMetadata } from '$lib/archive/metadata';
	import { extractArchivePages } from '$lib/archive/pages';
	import { uploadImage, type UploadMode } from '$lib/archive/upload';
	import { processComicArchive, type ComicFormData, type SeriesInput, type StepStatus } from '$lib/archive/process';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';
	import SeriesField, { type SeriesValue } from './SeriesField.svelte';
	import { goto } from '$app/navigation';
	import { languageOptions, ageRatingOptions, readingDirectionOptions, optionLabel } from './options.js';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { LoaderCircle, Check, X, ImagePlus } from 'lucide-svelte';

	let {
		uploadMode = 'backend',
		uploadPartSizeMB = 100,
		uploadConcurrency = 4,
	}: { uploadMode?: UploadMode; uploadPartSizeMB?: number; uploadConcurrency?: number } = $props();

	let archiveFile = $state<File | null>(null);
	let extracting = $state(false);
	let metadataFound = $state(false);
	let pageCount = $state(0);
	let error = $state('');
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;

	let title = $state('');
	let description = $state('');
	let category = $state('');
	let genre = $state('');
	let author = $state('');
	let contentLanguage = $state('en');
	let ageRating = $state('all_ages');
	let tags = $state('');
	let readingDirection = $state<'ltr' | 'rtl'>('ltr');
	let isbn = $state('');
	let upc = $state('');
	let issn = $state('');
	let volume = $state('');
	let issueNumber = $state('');

	let series = $state<SeriesValue>({});

	let coverKey = $state('');
	let coverPreview = $state('');
	let coverProgress = $state(0);
	let coverUploading = $state(false);

	let previewSlots = $state<{ file: File | null; preview: string; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
	]);

	let submitting = $state(false);
	let step = $state<'upload' | 'metadata'>('upload');
	let dragOver = $state(false);

	type StepDef = { id: string; label: string; status: StepStatus; message?: string };
	let steps = $state<StepDef[]>([]);
	const stepDefs = [
		{ id: 'archive', label: 'Uploading archive' },
		{ id: 'pages', label: 'Extracting pages' },
		{ id: 'upload', label: 'Uploading pages' },
		{ id: 'publish', label: 'Publishing comic' },
	];

	function setStep(id: string, status: StepStatus, message?: string) {
		const existing = steps.find((s) => s.id === id);
		if (existing) {
			existing.status = status;
			existing.message = message;
		} else {
			steps.push({ id, label: stepDefs.find((d) => d.id === id)?.label || id, status, message });
		}
	}

	function formatSize(bytes: number): string {
		if (bytes > 1048576) return (bytes / 1048576).toFixed(1) + ' MB';
		if (bytes > 1024) return Math.round(bytes / 1024) + ' KB';
		return bytes + ' B';
	}

	async function handleCover(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		coverPreview = URL.createObjectURL(file);
		coverUploading = true;
		try {
			coverKey = await uploadImage(file, (pct) => (coverProgress = pct), uploadMode);
		} catch {
			error = 'Cover upload failed';
		} finally {
			coverUploading = false;
		}
	}

	async function handlePreview(e: Event, i: number) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		previewSlots[i].file = file;
		previewSlots[i].preview = URL.createObjectURL(file);
		previewSlots[i].uploading = true;
		try {
			previewSlots[i].key = await uploadImage(file, (pct) => (previewSlots[i].progress = pct), uploadMode);
		} catch {
			error = 'Preview upload failed';
		} finally {
			previewSlots[i].uploading = false;
		}
	}

	function addPreview() {
		if (previewSlots.length < 10) previewSlots.push({ file: null, preview: '', key: '', progress: 0, uploading: false });
	}
	function removePreview(i: number) {
		if (previewSlots.length > 2) {
			previewSlots.splice(i, 1);
		} else {
			previewSlots[i] = { file: null, preview: '', key: '', progress: 0, uploading: false };
		}
	}

	async function handleFile(file: File) {
		archiveFile = file;
		extracting = true;
		metadataFound = false;
		error = '';

		try {
			const meta = await extractComicMetadata(file);
			if (meta) {
				metadataFound = true;
				if (meta.title) title = meta.title;
				if (meta.author) author = meta.author;
				if (meta.description) description = meta.description;
				if (meta.content_language || meta.language) contentLanguage = (meta.content_language || meta.language)!;
				if (meta.age_rating) ageRating = meta.age_rating;
				if (meta.tags) tags = meta.tags.join(', ');
				if (meta.volume) volume = meta.volume;
				if (meta.issue_number) issueNumber = meta.issue_number;
			}
			if (!title) {
				title = file.name.replace(/\.[^.]+$/, '').replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
			}
			const pages = await extractArchivePages(file).catch(() => null);
			pageCount = pages?.length || 0;
		} catch (err) {
			error = (err as Error).message;
		} finally {
			extracting = false;
		}
		step = 'metadata';
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		const file = e.dataTransfer?.files?.[0];
		if (file) handleFile(file);
	}
	function onFileSelect(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (file) handleFile(file);
	}

	function buildFormData(): ComicFormData {
		return {
			title,
			author,
			description,
			content_language: contentLanguage,
			category,
			genre,
			age_rating: ageRating,
			tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
			reading_direction: readingDirection,
			isbn,
			upc,
			issn,
			volume,
			issue_number: issueNumber,
			is_premium: false,
			min_tier_id: '',
		};
	}

	async function submit() {
		if (!archiveFile) { error = 'Archive file is required'; return; }
		if (!title) { error = 'Title is required'; return; }
		if (!coverKey) { error = 'Cover image is required'; return; }
		const previewKeys = previewSlots.filter((p) => p.key).map((p) => p.key);
		if (previewKeys.length < 2) { error = 'At least 2 preview images are required'; return; }
		if (series.series_title && !series.series_title.trim()) { error = 'Series title is required when creating a new series'; return; }

		submitting = true;
		error = '';
		steps = [];

		try {
			await processComicArchive({
				archiveFile,
				coverKey,
				previewKeys,
				form: buildFormData(),
				series: series as SeriesInput,
				uploadMode,
				uploadPartSizeMB,
				uploadConcurrency,
				turnstileToken: turnstileToken || '',
				onStep: (id, status, message) => setStep(id, status, message),
			});
			await goto('/upload?tab=list');
		} catch (err) {
			error = (err as Error).message || 'Failed to create comic';
			const active = steps.find((s) => s.status === 'active');
			if (active) active.status = 'error';
		} finally {
			submitting = false;
			turnstileToken = null;
			turnstileReset++;
		}
	}

	function reset() {
		archiveFile = null;
		extracting = false;
		metadataFound = false;
		pageCount = 0;
		title = ''; description = ''; tags = ''; author = '';
		readingDirection = 'ltr'; isbn = ''; upc = ''; issn = ''; volume = ''; issueNumber = '';
		category = ''; genre = '';
		coverKey = ''; coverPreview = '';
		previewSlots = [
			{ file: null, preview: '', key: '', progress: 0, uploading: false },
			{ file: null, preview: '', key: '', progress: 0, uploading: false },
		];
		series = {};
		step = 'upload';
	}
</script>

<div class="max-w-2xl mx-auto">
	{#if step === 'upload'}
		<div
			class="border-2 border-dashed rounded-xl p-12 text-center transition-colors {dragOver ? 'border-primary bg-primary/5' : archiveFile ? 'border-green-500 bg-green-50 dark:bg-green-900/10' : 'border-border hover:border-primary/50'}"
			ondragover={(e) => { e.preventDefault(); dragOver = true; }}
			ondragleave={() => (dragOver = false)}
			ondrop={onDrop}
			onclick={() => document.getElementById('archive-input')?.click()}
			role="button"
			tabindex="0"
			onkeydown={(e) => { if (e.key === 'Enter') document.getElementById('archive-input')?.click(); }}
		>
			<input id="archive-input" type="file" accept=".cbr,.cbz,.zip,.rar,.7z,.pdf" onchange={onFileSelect} class="hidden" />

			{#if extracting}
				<div class="space-y-3">
					<div class="w-12 h-12 border-3 border-primary/30 border-t-primary rounded-full animate-spin mx-auto"></div>
					<p class="text-sm text-muted-foreground">Extracting metadata and pages…</p>
				</div>
			{:else}
				<div class="space-y-2">
					<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-muted-foreground mx-auto"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>
					<p class="text-sm text-muted-foreground">Drop an archive file here or click to browse</p>
					<p class="text-xs text-muted-foreground">Supports .cbr, .cbz, .zip, .rar, .7z, .pdf</p>
				</div>
			{/if}
		</div>

		{#if error}
			<p class="text-sm text-destructive mt-2">{error}</p>
		{/if}
	{:else}
		<div class="space-y-4">
			<button onclick={reset} class="text-sm text-muted-foreground hover:text-foreground">&larr; Choose different file</button>

			<div class="p-3 rounded-lg bg-green-50 dark:bg-green-900/10 border border-green-200 dark:border-green-800 text-sm">
				<p class="font-medium text-green-700 dark:text-green-400">{archiveFile?.name}</p>
				<p class="text-xs text-muted-foreground">{formatSize(archiveFile?.size || 0)} — {pageCount > 0 ? `${pageCount} pages detected` : 'pages will be extracted on publish'}</p>
			</div>

			{#if metadataFound}
				<div class="p-3 rounded-lg bg-blue-50 dark:bg-blue-900/10 border border-blue-200 dark:border-blue-800 text-sm">
					<p class="font-medium text-blue-700 dark:text-blue-400">Metadata imported from comic.json</p>
					<p class="text-xs text-muted-foreground mt-0.5">Review and adjust the fields below before publishing.</p>
				</div>
			{/if}

			<div class="space-y-3">
				<div class="space-y-1.5">
					<Label for="title">Title *</Label>
					<Input id="title" bind:value={title} placeholder="Comic title" required />
				</div>

				<div class="space-y-1.5">
					<Label for="author">Author</Label>
					<Input id="author" bind:value={author} placeholder="Author name" />
				</div>

				<div class="space-y-1.5">
					<Label for="desc">Description</Label>
					<Textarea id="desc" bind:value={description} rows={2} placeholder="Short description..." />
				</div>

				<div class="flex gap-3">
					<div class="space-y-1.5 flex-1">
						<Label for="lang">Language</Label>
						<Select type="single" bind:value={contentLanguage}>
							<SelectTrigger id="lang" class="w-full">{optionLabel(languageOptions, contentLanguage)}</SelectTrigger>
							<SelectContent>
								{#each languageOptions as lang (lang.value)}
									<SelectItem value={lang.value}>{lang.label}</SelectItem>
								{/each}
							</SelectContent>
						</Select>
					</div>
					<div class="space-y-1.5 flex-1">
						<Label for="rating">Age Rating</Label>
						<Select type="single" bind:value={ageRating}>
							<SelectTrigger id="rating" class="w-full">{optionLabel(ageRatingOptions, ageRating)}</SelectTrigger>
							<SelectContent>
								{#each ageRatingOptions as rating (rating.value)}
									<SelectItem value={rating.value}>{rating.label}</SelectItem>
								{/each}
							</SelectContent>
						</Select>
					</div>
				</div>

				<div class="space-y-1.5">
					<Label for="tags">Tags (comma-separated)</Label>
					<Input id="tags" bind:value={tags} placeholder="sci-fi, action, adventure" />
				</div>

				<div class="flex gap-3">
					<div class="space-y-1.5 flex-1">
						<Label for="category">Category</Label>
						<Input id="category" bind:value={category} placeholder="e.g. Manga, Webcomic" />
					</div>
					<div class="space-y-1.5 flex-1">
						<Label for="genre">Genre</Label>
						<Input id="genre" bind:value={genre} placeholder="e.g. Action, Romance" />
					</div>
				</div>

				<div class="space-y-1.5">
					<Label for="rd">Reading Direction</Label>
					<Select type="single" bind:value={readingDirection}>
						<SelectTrigger id="rd" class="w-full">{optionLabel(readingDirectionOptions, readingDirection)}</SelectTrigger>
						<SelectContent>
							{#each readingDirectionOptions as direction (direction.value)}
								<SelectItem value={direction.value}>{direction.label}</SelectItem>
							{/each}
						</SelectContent>
					</Select>
				</div>

				<div class="space-y-1.5">
					<Label>Identifiers (optional — set one)</Label>
					<div class="grid grid-cols-3 gap-2">
						<Input bind:value={isbn} placeholder="ISBN (graphic novels)" />
						<Input bind:value={upc} placeholder="UPC (single issues)" />
						<Input bind:value={issn} placeholder="ISSN" />
					</div>
				</div>

				<div class="space-y-1.5">
					<Label>Volume / Chapter & Issue</Label>
					<div class="grid grid-cols-2 gap-2">
						<Input bind:value={volume} placeholder="Vol. 2 or Ch. 5" />
						<Input bind:value={issueNumber} placeholder="Issue number (e.g. #12)" />
					</div>
				</div>

				<div class="space-y-1.5">
					<Label>Series</Label>
					<SeriesField value={series} onChange={(v) => (series = v)} {genre} {category} />
				</div>

				<div class="space-y-1.5">
					<Label>Cover Image *</Label>
					<div class="aspect-[3/4] max-w-[240px] rounded-lg border-2 border-dashed border-border overflow-hidden relative bg-muted/30">
						{#if coverPreview}
							<img src={coverPreview} alt="Cover preview" class="w-full h-full object-cover" />
							<button type="button" onclick={() => { coverPreview = ''; coverKey = ''; }} class="absolute top-2 right-2 bg-red-500 text-white rounded-full size-5 flex items-center justify-center text-xs z-10">&times;</button>
							{#if coverUploading}
								<div class="absolute inset-0 bg-black/60 flex flex-col items-center justify-center text-white">
									<div class="w-12 h-12 border-2 border-white/30 border-t-white rounded-full animate-spin mb-2"></div>
									<span class="text-xs font-medium">{coverProgress}%</span>
								</div>
							{/if}
						{:else}
							<input type="file" accept="image/*" onchange={handleCover} class="absolute inset-0 opacity-0 cursor-pointer z-10" />
							<div class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
								<ImagePlus class="size-7" />
								<span class="text-xs mt-1">Add cover</span>
							</div>
						{/if}
					</div>
				</div>

				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<Label>Preview Images (min 2) — hosted on Cloudflare</Label>
						{#if previewSlots.length < 10}
							<button type="button" onclick={addPreview} class="text-xs text-primary hover:underline">+ Add</button>
						{/if}
					</div>
					<div class="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-3">
						{#each previewSlots as slot, i (i)}
							<div class="aspect-[2/3] rounded-lg border-2 border-dashed border-border overflow-hidden relative {slot.preview ? 'border-solid' : ''}">
								{#if slot.preview}
									<img src={slot.preview} alt="Preview" class="w-full h-full object-cover" />
									<button type="button" onclick={() => removePreview(i)} class="absolute top-1 right-1 bg-red-500 text-white rounded-full size-4 flex items-center justify-center text-[10px] z-10">&times;</button>
									{#if slot.uploading}
										<div class="absolute inset-0 bg-black/60 flex flex-col items-center justify-center text-white">
											<div class="w-8 h-8 border-2 border-white/30 border-t-white rounded-full animate-spin mb-1"></div>
											<span class="text-[10px]">{slot.progress}%</span>
										</div>
									{/if}
								{:else}
									<input type="file" accept="image/*" onchange={(e) => handlePreview(e, i)} class="absolute inset-0 opacity-0 cursor-pointer z-10" />
									<div class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
										<ImagePlus class="size-5" />
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>

				{#if submitting && steps.length > 0}
					<div class="space-y-1.5 rounded-lg border border-border p-4">
						<p class="text-sm font-medium mb-2">Processing…</p>
						{#each steps as s (s.id)}
							<div class="flex items-center gap-2 text-sm">
								{#if s.status === 'done'}
									<Check class="size-4 text-green-600 dark:text-green-400" />
								{:else if s.status === 'error'}
									<X class="size-4 text-destructive" />
								{:else}
									<LoaderCircle class="size-4 animate-spin text-muted-foreground" />
								{/if}
								<span class="{s.status === 'error' ? 'text-destructive' : 'text-muted-foreground'}">{s.label}</span>
								{#if s.message}<span class="text-xs text-muted-foreground">— {s.message}</span>{/if}
							</div>
						{/each}
					</div>
				{/if}

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}

				<Turnstile action="comic_upload" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />

				<Button class="w-full" onclick={submit} disabled={submitting || !title || (turnstileRequired && !turnstileToken)}>
					{submitting ? 'Publishing…' : 'Publish Comic'}
				</Button>
			</div>
		</div>
	{/if}
</div>
