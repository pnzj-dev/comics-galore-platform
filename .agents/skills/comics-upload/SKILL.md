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

- **V1** — manual form path only (must work).
- **SOON** — archive + `comic.json` via libarchive.js in the browser, same payload and API.

## Upload sessions

- Use sessions for large/resumable uploads when implementing recovery.
- Confirm parts, then finalise by calling `POST /comics` (not a separate create semantics).

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
