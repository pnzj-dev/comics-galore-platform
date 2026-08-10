# API Surface – Comics Galore (key endpoints)

## Uploader – Upload Sessions (presigned)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/upload-sessions` | Create session (mode: manual \| archive) | Uploader |
| GET | `/upload-sessions/:id` | Resume / inspect session | Uploader |
| POST | `/upload-sessions/:id/presign` | Get presigned URL(s) or multipart part URLs | Uploader |
| POST | `/upload-sessions/:id/parts` | Confirm completed parts (returns / records file keys) | Uploader |
| DELETE | `/upload-sessions/:id` | Abort session | Uploader |
| GET | `/comics/metadata-schema` | Public JSON Schema for archive mode | No |

## Unified Comic Creation

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/comics` | **Single creation endpoint** used by both Manual and Archive paths | Uploader |
| GET | `/uploader/comics` | Paginated list of my comics (newest first) | Uploader |

### Creation payload (conceptual)

```json
{
  "title": "...",
  "description": "...",
  "series_id": "...",
  "series_order": 1,
  "tags": ["..."],
  "cover_key": "s3-key-from-presign",
  "file_key": "s3-key-of-main-archive-or-primary",
  "page_keys": ["..."],
  "file_size_bytes": 123456789,
  "min_tier_id": null,
  "scheduled_for": null,
  "upload_session_id": "..."
}
```

- Manual form: user fills fields; file inputs produce keys via presigned uploads; then `POST /comics`.
- Archive mode: libarchive.js + metadata JSON automatically build the **exact same payload**; then `POST /comics`.
- Backend creates the comic with `status = pending_review`.

## Comic Review

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/moderation/comics` | Pending comics queue | Moderator/Admin |
| POST | `/moderation/comics/:id/approve` | Publish | Moderator/Admin |
| POST | `/moderation/comics/:id/reject` | Reject with reason | Moderator/Admin |

## Comments API

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/comics/:id/comments` | List threaded comments | No |
| POST | `/comics/:id/comments` | Create comment (or reply) | Yes |
| DELETE | `/comments/:id` | Delete comment (owner or mod/admin); cascades to replies | Yes |
| GET (SSE) | `/comments-stream/:id` | SSE stream of new comments for a comic | No |

## Tiers & Plans

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/tiers` | List all tiers | No |
| GET | `/plans` | List all plans (tiers × intervals matrix) | No |
| GET | `/plans/ready` | Whether all plans have provider plan IDs | No |
| GET (admin) | `/admin/plans/matrix-status` | Admin check if plan matrix is complete | Admin |
| PATCH (admin) | `/admin/plans/:id` | Manually set a plan's provider_plan_id (refuses if already linked) | Admin |
| POST (admin) | `/admin/plans/:id/auto-link` | Auto-create NowPayments plan via API + store the ID | Admin |

The `auto-link` endpoint calls the NowPayments `POST /v1/subscriptions/plans` API to create a remote plan and saves the returned ID as `provider_plan_id`. The `PATCH` endpoint has a lock guard preventing re-linking of already-configured plans.

## Other domains
Auth, users, tiers, intervals, plans, subscriptions, webhooks, settings, social, comments, messaging, support, admin KPIs & datalists remain as previously specified.

## Notes

- Binary content never goes through the Encore backend.
- File keys obtained after successful presigned uploads are the only references the creation API needs.
- One creation API → one validation path → one `pending_review` pipeline.


## Admin – extended operations

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/admin/comics/bulk` | Bulk publish / reject / feature / delete | Admin |
| POST | `/admin/users/bulk` | Bulk role / ban / tier changes | Admin |
| GET | `/admin/recycle-bin` | Soft-deleted comics (and optionally users) | Admin |
| POST | `/admin/recycle-bin/:id/restore` | Restore | Admin |
| GET | `/admin/users/:id/detail` | Full user drawer data | Admin |
| POST | `/admin/users/:id/ban` | Ban / suspend with reason + duration | Admin |
| POST | `/admin/users/:id/impersonate` | Start audited impersonation session | Admin |
| POST | `/admin/subscriptions/grant` | Manual grant / extend / revoke | Admin |
| GET/POST | `/admin/coupons` | Simple promo codes CRUD | Admin |
| GET | `/admin/payments/past-due` | Failed / past-due list | Admin |
| GET | `/admin/ai/queue` | AI uncertain decisions queue | Admin |
| GET | `/admin/ai/decisions` | AI decision log | Admin |
| GET | `/admin/jobs` | Background job status | Admin |
| POST | `/admin/jobs/:id/retry` | Retry dead-letter / failed job | Admin |
| GET | `/admin/system/storage` | Storage usage approximation | Admin |
| POST | `/admin/broadcasts` | Send announcement (all or by tier) | Admin |
| GET | `/admin/export/:resource` | CSV export (users, comics, payments…) | Admin |


## Public & account – additions

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/tags` | List tags | No |
| GET | `/tags/:slug` | Tag page (comics) | Opt |
| POST | `/series/:id/follow` | Follow series | Yes |
| DELETE | `/series/:id/follow` | Unfollow series | Yes |
| GET | `/me/notification-preferences` | Get prefs | Yes |
| PATCH | `/me/notification-preferences` | Update prefs | Yes |
| POST | `/auth/verify-email` | Confirm email (token) | No |
| POST | `/auth/resend-verification` | Resend verification | Yes |
| POST | `/auth/password-reset/request` | Request reset | No |
| POST | `/auth/password-reset/confirm` | Confirm reset | No |
| GET | `/legal/terms` | Terms (or static) | No |
| GET | `/legal/privacy` | Privacy | No |
| GET | `/legal/dmca` | DMCA / copyright | No |

Comics payloads include `age_rating`. Public list/detail endpoints may filter or gate mature content according to settings / user preference.


## Media

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/media/resolve` | Resolve delivery URL(s) for asset id/key according to `image_serving_mode` | Opt |
| POST | `/media/cloudflare/upload-url` | Credential/url for cover or preview upload to Cloudflare Images | Uploader |
| (existing) | `/upload-sessions/.../presign` | S3 presign for **pages** and **original** archives only | Uploader |

Creation payload may reference:
- `cover` → Cloudflare Images id/url (`kind=cover`)
- `previews[]` → Cloudflare Images (`kind=preview`)
- `pages[]` / `file_key` → S3 keys (`kind=page` / `original`)

Admin settings include `image_serving_mode` and proxy base URL.


## Language & i18n

- Comic create/update payloads include required `content_language`.
- Public `GET /comics` supports filter `language=` (and facets).
- `PATCH /users/me` may update `ui_locale`.
- Admin settings expose `default_ui_locale`, `enabled_ui_locales`, `default_content_language`.
