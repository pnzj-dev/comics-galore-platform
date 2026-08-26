// Client-side archive page extraction via fflate (pure JS, no WASM worker).
//
// Extracts image pages from a .cbz/.zip archive in the browser so the uploader
// can upload each page individually and record its dimensions. Used by the
// archive upload tab. Non-zip archives (PDF/RAR/7z) return null so the
// server-side extraction worker handles them.

import { unzipSync } from 'fflate';

const IMAGE_EXTENSIONS = /\.(jpe?g|png|gif|webp|avif|bmp)$/i;

export interface ExtractedPage {
	file: File;
	name: string;
	width: number;
	height: number;
}

/** Determine whether a filename is an image page. */
export function isPageFilename(name: string): boolean {
	return IMAGE_EXTENSIONS.test(name);
}

/** Read an image File's pixel dimensions. */
export function imageDimensions(file: File): Promise<{ width: number; height: number }> {
	return new Promise((resolve, reject) => {
		const url = URL.createObjectURL(file);
		const img = new Image();
		img.onload = () => {
			URL.revokeObjectURL(url);
			resolve({ width: img.naturalWidth, height: img.naturalHeight });
		};
		img.onerror = () => {
			URL.revokeObjectURL(url);
			reject(new Error('could not read image dimensions'));
		};
		img.src = url;
	});
}

/**
 * Extract and sort image pages from a zip-family archive (.cbz/.zip).
 * Returns null when the archive is not a supported page-archive format
 * (e.g. PDF, RAR, 7z — those require server-side extraction).
 */
export async function extractArchivePages(file: File): Promise<ExtractedPage[] | null> {
	if (!/\.(cbz|zip)$/i.test(file.name)) {
		return null;
	}

	let files: Record<string, Uint8Array>;
	try {
		files = unzipSync(new Uint8Array(await file.arrayBuffer()));
	} catch {
		return null;
	}

	const imageEntries = Object.entries(files)
		.filter(([name]) => isPageFilename(name))
		// Natural page order by filename (handles 001.jpg … 010.jpg).
		.sort((a, b) => a[0].localeCompare(b[0], undefined, { numeric: true }));

	const pages: ExtractedPage[] = [];
	for (const [path, data] of imageEntries) {
		const name = path.split('/').pop() || path;
		const pageFile = new File([new Uint8Array(data)], name);
		try {
			const dims = await imageDimensions(pageFile);
			pages.push({ file: pageFile, name, width: dims.width, height: dims.height });
		} catch {
			// A non-image entry slipped through; skip it.
		}
	}
	return pages;
}
