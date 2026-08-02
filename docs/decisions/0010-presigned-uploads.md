# ADR 0010 – Presigned Uploads + Unified Creation Payload

## Status
Accepted

## Context
Comic archives can be large. Uploading them through the backend wastes bandwidth and CPU. Long uploads need crash recovery. We also want a single creation path for both manual and archive modes.

## Decision
- All binary uploads use **S3 presigned URLs** (multipart when appropriate). The browser never sends file bodies to the Encore API.
- After upload the frontend receives **file keys**.
- Those keys are placed into a **form payload**.
- The archive path (libarchive.js + metadata JSON) automatically builds the **identical form payload** that the manual form would produce.
- Both paths call the **same** `POST /comics` endpoint.
- An `upload_session` entity enables resume after crashes.
- Comics are always created with status `pending_review`.

## Consequences
- One creation API, one validation path, one review pipeline.
- Backend stays lean.
- Frontend owns progress, resume, and payload assembly.
