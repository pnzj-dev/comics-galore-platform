# ADR 0014 – Media Storage & Image Delivery

## Status
Accepted

## Context
Covers and previews need CDN/transforms; full comic pages are large and should stay on cheap object storage. Operators need flexibility in how URLs are delivered.

## Decision
1. **Storage by kind**
   - `cover` and `preview` → **Cloudflare Images only**
   - `page` and `original` (archives) → **S3-compatible storage only**
2. Every asset has an explicit **kind** flag (`cover` | `preview` | `page` | `original`).
3. **Serving mode** is an admin setting (default `direct`):
   - `direct` – use underlying storage URL
   - `imgproxy` – rewrite via configured imgproxy-compatible service
   - `cloudflare_images` – prefer Cloudflare Images delivery/variants
4. Application code resolves public image URLs through a single media helper so mode changes do not require content migration.

## Consequences
- Upload paths differ: CF Images API for cover/preview; S3 presigned sessions for pages/archives.
- Reader and grids never hard-code raw bucket URLs without going through the resolver when a proxy mode is enabled.
- Mis-tagging kind is a validation error at upload/finalise time.
