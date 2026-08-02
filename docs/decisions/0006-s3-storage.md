# ADR 0006 – Object Storage: S3-Compatible

## Status
Accepted

## Context
Comic archives and extracted covers are large binary objects that should not live in PostgreSQL.

## Decision
- Use any S3-compatible object storage (AWS S3, MinIO, Cloudflare R2, etc.).
- Original archives and generated covers are stored under deterministic keys.
- Downloads are served via short-lived pre-signed URLs.
- Upload can be direct-to-S3 (pre-signed) or through Encore, depending on final implementation simplicity.

## Consequences
- A thin `storage` service or package handles signed URL generation.
- Background jobs perform extraction after the object is safely stored.
- File size is recorded for quota accounting.
```
