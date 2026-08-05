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

	let coverFile = $state<File | null>(null);
	let coverPreview = $state('');
	let coverKey = $state('');
	let coverProgress = $state(0);
	let coverUploading = $state(false);

	let previewSlots = $state<{ file: File | null; preview: string; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, preview: '', key: '', progress: 0, uploading: false },
		{ file: null, preview: '', key: '', progress: 0, uploading: false }
	]);

	let archiveSlots = $state<{ file: File | null; name: string; size: number; key: string; progress: number; uploading: boolean }[]>([
		{ file: null, name: '', size: 0, key: '', progress: 0, uploading: false }
	]);

	async function uploadFile(file: File, onProgress: (pct: number) => void, endpoint?: string): Promise<string> {
		// Cloudflare image upload (cover/preview)
		if (endpoint) {
			const res = await api.post<{ uploadURL: string; imageID: string }>(endpoint, {});
			const uploadUrl = res.uploadURL.endsWith('/') ? res.uploadURL : res.uploadURL;
			const xhr = new XMLHttpRequest();
			await new Promise<void>((resolve, reject) => {
				xhr.upload.addEventListener('progress', (e) => {
					if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
				});
				xhr.addEventListener('load', () => xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error('Upload failed')));
				xhr.addEventListener('error', () => reject(new Error('Upload failed')));
				xhr.open('PUT', uploadUrl);
				xhr.send(file);
			});
			return res.imageID;
		}

		// S3 upload session (archives)
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
		await api.post(`/upload-sessions/${session.id}/parts`, { number: 1, key: presign.key, size: file.size, etag: xhr.getResponseHeader('ETag') || '' });
		return presign.key;
	}

	async function handleCover(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		coverFile = file; coverPreview = URL.createObjectURL(file); coverUploading = true;
		try { coverKey = await uploadFile(file, (pct) => coverProgress = pct, '/media/cloudflare/upload-url'); } catch (err) { error = 'Cover upload failed'; }
		finally { coverUploading = false; }
	}
	function removeCover() { coverFile = null; coverPreview = ''; coverKey = ''; coverProgress = 0; }

	async function handlePreview(e: Event, i: number) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		previewSlots[i].file = file; previewSlots[i].preview = URL.createObjectURL(file); previewSlots[i].uploading = true;
		try { previewSlots[i].key = await uploadFile(file, (pct) => previewSlots[i].progress = pct, '/media/cloudflare/upload-url'); } catch (err) { error = 'Preview upload failed'; }
		finally { previewSlots[i].uploading = false; }
	}
	function addPreview() { if (previewSlots.length < 10) previewSlots.push({ file: null, preview: '', key: '', progress: 0, uploading: false }); }
	function removePreview(i: number) { if (previewSlots.length > 2) previewSlots.splice(i, 1); }

	async function handleArchive(e: Event, i: number) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		archiveSlots[i].file = file; archiveSlots[i].name = file.name; archiveSlots[i].size = file.size; archiveSlots[i].uploading = true;
		try { archiveSlots[i].key = await uploadFile(file, (pct) => archiveSlots[i].progress = pct); } catch (err) { error = 'Archive upload failed'; }
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
		submitting = true; error = '';
		try {
			const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);
			const previewKeys = previewSlots.filter(p => p.key).map(p => p.key);
			await api.post('/comics', {
				title, description, content_language: contentLanguage,
				cover_key: coverKey, file_key: archiveKeys[0],
				page_keys: [...previewKeys, ...archiveKeys],
				file_size_bytes: archiveSlots.reduce((sum, a) => sum + (a.size || 0), 0),
				age_rating: ageRating, tags: tagList
			});
			await goto('/upload');
		} catch (err) { error = (err as Error).message; }
		finally { submitting = false; }
	}
</script>

