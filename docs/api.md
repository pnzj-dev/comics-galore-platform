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

### Creation payload (`POST /comics`)

```json
{
  "title": "...",
  "author": "...",
  "description": "...",
  "content_language": "en",
  "category": "Manga",
  "genre": "Action",
  "cover_key": "cf-image-id-or-s3-key",
  "file_key": "s3-key-of-original-archive",
  "page_keys": ["preview-cf-id", "...", "page-s3-key"],
  "page_dimensions": [{ "width": 800, "height": 1200 }],
  "reading_direction": "ltr",
  "archive_mimetype": "application/vnd.comicbook+zip",
  "isbn": "", "upc": "", "issn": "",
  "volume": "", "issue_number": "",
  "file_size_bytes": 123456789,
  "min_tier_id": "",
  "age_rating": "all_ages",
  "is_premium": false,
  "tags": ["..."],
  "upload_session_id": "",
  "series_id": "existing-series-uuid",
  "series_title": "",
  "series_genre": "", "series_category": "", "series_schedule_day": ""
}
```

- **Series association** (ADR `0023-series-association.md`): pass `series_id` to attach to an existing series, or `series_title` (+ optional `series_genre`/`series_category`/`series_schedule_day`) to create a new series inline. `series_order` is auto-incremented.
- **Manual form** builds a self-describing `.cbz` (see ADR `0024-comic-archive-build.md`); both Manual and Archive paths then converge on the same extract → upload → publish pipeline.
- Backend creates the comic with `status = pending_review`.

### Series discovery & filter

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/series` | List all series | No |
| GET | `/series-search` | Paginated + filtered series (`search`, `category`, `page`, `limit`) — DB-side filtering | No |
| GET | `/series-categories` | Distinct category values (facet list) | No |
| POST | `/series` | Create a series (title, description, genre, category, schedule_day, cover_key) | Uploader |
| GET | `/series/:id` | Series detail | No |
| GET | `/series/:id/comics` | Series comics (paginated) | No |
| GET | `/admin/series` | Admin series datalist (search/sort/filter/paginate) | Admin |

`/series-search` and `/series-categories` power the public "Browse Series" page and the upload form's series picker (search + category + "load more" — all server-side).

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
| POST | `/comments/:id/flag` | Flag a comment for review (idempotent per user) | Yes |
| GET | `/moderation/flags` | List open comment flags (moderation queue) | Moderator/Admin |
| POST | `/moderation/flags/:id/resolve` | Resolve a comment flag | Moderator/Admin |
| GET (SSE) | `/comments-stream/:id` | SSE stream of new comments for a comic | No |
| GET | `/comics-language-facets` | Language facet counts for published comics | No |
| GET | `/tags` | Popular tag counts for published comics (limit 20) | No |
| GET | `/favorites` | Paginated list of the current user's favorited comics | Yes |

> **Mature content policy** — when the admin setting `forbid_mature_for_free` is enabled, free-tier and anonymous callers never see mature/explicit comics: all list endpoints filter them out, `GetComic` returns `mature_locked=true` with `page_urls` withheld (cover kept for a blurred teaser), and `RecordDownload` rejects the download. Staff (admin/moderator) and paid tiers are exempt. The policy is exposed to other services via the private `GET /auth/content-policy` endpoint.

> **Search & filter** — `GET /comics` accepts `search` + `search_field` (`title` | `description` | `author`, empty = all fields) for full-text filtering, and `tag` for exact tag match. Combined with `language`, `sort`, `exclude_mature`, `page`, `limit`.

> **Pagination convention** — every list endpoint that returns a grid uses `page` + `limit` query params and returns a `total` count. `limit` defaults to 20 and is capped at 50. Public grids (`/comics`, `/series/:id/comics`, `/tags/:tag`, `/reading-lists/:id`) and authenticated grids (`/favorites`) all follow this convention. The frontend renders them via the shared `Pagination.svelte` component.

## Tiers & Plans

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/tiers` | List all tiers | No |
| GET | `/plans` | List all plans (tiers × intervals matrix) | No |
| GET | `/plans/ready` | Whether all plans have provider plan IDs | No |
| GET (admin) | `/admin/plans/matrix-status` | Admin check if plan matrix is complete | Admin |
| PATCH (admin) | `/admin/plans/:id` | Manually set a plan's provider_plan_id (refuses if already linked) | Admin |
| POST (admin) | `/admin/plans/link/:id` | Auto-create NowPayments plan via API + store the ID | Admin |
| POST (admin) | `/admin/plans/unlink/:id` | Unlink a single plan (clear provider_plan_id) | Admin |
| POST (admin) | `/admin/plans/unlink` | Unlink all plans (clear every provider_plan_id) | Admin |

