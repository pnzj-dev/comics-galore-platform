// Client-side archive parsing via libarchive.js (WASM worker).
//
// Used by the archive upload tab to locate and parse `comic.json` /
// `metadata.json` inside a .cbz/.zip/.rar/.7z archive, then auto-build the
// same payload the manual form produces (ADR 0009).

import { Archive } from 'libarchive.js';

let initialized = false;

function ensureInit() {
	if (initialized) return;
	Archive.init({
		workerUrl: '/libarchive/worker-bundle.js',
	});
	initialized = true;
}

/** Metadata parsed from comic.json / metadata.json inside an archive. */
export interface ComicMetadata {
	title?: string;
	author?: string;
	description?: string;
	synopsis?: string;
	series?: string;
	language?: string;
	content_language?: string;
	age_rating?: string;
	tags?: string[];
}

const METADATA_FILENAMES = ['comic.json', 'metadata.json'];

function isMetadataFilename(name: string): boolean {
	const base = name.toLowerCase().split('/').pop() ?? '';
	return METADATA_FILENAMES.includes(base);
}

/** Normalize the free-form metadata JSON into a consistent shape. */
export function normalizeMetadata(raw: Record<string, unknown>): ComicMetadata {
	const str = (v: unknown): string | undefined => (typeof v === 'string' ? v : undefined);
	const tags = (v: unknown): string[] | undefined =>
		Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : undefined;

	return {
		title: str(raw.title),
		author: str(raw.author),
		description: str(raw.description),
		synopsis: str(raw.synopsis),
		series: str(raw.series),
		language: str(raw.language),
		content_language: str(raw.content_language),
		age_rating: str(raw.age_rating),
		tags: tags(raw.tags),
	};
}

/**
 * Open an archive File, locate comic.json / metadata.json, and return the
 * parsed metadata. Returns `null` when no metadata file is present (the form
 * falls back to manual entry).
 */
export async function extractComicMetadata(file: File): Promise<ComicMetadata | null> {
	ensureInit();

	const archive = await Archive.open(file);
	try {
		const filesObject = await archive.getFilesObject();

		const find = (obj: unknown): CompressedFile | null => {
			if (!obj || typeof obj !== 'object') return null;
			for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
				if (value && typeof value === 'object') {
					if (isMetadataFilename(key) && typeof (value as { extract?: unknown }).extract === 'function') {
						return value as CompressedFile;
					}
					const found = find(value);
					if (found) return found;
				}
			}
			return null;
		};

		const metaFile = find(filesObject);
		if (!metaFile) return null;

		const extracted = (await metaFile.extract()) as File;
		const text = await extracted.text();
		const parsed = JSON.parse(text) as Record<string, unknown>;
		return normalizeMetadata(parsed);
	} catch (err) {
		// Invalid/missing JSON is not fatal — fall back to manual entry.
		console.warn('[archive] metadata parse failed:', err);
		return null;
	} finally {
		await archive.close();
	}
}

/** Minimal structural type for the libarchive CompressedFile we rely on. */
interface CompressedFile {
	extract(): Promise<File>;
}
