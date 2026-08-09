<script lang="ts">
	import { encore } from '$lib/api/encore';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { goto } from '$app/navigation';

	let archiveFile = $state<File | null>(null);
	let uploadProgress = $state(0);
	let uploadKey = $state('');
	let uploading = $state(false);
	let error = $state('');

	// Metadata form fields
	let title = $state('');
	let description = $state('');
	let contentLanguage = $state('en');
	let ageRating = $state('all_ages');
	let tags = $state('');

	let submitting = $state(false);
	let step = $state<'upload' | 'metadata'>('upload');
	let dragOver = $state(false);

	function formatSize(bytes: number): string {
		if (bytes > 1048576) return (bytes / 1048576).toFixed(1) + ' MB';
		if (bytes > 1024) return Math.round(bytes / 1024) + ' KB';
		return bytes + ' B';
	}

	async function handleFile(file: File) {
		archiveFile = file;
		uploading = true;
		uploadProgress = 0;
		error = '';

		try {
			const session = await encore.upload.CreateSession({ mode: 'archive' });
			const presign = await encore.upload.PresignUpload(session.id, { number: 1, key: file.name });

			const xhr = new XMLHttpRequest();
			await new Promise<void>((resolve, reject) => {
				xhr.upload.addEventListener('progress', (e) => {
					if (e.lengthComputable) uploadProgress = Math.round((e.loaded / e.total) * 100);
				});
				xhr.addEventListener('load', () => xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error('Upload failed')));
				xhr.addEventListener('error', () => reject(new Error('Upload failed')));
				xhr.open('PUT', presign.url);
				xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
				xhr.send(file);
			});

			await encore.upload.ConfirmPart(session.id, { number: 1, key: presign.key, size: file.size, etag: xhr.getResponseHeader('ETag') || '' });
			uploadKey = presign.key;

			// Auto-fill title from filename
			if (!title) {
				title = file.name.replace(/\.[^.]+$/, '').replace(/[-_]/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
			}

			step = 'metadata';
		} catch (err) {
			error = (err as Error).message;
		} finally {
			uploading = false;
		}
	}

	function onDrop(e: DragEvent) {
		e.preventDefault(); dragOver = false;
		const file = e.dataTransfer?.files?.[0];
		if (file) handleFile(file);
	}

	function onFileSelect(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (file) handleFile(file);
	}

	async function submit() {
		if (!title) { error = 'Title is required'; return; }
		submitting = true; error = '';

		try {
			const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);
			await encore.comics.CreateComic({
				title, description, content_language: contentLanguage,
				cover_key: uploadKey, file_key: uploadKey,
				page_keys: [uploadKey],
				file_size_bytes: archiveFile?.size || 0,
				age_rating: ageRating as any, tags: tagList
			});
			await goto('/upload');
		} catch (err) {
			error = (err as Error).message;
		} finally {
			submitting = false;
		}
	}

	function reset() {
		archiveFile = null;
		uploadKey = '';
		uploadProgress = 0;
		title = ''; description = ''; tags = '';
		step = 'upload';
	}
</script>

<div class="max-w-2xl mx-auto">
	{#if step === 'upload'}
		<!-- Drop Zone -->
		<div
			class="border-2 border-dashed rounded-xl p-12 text-center transition-colors {dragOver ? 'border-primary bg-primary/5' : archiveFile ? 'border-green-500 bg-green-50 dark:bg-green-900/10' : 'border-border hover:border-primary/50'}"
			ondragover={(e) => { e.preventDefault(); dragOver = true; }}
			ondragleave={() => dragOver = false}
			ondrop={onDrop}
			onclick={() => document.getElementById('archive-input')?.click()}
			role="button"
			tabindex="0"
			onkeydown={(e) => { if (e.key === 'Enter') document.getElementById('archive-input')?.click(); }}
		>
			<input id="archive-input" type="file" accept=".cbr,.cbz,.zip,.rar,.7z,.pdf" onchange={onFileSelect} class="hidden" />

			{#if archiveFile}
				<div class="space-y-2">
					<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-green-500 mx-auto"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
					<p class="text-sm font-medium text-green-600">{archiveFile.name}</p>
					<p class="text-xs text-muted-foreground">{formatSize(archiveFile.size)}</p>
				</div>
			{:else if uploading}
				<div class="space-y-3">
					<div class="w-12 h-12 border-3 border-primary/30 border-t-primary rounded-full animate-spin mx-auto"></div>
					<div class="w-full max-w-xs mx-auto bg-muted rounded-full h-2 overflow-hidden">
						<div class="bg-primary h-full rounded-full transition-all duration-300" style="width:{uploadProgress}%"></div>
					</div>
					<p class="text-sm text-muted-foreground">{uploadProgress}% uploaded</p>
				</div>
			{:else}
				<div class="space-y-2">
					<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-muted-foreground mx-auto"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" x2="12" y1="3" y2="15"/></svg>
					<p class="text-sm text-muted-foreground">Drop an archive file here or click to browse</p>
					<p class="text-xs text-muted-foreground">Supports .cbr, .cbz, .zip, .rar, .7z, .pdf</p>
				</div>
			{/if}
		</div>

		{#if archiveFile && !uploading}
			<div class="flex gap-2 mt-4">
				<Button class="w-full" variant="outline" onclick={reset} size="sm">Choose Different File</Button>
			</div>
		{/if}

	{:else}
		<!-- Metadata Form -->
		<div class="space-y-4">
			<div class="flex items-center gap-2">
				<button onclick={reset} class="text-sm text-muted-foreground hover:text-foreground">&larr; Choose different file</button>
			</div>

			<div class="p-3 rounded-lg bg-green-50 dark:bg-green-900/10 border border-green-200 dark:border-green-800 text-sm">
				<p class="font-medium text-green-700 dark:text-green-400">{archiveFile?.name}</p>
				<p class="text-xs text-muted-foreground">{formatSize(archiveFile?.size || 0)} — Uploaded successfully</p>
			</div>

			<div class="space-y-1.5">
				<Label for="title">Title *</Label>
				<Input id="title" bind:value={title} placeholder="Comic title" required />
			</div>

			<div class="space-y-1.5">
				<Label for="desc">Description</Label>
				<textarea id="desc" bind:value={description} rows="2" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Short description..."></textarea>
			</div>

			<div class="flex gap-3">
				<div class="space-y-1.5 flex-1">
					<Label for="lang">Language</Label>
					<select id="lang" bind:value={contentLanguage} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="en">English</option>
						<option value="ja">Japanese</option>
						<option value="es">Spanish</option>
						<option value="ko">Korean</option>
						<option value="fr">French</option>
					</select>
				</div>
				<div class="space-y-1.5 flex-1">
					<Label for="rating">Age Rating</Label>
					<select id="rating" bind:value={ageRating} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="all_ages">All Ages</option>
						<option value="teen">Teen</option>
						<option value="mature">Mature</option>
						<option value="explicit">Explicit</option>
					</select>
				</div>
			</div>

			<div class="space-y-1.5">
				<Label for="tags">Tags (comma-separated)</Label>
				<Input id="tags" bind:value={tags} placeholder="sci-fi, action, adventure" />
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Button class="w-full" onclick={submit} disabled={submitting || !title}>
				{submitting ? 'Publishing...' : 'Publish Comic'}
			</Button>
		</div>
	{/if}
</div>
