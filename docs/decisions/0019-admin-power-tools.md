# ADR 0019 – Admin Power Tools

## Status
Proposed

## Context
The v1 admin panel covers core lists, moderation, settings, and KPI cards. The full vision (product.md "Admin Control Panel") adds impersonation, saved views, CSV export, background-job oversight, Staff Picks ordering, and storage usage. These extend existing services rather than introducing new domains.

## Decision

### Audited impersonation
- `POST /admin/users/:id/impersonate` (admin) issues a short-lived JWT carrying an `impersonated_as` claim (original admin id kept in `actor_id`).
- The auth handler accepts the claim and resolves `AuthData` to the impersonated user; every action is recorded in `audit_logs` with `actor_id` = admin and `impersonated_as` = target.
- No separate session table; revoke = logout.

### Saved datalist views
- `saved_views(id, admin_id, resource, name, filters JSONB, created_at)` in the relevant service (admin-only, per resource).
- `GET/POST/DELETE /admin/views`.

### CSV export
- `GET /admin/export/:resource` (users, comics, payments) — streams CSV from the owning service.

### Background jobs / dead-letter
- `job_runs(id, name, status pending|running|done|failed, attempts, error, payload JSONB, next_retry_at, created_at)` in the service that runs the job.
- `GET /admin/jobs`, `POST /admin/jobs/:id/retry`. Backed by Encore workers/pubsub.

### Staff Picks ordering
- `staff_picks(comic_id PK, position int, created_at)` in comics; admin reorder UI; public home reads this for the featured rail.

### Storage usage
- Extend existing `GET /admin/storage-stats` (upload service) with S3 bucket size approximation.

## Consequences
- No new database; changes land in existing services.
- Impersonation is explicit, audited, and revocable via the token lifecycle.
- Admin UX grows via the existing datalist (`AdminTable`) and drawer components.

## References
- ADR `0003-auth.md`, `0016-service-communication.md`, `docs/product.md` (Admin), `docs/ui.md` (Admin Control Panel)
