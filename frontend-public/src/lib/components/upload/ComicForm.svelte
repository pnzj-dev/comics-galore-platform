<script lang="ts">
	import { buildComicArchive } from '$lib/archive/build';
	import { uploadImage, type UploadMode } from '$lib/archive/upload';
	import { processComicArchive, type ComicFormData, type SeriesInput, type StepStatus } from '$lib/archive/process';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client as zodClient } from 'sveltekit-superforms/adapters';
	import { z } from 'zod';
	import { get } from 'svelte/store';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import SeriesField, { type SeriesValue } from './SeriesField.svelte';
	import { goto } from '$app/navigation';
	import { languageOptions, ageRatingOptions, readingDirectionOptions, optionLabel } from './options.js';
	import { formatBytes } from '$lib/utils/format';
	import Turnstile from '$lib/components/common/Turnstile.svelte';
	import { TURNSTILE_SITEKEY } from '$lib/utils/turnstile';
	import { LoaderCircle, Check, X, ImagePlus } from 'lucide-svelte';

	const comicSchema = z.object({
		title: z.string().trim().min(1, 'Title is required'),
		author: z.string(),
		description: z.string(),
		category: z.string(),
		genre: z.string(),
		contentLanguage: z.string(),
		ageRating: z.string(),
		tags: z.string(),
		readingDirection: z.string(),
		isbn: z.string(),
		upc: z.string(),
		issn: z.string(),
		volume: z.string(),
		issueNumber: z.string(),
	});

	let {
		uploadMode = 'backend',
		pagePreviewThreshold = 20,
		uploadPartSizeMB = 100,
		uploadConcurrency = 4,
	}: {
		uploadMode?: UploadMode;
		pagePreviewThreshold?: number;
		uploadPartSizeMB?: number;
		uploadConcurrency?: number;
	} = $props();

	let error = $state('');
	let submitting = $state(false);
	let turnstileToken = $state<string | null>(null);
	let turnstileReset = $state(0);

	const turnstileRequired = !!TURNSTILE_SITEKEY;

	let series = $state<SeriesValue>({});

	let coverKey = $state('');
	let coverPreview = $state('');
	let coverProgress = $state(0);
	let coverUploading = $state(false);

	let previewSlots = $state<{ file: File | null; preview: string; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
	]);

	let pageFiles = $state<{ file: File; name: string; preview: string }[]>([]);
	const showCompactList = $derived(pageFiles.length > pagePreviewThreshold);

	type StepDef = { id: string; label: string; status: StepStatus; message?: string };
	let steps = $state<StepDef[]>([]);

	const stepDefs = [
		{ id: 'build', label: 'Building archive' },
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

	async function handleCover(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		coverPreview = URL.createObjectURL(file);
		coverUploading = true;
		try {
			coverKey = await uploadImage(file, (pct) => (coverProgress = pct), uploadMode);
		} catch (err) {
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

	function handlePages(e: Event) {
		const files = Array.from((e.target as HTMLInputElement).files || []);
		for (const file of files) {
			if (!/\.(jpe?g|png|gif|webp|avif|bmp)$/i.test(file.name)) continue;
			// Only build object-URL thumbnails below the preview threshold;
			// above it we render a compact list to avoid loading many images.
			const preview = pageFiles.length < pagePreviewThreshold ? URL.createObjectURL(file) : '';
			pageFiles.push({ file, name: file.name, preview });
		}
		pageFiles.sort((a, b) => a.name.localeCompare(b.name, undefined, { numeric: true }));
	}
	function removePage(i: number) {
		const page = pageFiles[i];
		if (page.preview) URL.revokeObjectURL(page.preview);
		pageFiles.splice(i, 1);
	}

	function buildMetadata(): Record<string, unknown> {
		const f = get(form);
		return {
			title: f.title,
			author: f.author,
			description: f.description,
			category: f.category,
			genre: f.genre,
			language: f.contentLanguage,
			content_language: f.contentLanguage,
			age_rating: f.ageRating,
			reading_direction: f.readingDirection,
			tags: f.tags.split(',').map((t) => t.trim()).filter(Boolean),
			volume: f.volume,
			issue_number: f.issueNumber,
		};
	}

	function buildFormData(): ComicFormData {
		const f = get(form);
		return {
			title: f.title,
			author: f.author,
			description: f.description,
			content_language: f.contentLanguage,
			category: f.category,
			genre: f.genre,
			age_rating: f.ageRating,
			tags: f.tags.split(',').map((t) => t.trim()).filter(Boolean),
			reading_direction: f.readingDirection,
			isbn: f.isbn,
			upc: f.upc,
			issn: f.issn,
			volume: f.volume,
			issue_number: f.issueNumber,
			is_premium: false,
			min_tier_id: '',
		};
	}

	const { form, errors, enhance } = superForm(
		{
			title: '',
			author: '',
			description: '',
			category: '',
			genre: '',
			contentLanguage: 'en',
			ageRating: 'all_ages',
			tags: '',
			readingDirection: 'ltr',
			isbn: '',
			upc: '',
			issn: '',
			volume: '',
			issueNumber: '',
		},
		{
			SPA: true,
			validationMethod: 'submit-only',
			validators: zodClient(comicSchema),
			onUpdate: async ({ form: f }) => {
				if (!f.valid) return;
				if (!coverKey) { error = 'Cover image is required'; return; }
				const previewKeys = previewSlots.filter((p) => p.key).map((p) => p.key);
				if (previewKeys.length < 2) { error = 'At least 2 preview images are required'; return; }
				if (pageFiles.length === 0) { error = 'At least 1 page image is required'; return; }
				if (series.series_title && !series.series_title.trim()) { error = 'Series title is required when creating a new series'; return; }

				submitting = true;
				error = '';
				steps = [];

				try {
					setStep('build', 'active', 'Building archive with metadata…');
					const cbz = await buildComicArchive(buildMetadata(), pageFiles);
					setStep('build', 'done');

					await processComicArchive({
						archiveFile: cbz,
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
			},
		},
	);
</script>

<Card>
	<CardHeader>
		<CardTitle>Create New Comic</CardTitle>
	</CardHeader>
	<CardContent>
		<form method="POST" use:enhance class="space-y-6">
			<div class="grid md:grid-cols-[1fr_320px] gap-8">
				<div class="space-y-3">
					<div class="space-y-1.5">
						<Label for="title">Title *</Label>
						<Input id="title" bind:value={$form.title} placeholder="Comic title" />
						{#if $errors.title}<p class="text-xs text-destructive">{$errors.title}</p>{/if}
					</div>

					<div class="space-y-1.5">
						<Label for="author">Author</Label>
						<Input id="author" bind:value={$form.author} placeholder="Author name" />
					</div>

					<div class="space-y-1.5">
						<Label for="description">Description</Label>
						<Textarea id="description" bind:value={$form.description} rows={2} placeholder="Short description..." />
					</div>

					<div class="flex gap-3">
						<div class="space-y-1.5 flex-1">
							<Label for="lang">Language</Label>
							<Select type="single" bind:value={$form.contentLanguage}>
								<SelectTrigger id="lang" class="w-full">{optionLabel(languageOptions, $form.contentLanguage)}</SelectTrigger>
								<SelectContent>
									{#each languageOptions as lang (lang.value)}
										<SelectItem value={lang.value}>{lang.label}</SelectItem>
									{/each}
								</SelectContent>
							</Select>
						</div>
						<div class="space-y-1.5 flex-1">
							<Label for="rating">Age Rating</Label>
							<Select type="single" bind:value={$form.ageRating}>
								<SelectTrigger id="rating" class="w-full">{optionLabel(ageRatingOptions, $form.ageRating)}</SelectTrigger>
								<SelectContent>
									{#each ageRatingOptions as rating (rating.value)}
										<SelectItem value={rating.value}>{rating.label}</SelectItem>
									{/each}
								</SelectContent>
							</Select>
						</div>
					</div>

					<div class="space-y-1.5">
						<Label for="category">Category</Label>
						<Input id="category" bind:value={$form.category} placeholder="e.g. Manga, Webcomic" />
					</div>

					<div class="space-y-1.5">
						<Label for="genre">Genre</Label>
						<Input id="genre" bind:value={$form.genre} placeholder="e.g. Action, Romance, Drama" />
					</div>

					<div class="space-y-1.5">
						<Label for="tags">Tags (comma-separated)</Label>
						<Input id="tags" bind:value={$form.tags} placeholder="action, adventure, fantasy" />
					</div>

					<div class="space-y-1.5">
						<Label for="reading-direction">Reading Direction</Label>
						<Select type="single" bind:value={$form.readingDirection}>
							<SelectTrigger id="reading-direction" class="w-full">{optionLabel(readingDirectionOptions, $form.readingDirection)}</SelectTrigger>
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
							<Input bind:value={$form.isbn} placeholder="ISBN (graphic novels)" />
							<Input bind:value={$form.upc} placeholder="UPC (single issues)" />
							<Input bind:value={$form.issn} placeholder="ISSN" />
						</div>
					</div>

					<div class="space-y-1.5">
						<Label>Volume / Chapter & Issue</Label>
						<div class="grid grid-cols-2 gap-2">
							<Input bind:value={$form.volume} placeholder="Vol. 2 or Ch. 5" />
							<Input bind:value={$form.issueNumber} placeholder="Issue number (e.g. #12)" />
						</div>
					</div>

					<div class="space-y-1.5">
						<Label>Series</Label>
						<SeriesField value={series} onChange={(v) => (series = v)} genre={$form.genre} category={$form.category} />
					</div>
				</div>

				<div class="space-y-1.5">
					<Label>Cover Image *</Label>
					<div class="aspect-[3/4] rounded-lg border-2 border-dashed border-border overflow-hidden relative bg-muted/30">
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

			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<Label>Pages (comic page images) — stored on S3</Label>
					<input id="pages-input" type="file" accept="image/*" multiple onchange={handlePages} class="hidden" />
					<button type="button" onclick={() => document.getElementById('pages-input')?.click()} class="text-xs text-primary hover:underline flex items-center gap-1">
						<ImagePlus class="size-4" /> Add pages
					</button>
				</div>
				{#if pageFiles.length > 0}
					{#if showCompactList}
						<div class="rounded-lg border border-border divide-y divide-border max-h-72 overflow-y-auto">
							<div class="px-3 py-2 text-xs text-muted-foreground flex items-center justify-between sticky top-0 bg-background">
								<span>{pageFiles.length} pages</span>
							</div>
							{#each pageFiles as page, i (page.name + i)}
								<div class="flex items-center gap-2 px-3 py-1.5 text-sm">
									<span class="flex-1 truncate">{page.name}</span>
									<span class="text-xs text-muted-foreground shrink-0">{formatBytes(page.file.size)}</span>
									<button type="button" onclick={() => removePage(i)} class="text-muted-foreground hover:text-destructive shrink-0" aria-label="Remove page">&times;</button>
								</div>
							{/each}
						</div>
					{:else}
						<div class="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 gap-2">
							{#each pageFiles as page, i (page.name + i)}
								<div class="relative aspect-[2/3] rounded-lg overflow-hidden border border-border bg-muted">
									<img src={page.preview} alt={page.name} class="w-full h-full object-cover" />
									<button type="button" onclick={() => removePage(i)} class="absolute top-1 right-1 bg-red-500 text-white rounded-full size-4 flex items-center justify-center text-[10px] z-10">&times;</button>
								</div>
							{/each}
						</div>
					{/if}
				{:else}
					<div class="rounded-lg border-2 border-dashed border-border p-6 text-center text-sm text-muted-foreground">
						Select the comic's page images. They'll be packed into a downloadable archive with a metadata.json.
					</div>
				{/if}
			</div>

			{#if submitting && steps.length > 0}
				<div class="space-y-1.5 rounded-lg border border-border p-4">
					<p class="text-sm font-medium mb-2">Processing…</p>
					{#each steps as step (step.id)}
						<div class="flex items-center gap-2 text-sm">
							{#if step.status === 'done'}
								<Check class="size-4 text-green-600 dark:text-green-400" />
							{:else if step.status === 'error'}
								<X class="size-4 text-destructive" />
							{:else}
								<LoaderCircle class="size-4 animate-spin text-muted-foreground" />
							{/if}
							<span class="{step.status === 'error' ? 'text-destructive' : 'text-muted-foreground'}">{step.label}</span>
							{#if step.message}<span class="text-xs text-muted-foreground">— {step.message}</span>{/if}
						</div>
					{/each}
				</div>
			{/if}

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Turnstile action="comic_upload" onToken={(t) => (turnstileToken = t)} resetSignal={turnstileReset} />

			<Button type="submit" class="w-full" disabled={submitting || (turnstileRequired && !turnstileToken)}>
				{submitting ? 'Creating…' : 'Create Comic'}
			</Button>
		</form>
	</CardContent>
</Card>