<Card class="max-w-5xl mx-auto">
	<CardHeader>
		<CardTitle>Create New Comic</CardTitle>
	</CardHeader>
	<CardContent>
		<form class="space-y-6" onsubmit={(e) => e.preventDefault()}>

			<!-- Row 1: Metadata left + Cover right -->
			<div class="grid md:grid-cols-[1fr_320px] gap-8">
				<!-- Metadata column -->
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
						<Label for="description">Description</Label>
						<textarea id="description" bind:value={description} rows="2" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Short description..."></textarea>
					</div>

					<div class="space-y-1.5">
						<Label for="synopsis">Synopsis</Label>
						<textarea id="synopsis" bind:value={synopsis} rows="2" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Full synopsis..."></textarea>
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
						<Label for="category">Category</Label>
						<Input id="category" bind:value={category} placeholder="e.g. Manga, Webcomic" />
					</div>

					<div class="space-y-1.5">
						<Label for="tags">Tags (comma-separated)</Label>
						<Input id="tags" bind:value={tags} placeholder="action, adventure, fantasy" />
					</div>
				</div>

				<!-- Cover column -->
				<div class="space-y-1.5">
					<Label>Cover Image *</Label>
					<div class="aspect-[3/4] rounded-lg border-2 border-dashed border-border overflow-hidden relative bg-muted/30">
						{#if coverPreview}
							<img src={coverPreview} alt="Cover preview" class="w-full h-full object-cover" />
							<button type="button" onclick={removeCover} class="absolute top-2 right-2 bg-red-500 text-white rounded-full size-5 flex items-center justify-center text-xs z-10">&times;</button>
							{#if coverUploading}
								<div class="absolute inset-0 bg-black/60 flex flex-col items-center justify-center text-white">
									<div class="w-12 h-12 border-2 border-white/30 border-t-white rounded-full animate-spin mb-2"></div>
									<span class="text-xs font-medium">{coverProgress}%</span>
								</div>
							{/if}
						{:else}
							<input type="file" accept="image/*" onchange={handleCover} class="absolute inset-0 opacity-0 cursor-pointer z-10" />
							<div class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
								<svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
								<span class="text-xs mt-1">Add cover</span>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Row 2: Preview Images (full width) -->
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<Label>Preview Images (min 2)</Label>
					{#if previewSlots.length < 10}
						<button type="button" onclick={addPreview} class="text-xs text-primary hover:underline">+ Add</button>
					{/if}
				</div>
				<div class="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-7 gap-3">
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
									<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</div>

			<!-- Row 3: Archive Files (full width) -->
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<Label>Archive Files (min 1)</Label>
					{#if archiveSlots.length < 10}
						<button type="button" onclick={addArchive} class="text-xs text-primary hover:underline">+ Add</button>
					{/if}
				</div>
				<div class="grid grid-cols-3 sm:grid-cols-5 gap-3">
					{#each archiveSlots as slot, i (i)}
						<div class="aspect-square rounded-lg border-2 border-dashed border-border overflow-hidden relative {slot.key ? 'border-purple-300 bg-purple-50 dark:bg-purple-900/20 border-solid' : ''}">
							{#if slot.key}
								<div class="w-full h-full flex flex-col items-center justify-center p-2 text-center">
									<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="text-purple-500 flex-shrink-0"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
									<span class="text-[10px] text-purple-600 font-medium truncate max-w-full mt-1">{slot.name}</span>
									<span class="text-[9px] text-muted-foreground">{formatSize(slot.size)}</span>
								</div>
								<button type="button" onclick={() => removeArchive(i)} class="absolute top-1 right-1 bg-red-500 text-white rounded-full size-4 flex items-center justify-center text-[10px] z-10">&times;</button>
							{:else if slot.uploading}
								<div class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
									<div class="w-8 h-8 border-2 border-primary/30 border-t-primary rounded-full animate-spin mb-1"></div>
									<span class="text-[10px]">{slot.progress}%</span>
								</div>
							{:else}
								<input type="file" accept=".cbr,.cbz,.pdf,.zip,.rar,.7z" onchange={(e) => handleArchive(e, i)} class="absolute inset-0 opacity-0 cursor-pointer z-10" />
								<div class="w-full h-full flex flex-col items-center justify-center text-muted-foreground">
									<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" x2="12" y1="15" y2="3"/></svg>
									<span class="text-[10px] mt-1">Archive</span>
								</div>
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
