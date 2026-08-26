// Client-side comic archive builder via fflate (pure JS, no WASM worker).
//
// Builds a `.cbz` (zip) containing a `metadata.json` (all the form metadata)
// plus the ordered page images, so the manual creation path produces the same
// self-describing archive the archive-upload tab consumes.

import { zipSync, strToU8 } from 'fflate';

export interface ComicPageInput {
	file: File;
	name: string;
}

/**
 * Build a .cbz archive from metadata + page images. Pages are ordered by name
 * (natural numeric sort) and stored as `page-001.jpg`, `page-002.jpg`, …
 */
export async function buildComicArchive(
	metadata: Record<string, unknown>,
	pages: ComicPageInput[],
): Promise<File> {
	const metadataJSON = JSON.stringify(metadata, null, 2);

	const sorted = [...pages].sort((a, b) =>
		a.name.localeCompare(b.name, undefined, { numeric: true }),
	);

	const entries: Record<string, Uint8Array> = {
		'metadata.json': strToU8(metadataJSON),
	};
	for (let i = 0; i < sorted.length; i++) {
		const p = sorted[i];
		const pathname = `page-${String(i + 1).padStart(3, '0')}${extensionOf(p.file) || '.jpg'}`;
		entries[pathname] = new Uint8Array(await p.file.arrayBuffer());
	}

	// level 0 = store (comic pages are already-compressed images).
	const archiveData = zipSync(entries, { level: 0 });
	return new File([archiveData], 'comic.cbz', {
		type: 'application/vnd.comicbook+zip',
	});
}

function extensionOf(file: File): string {
	const m = /\.[a-z0-9]+$/i.exec(file.name);
	return m ? m[0].toLowerCase() : '';
}
