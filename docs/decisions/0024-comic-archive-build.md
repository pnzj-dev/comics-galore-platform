# ADR 0024 – Comic Archive Build (Self-Describing `.cbz`)

## Status
Accepted

## Context
The Manual creation tab uploaded assets piecemeal — cover (Cloudflare), previews
(Cloudflare), and page/archive files (S3 presigned) — while the Archive tab
uploaded a single `.cbz` and extracted `comic.json`/pages client-side. Two
divergent pipelines, and the Archive tab had no proper cover (it reused the
archive blob).

## Decision

1. **Manual creation builds a self-describing archive.** On submit the client
   uses libarchive.js `Archive.write` to produce a `.cbz` containing
   `metadata.json` (all form metadata) plus the ordered page images. This is the
   "original" archive stored on S3 (`file_key`, downloadable).

2. **One shared pipeline.** Both tabs converge on a single client-side pipeline
   (`processComicArchive`): upload the archive → extract pages (libarchive.js) →
   upload each page to S3 (`page_keys` + `page_dimensions`) → `POST /comics`.
   Each step reports status to a verbose progress list in the form.

3. **Storage split is explicit.** `cover` and `preview` go to Cloudflare Images;
   `page` and `original` (the `.cbz`) go to S3. `page_keys` mixes Cloudflare
   preview ids (front, for the gallery) and S3 page keys (reader), resolved by
   `resolveCoverURL`/`resolvePageURLs`.

4. **The Archive tab is aligned.** It drops a `.cbz`, extracts `metadata.json`
   to prefill the form, requires a separate Cloudflare cover + previews, and runs
   the same shared pipeline (fixing the archive-as-cover bug).

## Consequences
- Removes the per-file presign complexity of the old Manual tab (page images
  are packed into one archive instead of N separate S3 uploads).
- Reader still needs individual page URLs, so pages are extracted and uploaded
  individually by the shared pipeline — but the *source of truth* is now one
  archive, identical across tabs.
- No new dependency: libarchive.js (already used for reading) also writes ZIP.

## References
- ADR `0009-comic-metadata-json.md`, ADR `0023-series-association.md`,
  `docs/architecture.md` (Comic archive build), `docs/product.md` (Tab 2).
