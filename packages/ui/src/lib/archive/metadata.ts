// Client-side archive parsing via fflate (pure JS, no WASM worker).
//
// Used by the archive upload tab to locate and parse `comic.json` /
// `metadata.json` inside a .cbz/.zip archive, then auto-build the same
// payload the manual form produces (ADR 0009).

import { unzipSync, strFromU8 } from 'fflate';

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
	volume?: string;
	issue_number?: string;
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
		volume: str(raw.volume),
		issue_number: str(raw.issue_number),
	};
}

/**
 * Open an archive File, locate comic.json / metadata.json, and return the
 * parsed metadata. Returns `null` when no metadata file is present (the form
 * falls back to manual entry).
 */
export async function extractComicMetadata(file: File): Promise<ComicMetadata | null> {
	try {
		const files = unzipSync(new Uint8Array(await file.arrayBuffer()));
		const metaName = Object.keys(files).find(isMetadataFilename);
		if (!metaName) return null;

		const text = strFromU8(files[metaName]);
		return normalizeMetadata(JSON.parse(text) as Record<string, unknown>);
	} catch (err) {
		// Invalid/missing JSON or non-zip archive is not fatal — fall back to
		// manual entry.
		console.warn('[archive] metadata parse failed:', err);
		return null;
	}
}
