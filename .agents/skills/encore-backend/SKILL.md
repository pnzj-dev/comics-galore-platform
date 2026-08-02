---
name: encore-backend
description: Implement Comics Galore Encore Go backend including services, SQL migrations, auth roles, and generated clients. Use when writing or changing backend APIs, migrations, or background tasks.
---

# Encore Backend – Comics Galore

## Rules

- Scaffold and structure with official Encore CLI only.
- PostgreSQL changes only via Encore SQL migrations.
- Prefer SQLC for queries. No raw SQL outside migrations and SQLC.
- Business logic lives in Encore services, not in the Wails Go shell.

## Service map (domain)

auth, users, tiers/intervals/plans, comics, upload sessions, social, moderation, subscriptions, webhooks, settings, admin, storage/media.

## Non-negotiables

- Startup check — at least one `admin` user or process exits with a clear log.
- Comics created via API start as `pending_review`.
- Downloads enforce quota server-side.
- Webhooks verified, idempotent, raw payload stored.
- Payments go through `PaymentsProvider` (v1 NowPayments only).

## Uploads

- Never accept multi-GB comic bodies on the API.
- Presign S3 for `page` and `original`.
- Cover/preview may use Cloudflare Images later; v1 may store cover on S3 with `kind=cover`.

## Client

- Prefer the generated Encore TypeScript client for SvelteKit and desktop frontends.

## References

- `docs/api.md`, `docs/database.md`, `docs/architecture.md`, `docs/v1-scope.md`
