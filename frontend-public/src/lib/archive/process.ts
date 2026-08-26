// Shared comic-archive creation pipeline.
//
// Both the Manual tab (which builds a .cbz) and the Archive tab (which accepts a
// dropped .cbz/.zip/…) converge here: upload the archive as the "original",
// extract page images client-side, upload each page to S3, then POST /comics.
// Progress is reported through an `onStep` callback so the UI can show a verbose
// step list.

import { encore } from '$lib/api/encore';
import { extractArchivePages } from './pages';
import { uploadFile, uploadArchive, type UploadMode } from './upload';
import { sanitizeFilename } from '$lib/utils/format';

export type StepStatus = 'active' | 'done' | 'error';

export interface ComicFormData {
	title: string;
	author: string;
	description: string;
	content_language: string;
	category: string;
	genre: string;
	age_rating: string;
	tags: string[];
	reading_direction: string;
	isbn: string;
	upc: string;
	issn: string;
	volume: string;
	issue_number: string;
	is_premium: boolean;
	min_tier_id: string;
}

export interface SeriesInput {
	series_id?: string;
	series_title?: string;
	series_genre?: string;
	series_category?: string;
	series_schedule_day?: string;
}

export interface ComicArchiveInput {
	archiveFile: File;
	coverKey: string;
	previewKeys: string[];
	form: ComicFormData;
	series?: SeriesInput;
	uploadMode: UploadMode;
	uploadPartSizeMB?: number;
	uploadConcurrency?: number;
	turnstileToken?: string;
	onStep?: (id: string, status: StepStatus, message?: string) => void;
}

export async function processComicArchive(input: ComicArchiveInput): Promise<void> {
	const {
		archiveFile,
		coverKey,
		previewKeys,
		form,
		series,
		uploadMode,
		uploadPartSizeMB,
		uploadConcurrency,
		turnstileToken,
		onStep,
	} = input;

	// The archive is always stored as .cbz: zip-family drops are renamed, and
	// CBR/RAR/PDF are re-packed server-side by the extraction worker. The object
	// key basename carries the display filename so presigned downloads get a
	// friendly name (browser derives it from the URL's last path segment).
	const isZipLike = /\.(cbz|zip)$/i.test(archiveFile.name);
	const niceBase = sanitizeFilename([form.author, form.title, form.volume, form.issue_number]);
	const origExt = (archiveFile.name.match(/\.[a-z0-9]+$/i)?.[0] || '').toLowerCase();
	const archiveKeyName = isZipLike ? `${niceBase}.cbz` : `${niceBase}${origExt}`;
	const archiveMimetype = isZipLike
		? 'application/vnd.comicbook+zip'
		: archiveFile.type || '';

	// 1. Upload the archive (original, downloadable) — split into parts when large.
	onStep?.('archive', 'active', 'Uploading archive…');
	const fileKey = await uploadArchive(archiveFile, undefined, uploadMode, archiveKeyName, uploadPartSizeMB, uploadConcurrency);
	onStep?.('archive', 'done');

	// 2. Extract page images + dimensions.
	onStep?.('pages', 'active', 'Extracting pages…');
	const pages = await extractArchivePages(archiveFile).catch(() => null);
	onStep?.('pages', 'done', pages ? `${pages.length} pages` : undefined);

	// 3. Upload each page to S3 (reader needs individual page URLs).
	const pageKeys: string[] = [];
	const pageDimensions: { width: number; height: number }[] = [];
	if (pages && pages.length > 0) {
		onStep?.('upload', 'active', 'Uploading pages…');
		for (const page of pages) {
			const key = await uploadFile(page.file, undefined, uploadMode);
			pageKeys.push(key);
			pageDimensions.push({ width: page.width, height: page.height });
		}
		onStep?.('upload', 'done');
	}

	// 4. Create the comic.
	onStep?.('publish', 'active', 'Publishing comic…');
	await encore.comics.CreateComic({
		title: form.title,
		author: form.author,
		description: form.description,
		content_language: form.content_language,
		category: form.category,
		genre: form.genre,
		cover_key: coverKey,
		file_key: fileKey,
		page_keys: [...previewKeys, ...pageKeys],
		page_dimensions: pageDimensions,
		reading_direction: form.reading_direction,
		isbn: form.isbn,
		upc: form.upc,
		issn: form.issn,
		volume: form.volume,
		issue_number: form.issue_number,
		archive_mimetype: archiveMimetype,
		file_size_bytes: archiveFile.size,
		min_tier_id: form.min_tier_id,
		age_rating: form.age_rating as never,
		is_premium: form.is_premium,
		tags: form.tags,
		upload_session_id: '',
		series_id: series?.series_id ?? '',
		series_title: series?.series_title ?? '',
		series_genre: series?.series_genre ?? '',
		series_category: series?.series_category ?? '',
		series_schedule_day: series?.series_schedule_day ?? '',
		turnstile_token: turnstileToken || '',
	});
	onStep?.('publish', 'done');
}