The auto-link endpoint (in the `tiers` service) calls the NowPayments `POST /v1/subscriptions/plans` API through the shared `nowpayments` package to create a remote plan and saves the returned ID as `provider_plan_id`. The `PATCH` endpoint has a lock guard preventing re-linking of already-configured plans. Note: NowPayments Recurring Payments and Custody/sub-partner responses are wrapped in a `result` object (`{"result": {...}}`); the `nowpayments` client unwraps `result` before parsing. The standard endpoints (`/auth`, `/estimate`) return top-level fields and are parsed as-is.

The `nowpayments` package exposes a client for the **core** NowPayments endpoints (auth/status, currencies, payments/invoices, master balance, customer/sub-partner management, and recurring payments + plans), with the full official spec saved at `backend/billing/nowpayments-openapi.yaml` for reference. **Known NowPayments docs bug:** `POST /v1/subscriptions` ("Create recurring payments" / "Create an email subscription") returns `result` as an **array** in practice, although the docs show a single object — `CreateSubscription` parses both shapes.

Webhook callback URLs (`ipn_callback_url`) are built on the backend from `encore.Meta().APIBaseURL` (the running app's own base URL) rather than a client-supplied host header; a valid `NgrokURL` secret is only used to override the callback URL when running locally.

The checkout flow fetches the merchant's accepted coins from `GET /billing/currencies` (public) — backed by NowPayments `GET /v1/merchant/coins` (the coins enabled in "coins settings"). `GET /billing/check-balance` returns `{ balances: { "<currency>": { amount, pending_amount } } }` and `POST /billing/estimate-price` returns `{ estimated_amount, from_currency, to_currency }`.

## Bot protection (Cloudflare Turnstile)

User-initiated write endpoints accept an optional `turnstile_token` and verify it server-side (via the private `turnstile` service → `POST /turnstile/verify`) before running their existing logic: `auth.Register`, `auth.Login`, `auth.RequestPasswordReset`, `comics.CreateComic`, `comics.CreateComment`, `social.CreateTicket`, `social.ReplyTicket`. The SvelteKit frontend renders the widget when `VITE_TURNSTILE_SITEKEY` is set and passes the token through. Verification is **inert** when `TURNSTILE_SECRET` is unset (dev), and otherwise fails closed on `success`, `action`, and `hostname ∈ TURNSTILE_HOSTNAMES`.

## Other domains
Auth, users, tiers, intervals, plans, subscriptions, webhooks, settings, social, comments, messaging, support, admin KPIs & datalists remain as previously specified.

## Internal (service-to-service)

Private endpoints callable only from other Encore services — never exposed publicly. Services communicate via these typed calls rather than reading each other's tables (see ADR `0016-service-communication.md`).

| Method | Path | Description | Caller |
|--------|------|-------------|--------|
| POST | `/auth/ensure-sub-partner-id` | Get-or-create the user's NowPayments `sub_partner_id` | billing |
| POST | `/auth/set-user-tier` | Set a user's tier (after subscription webhook activation) | billing |
| GET | `/internal/plans/:id` | Plan details (provider plan id, interval, tier name, price) | billing |
| POST | `/auth/notify-followers-new-comic` | Email followers of a newly published comic (respects prefs) | comics |

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
| POST | `/uploaders/:id/follow` | Follow an uploader | Yes |
| DELETE | `/uploaders/:id/follow` | Unfollow an uploader | Yes |
| GET | `/uploaders/:id/follow-status` | Whether current user follows an uploader | Yes |
| POST | `/reading-progress-batch` | Batch reading progress for a set of comic IDs | Yes |
| GET | `/me/notification-preferences` | Get prefs | Yes |
| PATCH | `/me/notification-preferences` | Update prefs | Yes |
| POST | `/auth/verify-email` | Confirm email (token) | No |
| POST | `/auth/resend-verification` | Resend verification | Yes |
| GET | `/auth/username-available` | Check a username's availability + format validity | No |
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
| GET | `/media/*key` | Serve a stored image/object (Cloudflare or S3, by mode) | No |
| GET | `/download/*key` | Archive download: stream (small) or presigned redirect (large) | User |

Creation payload may reference:
- `cover` → Cloudflare Images id/url (`kind=cover`)
- `previews[]` → Cloudflare Images (`kind=preview`)
- `pages[]` / `file_key` → S3 keys (`kind=page` / `original`)

Admin settings include `image_serving_mode` and proxy base URL.

### Upload (backend-streamed mode)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/upload/image` | Stream a cover/preview to Cloudflare Images (S3 fallback); returns `{ key }` | Uploader |
| POST | `/upload/file` | Stream an archive/page to S3 via multipart; returns `{ key }` | Uploader |
| POST | `/upload-sessions/:id/finalize` | Concatenate a session's split parts into a single object; returns `{ key }` | Uploader |


## Background jobs (observability)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/admin/jobs` | List recent job runs (filter by `job_name`, `status`) | Admin |
| POST | `/admin/jobs/ai-moderate-comics/run` | Manually trigger the AI moderation sweep | Admin |
| POST | `/admin/jobs/waiting-pay-expiry/run` | Manually trigger the WAITING_PAY expiry sweep | Admin |

Job recording (service-to-service, private): `POST /jobs/record-start` → `{ id }`, `POST /jobs/record-finish` (`id`, optional `error`). Cron jobs (`ai-moderate-comics`, `waiting-pay-expiry`) and pubsub handlers (`archive-extract`, `ai-moderation`) record runs; pubsub subscriptions use `RetryPolicy{MaxRetries: 5}`.

## Storage

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/admin/storage` | Enumerate the object bucket: total bytes + object count, per-prefix breakdown, Cloudflare Images count | Admin |

## Announcements

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/announcements` | Broadcasts for the caller (`target=all` + their tier) | No |
| POST | `/admin/broadcasts` | Create a broadcast (title, body, target, tier) | Admin |
| GET | `/admin/broadcasts` | List broadcasts | Admin |

## Dashboard (admin)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/admin/dashboard` | Aggregated KPI stats (users/comics/billing/reading/storage + trends) | Admin |
| GET | `/admin/dashboard-stream` | SSE stream pushing the aggregate every 15s | Admin |

The `dashboard` service aggregates the per-service stats endpoints, which are now service-internal (`private`): `auth.AdminDashboardStats`, `auth.GetSignupTrend`, `comics.GetComicsStats`, `billing.GetBillingStats`, `reading.GetReadingStats`, `reading.GetDownloadTrend`, `upload.GetStorageStats`.

## Language & i18n

- Comic create/update payloads include required `content_language`.
- Public `GET /comics` supports filter `language=` (and facets).
- `PATCH /users/me` may update `ui_locale`.
- Admin settings expose `default_ui_locale`, `enabled_ui_locales`, `default_content_language`.

## Planned LATER endpoints (not yet implemented)

Summarised from ADRs 0017–0021. All auth-required unless noted.

## Messaging & Support (`social`) — ADR 0017 (implemented)

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/messages/conversations` | List my conversations | Yes |
| POST | `/messages/start/:userId` | Get-or-create conversation | Yes |
| GET | `/messages/conversation/:id` | List messages (marks read) | Yes |
| POST | `/messages/conversation/:id` | Send message | Yes |
| POST | `/messages/conversation/:id/read` | Mark read | Yes |
| GET (SSE) | `/messages-stream/:userId` | Live new-message stream | Yes |
| POST | `/support/tickets` | Create ticket | Yes |
| GET | `/support/tickets` | My tickets | Yes |
| GET | `/support/tickets/:id` | Ticket thread | Yes |
| POST | `/support/tickets/:id/reply` | Reply (user or staff) | Yes |
| GET | `/admin/support/tickets` | All tickets | Moderator/Admin |
| POST | `/admin/support/tickets/:id/assign` | Assign | Admin |
| POST | `/admin/support/tickets/:id/resolve` | Resolve | Moderator/Admin |
| POST | `/admin/broadcasts` | Send announcement | Admin |

## AI Moderation — ADR 0018 (implemented)
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | `/admin/ai/queue` | Uncertain AI decisions | Moderator/Admin |
| POST | `/admin/ai/queue/:id/resolve` | Resolve queued item | Moderator/Admin |
| GET | `/admin/ai/decisions` | AI decision log | Admin |

## Admin Power Tools — ADR 0019 (implemented)
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | `/admin/users/:id/impersonate` | Start audited impersonation | Admin |
| GET/POST/DELETE | `/admin/views` | Saved datalist views | Admin |
| GET (raw) | `/admin/export/users` | CSV export of users | Admin |
| GET | `/staff-picks` | Featured comics rail | No |
| POST | `/admin/staff-picks` | Add a staff pick | Admin |
| DELETE | `/admin/staff-picks/:comicId` | Remove a staff pick | Admin |

## Billing Growth — ADR 0020 (implemented)
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET/POST | `/admin/coupons` | Coupon CRUD | Admin |
| POST | `/admin/subscriptions/:id/grant` | Manual grant/extend | Admin |
| POST | `/admin/subscriptions/:id/revoke` | Revoke subscription | Admin |
| GET | `/admin/payments/past-due` | Failed/past-due list | Admin |

## Social Engagement — ADR 0021 (implemented)
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST/GET | `/reading-lists` | Create/list my shelves (`GET` accepts `comic_id` and returns `has_comic` + `comic_count`) | Yes |
| PATCH | `/reading-lists/:id` | Rename + toggle `is_public` | Yes |
| DELETE | `/reading-lists/:id` | Delete shelf (cascades items) | Yes |
| GET | `/reading-lists/:id` | Public shelf (if `is_public`) | Opt |
| GET | `/reading-lists/:id/mine` | Owner view of a shelf (works for private shelves) | Yes |
| POST | `/reading-lists/:id/items` | Add comic to shelf | Yes |
| DELETE | `/reading-lists/:id/items/:comicId` | Remove from shelf | Yes |
| GET | `/comics/:id/related` | "People also liked" | No |
