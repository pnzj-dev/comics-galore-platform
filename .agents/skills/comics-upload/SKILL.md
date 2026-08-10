---
name: comics-upload
description: Implement Comics Galore comic creation including presigned uploads, upload sessions, unified POST /comics payload, and pending_review. Use when building uploader flows, S3 presign, or archive ingestion.
---

# Comic upload & creation

## Unified model

1. Upload binaries via presigned URLs (browser to S3).
2. Collect **file keys**.
3. Build one form payload (metadata + keys + kinds).
4. `POST /comics` — single creation API.
5. Status is always `pending_review` until human (or later AI) approval.

## V1 vs later

- **V1** — manual form path with full-width 3-section layout:
  Row 1: metadata fields (`1fr`) + 3:4 cover dropzone (320px, Cloudflare `CloudflarePresignedUpload`)
  Row 2: preview image grid (Cloudflare, 2:3 aspect slots, min 2, max 10, "+ Add" button)
  Row 3: archive file grid (S3 presigned, square slots, min 1, max 10)
  Submit: standardized `submitting ? 'Publishing...' : 'Publish Comic'` button.
- **SOON** — archive + `comic.json` via libarchive.js in the browser, same payload and API.

## Upload sessions

- Use sessions for large/resumable uploads when implementing recovery.
- Confirm parts, then finalise by calling `POST /comics` (not a separate create semantics).

## Tab navigation (upload page)

Upload page uses **query-param tabs** for bookmarkable state:
```
/upload?tab=list     → My Comics grid
/upload?tab=manual   → ComicForm
/upload?tab=archive  → ArchiveForm
```
Tab param read server-side in `+page.server.ts` via `url.searchParams.get('tab')`. Switch tabs with `goto('/upload?tab=manual')` — no local `$state`.

**Server-side sessions**: Active upload sessions loaded in `+page.server.ts` alongside the comics list (via `Promise.allSettled`), not in `onMount`.

**Success redirect**: After creation, both ComicForm and ArchiveForm call `goto('/upload?tab=list')` to refresh the grid with the new comic.

## Form submission pattern

Cover and preview images use **Cloudflare** (`encore.upload.CloudflarePresignedUpload()`) with XHR progress tracking. Archives use **S3** (`CreateSession` → `PresignUpload` → upload via `fetch` → `ConfirmPart`).

Validation before submit: title required, cover required, at least one archive required. Errors shown as `<p class="text-sm text-destructive">`.

## Language

- Creation payload includes `content_language` (default `en`).
- Archive `comic.json` should map `language` / `content_language`.

## Asset kinds

- `cover` | `preview` | `page` | `original`
- Pages and originals on S3 in v1.
- Mis-tagged kind is a validation error.

## Never

- Stream full archives through Encore as the primary path.
- Publish without going through `pending_review`.

## References

- ADR `0010-presigned-uploads.md`, `0009-comic-metadata-json.md`, `docs/api.md`
