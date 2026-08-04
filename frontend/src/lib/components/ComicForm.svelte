<script lang="ts">
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { goto } from '$app/navigation';

	let title = $state('');
	let description = $state('');
	let synopsis = $state('');
	let category = $state('');
	let author = $state('');
	let contentLanguage = $state('en');
	let ageRating = $state('all_ages');
	let tags = $state('');
	let error = $state('');
	let submitting = $state(false);

	// Cover
	let coverFile = $state<File | null>(null);
	let coverPreview = $state('');
	let coverKey = $state('');
	let coverProgress = $state(0);
	let coverUploading = $state(false);

	// Previews (min 2, max 10)
	let previewSlots = $state<{ file: File | null; preview: string; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
		{ file: null, preview: '', key: '', progress: 0, uploading: false }
	]);

	// Archives (min 1, max 10)
	let archiveSlots = $state<{ file: File | null; name: string; size: number; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, name: '', size: 0, key: '', progress: 0, uploading: false }
	]);

	async function uploadFile(file: File, slotKey: string, onProgress: (pct: number) => void): Promise<string> {
		const session = await api.post<{ id: string }>('/upload-sessions', { mode: 'manual' });
		const presign = await api.post<{ url: string; key: string }>(`/upload-sessions/${session.id}/presign`, { number: 1, key: file.name });

		const xhr = new XMLHttpRequest();
		await new Promise<void>((resolve, reject) => {
			xhr.upload.addEventListener('progress', (e) => {
				if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
			});
			xhr.addEventListener('load', () => xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error('Upload failed')));
			xhr.addEventListener('error', () => reject(new Error('Upload failed')));
			xhr.open('PUT', presign.url);
			xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
			xhr.send(file);
		});

		await api.post(`/upload-sessions/${session.id}/parts`, {
			number: 1, key: presign.key, size: file.size,
			etag: xhr.getResponseHeader('ETag') || ''
		});

		return presign.key;
	}

	// Cover handlers
	async function handleCover(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		coverFile = file;
		coverPreview = URL.createObjectURL(file);
		coverUploading = true;
		try {
			coverKey = await uploadFile(file, 'cover', (pct) => coverProgress = pct);
		} catch (err) { error = 'Cover upload failed: ' + (err as Error).message; }
		finally { coverUploading = false; }
	}

	function removeCover() { coverFile = null; coverPreview = ''; coverKey = ''; coverProgress = 0; }

	// Preview handlers
	async function handlePreview(e: Event, i: number) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		previewSlots[i].file = file;
		previewSlots[i].preview = URL.createObjectURL(file);
		previewSlots[i].uploading = true;
		try {
			previewSlots[i].key = await uploadFile(file, `preview-${i}`, (pct) => previewSlots[i].progress = pct);
		} catch (err) { error = 'Preview upload failed: ' + (err as Error).message; }
		finally { previewSlots[i].uploading = false; }
	}

	function addPreview() { if (previewSlots.length < 10) previewSlots.push({ file: null, preview: '', key: '', progress: 0, uploading: false }); }
	function removePreview(i: number) { if (previewSlots.length > 2) previewSlots.splice(i, 1); }

	// Archive handlers
	async function handleArchive(e: Event, i: number) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		archiveSlots[i].file = file;
		archiveSlots[i].name = file.name;
		archiveSlots[i].size = file.size;
		archiveSlots[i].uploading = true;
		try {
			archiveSlots[i].key = await uploadFile(file, `archive-${i}`, (pct) => archiveSlots[i].progress = pct);
		} catch (err) { error = 'Archive upload failed: ' + (err as Error).message; }
		finally { archiveSlots[i].uploading = false; }
	}

	function addArchive() { if (archiveSlots.length < 10) archiveSlots.push({ file: null, name: '', size: 0, key: '', progress: 0, uploading: false }); }
	function removeArchive(i: number) { if (archiveSlots.length > 1) archiveSlots.splice(i, 1); }

	function formatSize(bytes: number): string {
		if (bytes > 1048576) return (bytes / 1048576).toFixed(1) + ' MB';
		if (bytes > 1024) return Math.round(bytes / 1024) + ' KB';
		return bytes + ' B';
	}

	async function submit() {
		if (!title) { error = 'Title is required'; return; }
		if (!coverKey) { error = 'Cover image is required'; return; }
		const archiveKeys = archiveSlots.filter(a => a.key).map(a => a.key);
		if (archiveKeys.length === 0) { error = 'At least one archive file is required'; return; }

		submitting = true;
		error = '';

		try {
			const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);
			const previewKeys = previewSlots.filter(p => p.key).map(p => p.key);

			await api.post('/comics', {
				title,
				description,
				content_language: contentLanguage,
				cover_key: coverKey,
				file_key: archiveKeys[0],
				page_keys: [...previewKeys, ...archiveKeys],
				file_size_bytes: archiveSlots.reduce((sum, a) => sum + (a.size || 0), 0),
				age_rating: ageRating,
				tags: tagList
			});

			await goto('/upload');
		} catch (err) {
			error = (err as Error).message;
		} finally {
			submitting = false;
		}
	}
</script>

