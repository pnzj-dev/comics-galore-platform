// Upload helpers for the manual/archive comic flows.
//
// Two modes, configured via AppSettings.upload_mode:
//   - "direct"  — the client uploads straight to storage via presigned URLs
//     (Cloudflare for cover/preview, S3 for archive/pages).
//   - "backend" — the client streams through the backend, which forwards to
//     the respective services (Cloudflare Images for cover/preview, S3 for
//     archive/pages).
import { encore } from '$lib/api/encore';

export type UploadMode = 'direct' | 'backend';

function xhrPut(url: string, blob: Blob, onProgress?: (loaded: number, total: number) => void): Promise<void> {
	return new Promise((resolve, reject) => {
		const xhr = new XMLHttpRequest();
		xhr.upload.addEventListener('progress', (e) => {
			if (e.lengthComputable) onProgress?.(e.loaded, e.total);
		});
		xhr.addEventListener('load', () =>
			xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error('Upload failed')),
		);
		xhr.addEventListener('error', () => reject(new Error('Upload failed')));
		xhr.open('PUT', url);
		xhr.send(blob);
	});
}

/** Wrap a percentage progress callback for xhrPut's byte-level callback. */
function asPct(onProgress?: (pct: number) => void): ((loaded: number, total: number) => void) | undefined {
	if (!onProgress) return undefined;
	return (loaded, total) => onProgress(Math.round((loaded / total) * 100));
}

function xhrPostJSON(url: string, body: XMLHttpRequestBodyInit, onProgress?: (pct: number) => void): Promise<string> {
	return new Promise((resolve, reject) => {
		const xhr = new XMLHttpRequest();
		xhr.upload.addEventListener('progress', (e) => {
			if (e.lengthComputable) onProgress?.(Math.round((e.loaded / e.total) * 100));
		});
		xhr.addEventListener('load', () => {
			if (xhr.status >= 200 && xhr.status < 300) {
				try {
					const data = JSON.parse(xhr.responseText);
					if (data?.key) resolve(data.key);
					else reject(new Error('Upload failed'));
				} catch {
					reject(new Error('Upload failed'));
				}
			} else {
				reject(new Error('Upload failed'));
			}
		});
		xhr.addEventListener('error', () => reject(new Error('Upload failed')));
		xhr.open('POST', url);
		xhr.send(body);
	});
}

/** Upload a cover/preview image. */
export async function uploadImage(
	file: File,
	onProgress: ((pct: number) => void) | undefined,
	uploadMode: UploadMode,
): Promise<string> {
	if (uploadMode === 'direct') {
		const res = await encore.upload.CloudflarePresignedUpload();
		await xhrPut(res.uploadURL, file, asPct(onProgress));
		return res.imageID;
	}
	return xhrPostJSON('/api/upload/image', file, onProgress);
}

/** Upload an archive or page file (preserving its filename as the key). */
export async function uploadFile(
	file: File,
	onProgress: ((pct: number) => void) | undefined,
	uploadMode: UploadMode,
	name?: string,
): Promise<string> {
	if (uploadMode === 'direct') {
		const session = await encore.upload.CreateSession({ mode: 'archive' });
		const presign = await encore.upload.PresignUpload(session.id, { number: 1, key: name || file.name });
		await xhrPut(presign.url, file, asPct(onProgress));
		await encore.upload.ConfirmPart(session.id, {
			number: 1,
			key: presign.key,
			size: file.size,
			etag: '',
		});
		return presign.key;
	}

	const formData = new FormData();
	formData.append('file', file, name || file.name);
	return xhrPostJSON('/api/upload/file', formData, onProgress);
}

/**
 * Upload an archive, splitting it into multiple presigned parts when it
 * exceeds the configured part size, then finalizing (server-side concat) into
 * a single object. Parts are uploaded concurrently (bounded by `concurrency`).
 * Small archives use the normal single-shot upload path.
 */
export async function uploadArchive(
	file: File,
	onProgress: ((pct: number) => void) | undefined,
	uploadMode: UploadMode,
	name?: string,
	partSizeMB?: number,
	concurrency?: number,
): Promise<string> {
	const maxBytes = (partSizeMB && partSizeMB > 0 ? partSizeMB : 0) * 1024 * 1024;
	if (!maxBytes || file.size <= maxBytes) {
		return uploadFile(file, onProgress, uploadMode, name);
	}

	const session = await encore.upload.CreateSession({ mode: 'archive' });
	const chunkCount = Math.ceil(file.size / maxBytes);

	const chunks: { index: number; blob: Blob }[] = [];
	for (let i = 0; i < chunkCount; i++) {
		const start = i * maxBytes;
		chunks.push({ index: i, blob: file.slice(start, Math.min(start + maxBytes, file.size)) });
	}

	const limit = Math.max(1, concurrency || 4);
	let next = 0;
	const chunkProgress = new Array<number>(chunkCount).fill(0);
	let totalUploaded = 0;

	const reportProgress = () => {
		onProgress?.(Math.min(99, Math.round((totalUploaded / file.size) * 100)));
	};

	const uploadChunk = async ({ index, blob }: { index: number; blob: Blob }): Promise<void> => {
		const partName = `${name || file.name}.part${String(index + 1).padStart(3, '0')}`;
		const presign = await encore.upload.PresignUpload(session.id, { number: index + 1, key: partName });
		await xhrPut(presign.url, blob, (loaded) => {
			const delta = loaded - chunkProgress[index];
			chunkProgress[index] = loaded;
			totalUploaded += delta;
			reportProgress();
		});
		await encore.upload.ConfirmPart(session.id, {
			number: index + 1,
			key: presign.key,
			size: blob.size,
			etag: '',
		});
	};

	const workers = Array.from({ length: Math.min(limit, chunkCount) }, async () => {
		while (next < chunkCount) {
			const chunk = chunks[next++];
			await uploadChunk(chunk);
		}
	});
	await Promise.all(workers);

	const final = await encore.upload.FinalizeMultipart(session.id, {
		output_name: name || file.name,
	});
	onProgress?.(100);
	return final.key;
}
