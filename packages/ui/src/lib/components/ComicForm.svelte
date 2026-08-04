<script lang="ts">
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { goto } from '$app/navigation';

	let title = $state('');
	let description = $state('');
	let contentLanguage = $state('en');
	let ageRating = $state('all_ages');
	let tags = $state('');
	let error = $state('');
	let uploading = $state(false);
	let uploadProgress = $state('');

	async function handleUpload(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		uploading = true;
		uploadProgress = 'Creating upload session...';
		error = '';

		try {
			const session = await api.post<{ id: string; s3_prefix: string }>('/upload-sessions', { mode: 'manual' });
			const key = file.name;
			const presign = await api.post<{ url: string; key: string }>(`/upload-sessions/${session.id}/presign`, { number: 1, key });

			uploadProgress = 'Uploading file...';
			const uploadRes = await fetch(presign.url, {
				method: 'PUT',
				body: file,
				headers: { 'Content-Type': file.type || 'application/octet-stream' }
			});

			if (!uploadRes.ok) {
				throw new Error('Upload failed');
			}

			uploadProgress = 'Confirming upload...';
			await api.post(`/upload-sessions/${session.id}/parts`, {
				number: 1,
				key: presign.key,
				size: file.size,
				etag: uploadRes.headers.get('ETag') || ''
			});

			uploadProgress = 'Creating comic...';

			const tagList = tags.split(',').map(t => t.trim()).filter(Boolean);

			await api.post('/comics', {
				title,
				description,
				content_language: contentLanguage,
				cover_key: presign.key,
				file_key: presign.key,
				page_keys: [presign.key],
				file_size_bytes: file.size,
				age_rating: ageRating,
				tags: tagList,
				upload_session_id: session.id
			});

			await goto('/upload');
		} catch (err) {
			error = (err as Error).message;
		} finally {
			uploading = false;
		}
	}
</script>

<Card class="max-w-2xl mx-auto">
	<CardHeader>
		<CardTitle>Create New Comic</CardTitle>
	</CardHeader>
	<CardContent>
		<form class="space-y-4" onsubmit={(e) => e.preventDefault()}>
			<div class="space-y-2">
				<Label for="title">Title *</Label>
				<Input id="title" bind:value={title} placeholder="Comic title" required />
			</div>

			<div class="space-y-2">
				<Label for="description">Description</Label>
				<textarea id="description" bind:value={description} rows="3" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" placeholder="Describe your comic..."></textarea>
			</div>

			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-2">
					<Label for="lang">Content Language</Label>
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
			</div>

			<div class="space-y-2">
				<Label for="tags">Tags (comma-separated)</Label>
				<Input id="tags" bind:value={tags} placeholder="action, adventure, fantasy" />
			</div>

			<div class="space-y-2">
				<Label>Comic File *</Label>
				<Input type="file" accept=".zip,.cbz,.pdf,image/*" onchange={handleUpload} disabled={uploading} />
				{#if uploadProgress}
					<p class="text-sm text-muted-foreground">{uploadProgress}</p>
				{/if}
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Button type="submit" disabled={uploading || !title} onclick={() => {}}>
				{uploading ? 'Creating...' : 'Create Comic'}
			</Button>
		</form>
	</CardContent>
</Card>