<Card class="max-w-3xl mx-auto">
	<CardHeader>
		<CardTitle>Create New Comic</CardTitle>
	</CardHeader>
	<CardContent>
		<form class="space-y-6" onsubmit={(e) => e.preventDefault()}>

			<!-- Text Fields -->
			<div class="grid sm:grid-cols-3 gap-4">
				<div class="sm:col-span-2 space-y-2">
					<Label for="title">Title *</Label>
					<Input id="title" bind:value={title} placeholder="Comic title" required />
				</div>
				<div class="space-y-2">
					<Label for="author">Author</Label>
					<Input id="author" bind:value={author} placeholder="Author name" />
				</div>
			</div>

			<div class="grid sm:grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label for="description">Description</Label>
					<textarea id="description" bind:value={description} rows="2" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Short description..."></textarea>
				</div>
				<div class="space-y-2">
					<Label for="synopsis">Synopsis</Label>
					<textarea id="synopsis" bind:value={synopsis} rows="2" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Full synopsis..."></textarea>
				</div>
			</div>

			<div class="grid sm:grid-cols-4 gap-4">
				<div class="space-y-2">
					<Label for="lang">Language</Label>
					<select id="lang" bind:value={contentLanguage} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="en">English</option>
						<option value="ja">Japanese</option>
						<option value="es">Spanish</option>
						<option value="ko">Korean</option>
						<option value="fr">French</option>
					</select>
				</div>
				<div class="space-y-2">
					<Label for="rating">Age Rating</Label>
					<select id="rating" bind:value={ageRating} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="all_ages">All Ages</option>
						<option value="teen">Teen</option>
						<option value="mature">Mature</option>
						<option value="explicit">Explicit</option>
					</select>
				</div>
				<div class="space-y-2">
					<Label for="category">Category</Label>
					<Input id="category" bind:value={category} placeholder="e.g. Manga" />
				</div>
				<div class="space-y-2">
					<Label for="tags">Tags</Label>
					<Input id="tags" bind:value={tags} placeholder="comma-separated" />
				</div>
			</div>

			<!-- Cover Image -->
			<div class="space-y-2">
				<Label>Cover Image *</Label>
				<div class="relative">
					<Input type="file" accept="image/*" onchange={handleCover} disabled={coverUploading} />
					{#if coverPreview}
						<div class="mt-2 relative inline-block">
							<img src={coverPreview} alt="Cover preview" class="h-32 rounded-lg border object-cover" />
							<button type="button" onclick={removeCover} class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full size-5 flex items-center justify-center text-xs">&times;</button>
							{#if coverUploading}
								<div class="absolute inset-0 bg-black/50 rounded-lg flex items-center justify-center text-white text-xs">{coverProgress}%</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- Preview Images -->
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<Label>Preview Images (min 2)</Label>
					{#if previewSlots.length < 10}
						<button type="button" onclick={addPreview} class="text-xs text-primary hover:underline">+ Add</button>
					{/if}
				</div>
				<div class="grid grid-cols-5 gap-3">
					{#each previewSlots as slot, i}
						<div class="aspect-[2/3] rounded-lg border-2 border-dashed border-border flex items-center justify-center overflow-hidden relative {slot.preview ? 'border-solid' : ''}">
							{#if slot.preview}
								<img src={slot.preview} alt="Preview" class="w-full h-full object-cover" />
								<button type="button" onclick={() => removePreview(i)} class="absolute top-1 right-1 bg-red-500 text-white rounded-full size-4 flex items-center justify-center text-[10px]">&times;</button>
								{#if slot.uploading}
									<div class="absolute inset-0 bg-black/50 flex items-center justify-center text-white text-xs">{slot.progress}%</div>
								{/if}
							{:else}
								<input type="file" accept="image/*" onchange={(e) => handlePreview(e, i)} class="absolute inset-0 opacity-0 cursor-pointer" />
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-muted-foreground"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
							{/if}
						</div>
					{/each}
				</div>
			</div>

			<!-- Archive Files -->
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<Label>Archive Files * (min 1)</Label>
					{#if archiveSlots.length < 10}
						<button type="button" onclick={addArchive} class="text-xs text-primary hover:underline">+ Add</button>
					{/if}
				</div>
				<div class="grid grid-cols-5 gap-3">
					{#each archiveSlots as slot, i}
						<div class="aspect-square rounded-lg border-2 border-dashed border-border flex flex-col items-center justify-center p-2 overflow-hidden relative {slot.key ? 'border-purple-300 bg-purple-50 dark:bg-purple-900/20' : ''}">
							{#if slot.key}
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-purple-500"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
								<span class="text-[10px] text-purple-600 font-medium truncate max-w-full mt-1">{slot.name}</span>
								<span class="text-[9px] text-muted-foreground">{formatSize(slot.size)}</span>
								<button type="button" onclick={() => removeArchive(i)} class="absolute top-1 right-1 bg-red-500 text-white rounded-full size-4 flex items-center justify-center text-[10px]">&times;</button>
							{:else if slot.uploading}
								<div class="text-xs text-muted-foreground text-center">{slot.progress}%</div>
							{:else}
								<input type="file" accept=".cbr,.cbz,.pdf,.zip,.rar,.7z" onchange={(e) => handleArchive(e, i)} class="absolute inset-0 opacity-0 cursor-pointer" />
								<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-muted-foreground"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
								<span class="text-[10px] text-muted-foreground mt-1">Archive</span>
							{/if}
						</div>
					{/each}
				</div>
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Button type="submit" class="w-full" disabled={submitting || !title} onclick={submit}>
				{submitting ? 'Creating...' : 'Publish Comic'}
			</Button>
		</form>
	</CardContent>
</Card>
