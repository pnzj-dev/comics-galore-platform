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

## Common Pitfalls & Fixes

### Cross-database foreign keys
Encore creates a **separate PostgreSQL database per service**. A `REFERENCES` constraint to a table in another service will fail. Use a plain UUID column and validate references at the application level.

### Path parameter case sensitivity
Function parameter names must match the path parameter **exactly** (case-sensitive).
```
Path: /reading/:comicId     ✓  func(ctx, comicId string)
Path: /reading/:comicId     ✗  func(ctx, comicID string)   // "comicID" != "comicId"
```

### Static path vs parameterized path conflicts
Under the same prefix, a static segment conflicts with a parameterized segment.
```
✗  /reading/:comicId   and   /reading/continue   // "continue" could be a value for :comicId
✓  /reading/:comicId   and   /reading-continue     // flat path avoids conflict
```

### API response types must be named structs
Encore requires **named structs** for all API response types. Raw maps or anonymous structs are rejected.
```
✗  func Foo(...) (map[string]X, error)
✓  type FooResponse struct { Items map[string]X `json:"items"` }
    func Foo(...) (*FooResponse, error)
```

### jsonb type casting in migrations
PostgreSQL does not implicitly cast text to `jsonb` in INSERT statements. Use `::jsonb`:
```sql
INSERT INTO plans (features) VALUES ('["item1","item2"]'::jsonb);
```

### Migration ordering for CHECK constraints
Always **UPDATE existing rows to conform** before adding a new CHECK constraint:
```sql
UPDATE plans SET interval = 'monthly' WHERE interval = 'lifetime';  -- 1. fix data
ALTER TABLE plans DROP CONSTRAINT IF EXISTS plans_interval_check;   -- 2. drop old
ALTER TABLE plans ADD CONSTRAINT ... CHECK (...);                    -- 3. add new
```

### Stale generated files (encore.gen.go)
After renaming/deleting types referenced by generated code, delete `encore.gen.go` files and let Encore regenerate them on next `encore run`.

## References

- `docs/api.md`, `docs/database.md`, `docs/architecture.md`, `docs/v1-scope.md`
