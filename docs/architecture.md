# Architecture Overview – Comics Galore

## Goals

- One backend (Encore).
- Two first-class clients: **Web** (SvelteKit) and **Desktop** (Wails).
- Maximum reuse of Svelte UI, forms, validation and client logic.

## High-Level Diagram

```
┌──────────────────┐     ┌──────────────────┐
│  SvelteKit Web   │     │  Wails Desktop   │
│  (frontend/)     │     │  (desktop/)      │
└────────┬─────────┘     └────────┬─────────┘
         │                        │
         │   both import from     │
         └──────────┬─────────────┘
                    │
         ┌──────────▼──────────┐
         │   packages/ui       │  Shared Svelte components,
         │   (shadcn-svelte,   │  stores, Zod schemas,
         │    forms, reader,   │  upload-session logic,
         │    comic cards…)    │  API client helpers
         └──────────┬──────────┘
                    │
                    │  HTTP / generated client
                    ▼
         ┌─────────────────────┐
         │   Encore Backend    │
         │   (backend/)        │
         └──────────┬──────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   PostgreSQL      S3       NowPayments
                   LLM       Resend
```

## Repository Layout

```
Comics-Galore/
├── backend/                     # Encore Go API (single source of truth)
│   ├── nowpayments/             # Shared non-service Go package:
│   │                            #   NowPayments client, types,
│   │                            #   PaymentsProvider interface, BuildCallbackURL
│   ├── auth/                    # users, roles, sub_partner_id, sessions, email
│   ├── billing/                 # subscriptions, deposits, webhooks, coupons
│   ├── tiers/                   # tiers × plans matrix, plan auto-link
│   ├── comics/                  # comics, moderation, social, AI decisions
│   ├── reading/                 # progress, downloads
│   ├── upload/                  # presigned uploads, media
│   ├── fixtures/                # test fixtures
│   ├── social/                  # messaging, support tickets, broadcasts
│   └── aiprovider/              # shared OpenAI-compatible client package
├── frontend-public/             # SvelteKit web app — comics-galore.com
├── frontend-admin/              # SvelteKit web app — admin.comics-galore.com
├── desktop/                     # Wails v2 application
│   ├── main.go / app.go         # Go side (bindings, native menus, etc.)
│   └── frontend/                # Vite + Svelte app that uses packages/ui
└── packages/
    └── ui/                      # Shared Svelte code
        ├── components/
        ├── forms/
        ├── lib/                 # Zod, stores, upload-session, API helpers
        └── ...
```

## Shared Layer (`packages/ui`)

Must contain (at minimum):

- shadcn-svelte based primitives and domain components (ComicCard, HeroComicCard, Reader, PlanModal, etc.)
- Superforms + Zod schemas for comic creation, login, settings…
- Upload Session client logic (presign → upload → collect file keys → build payload)
- Comic creation payload builder (used by both Manual and Archive tabs)
- Reader chrome (keyboard + swipe handlers)
- Theme / dark-mode stores
- Generated Encore TypeScript client wrappers (`$lib/api/encore.ts` → `encore-client.ts`)

Web and desktop both import from this package.  
Duplication of UI or validation logic is considered a bug.

## Web Client — Two Applications

### Public (`frontend-public/` — comics-galore.com)

- Full SSR / progressive enhancement for public pages.
- Route groups: `(site)`, `(app)`, `(auth)`. The `(site)` group holds the shared nav + `<main>` + footer (public + app pages); `(app)` adds the auth guard; `(auth)` renders a minimal (no nav/footer) shell for the standalone `/login` page.
- Uses `packages/ui` for almost all visual and form code.
- Handles home, browse, detail, reader, series, tags, pricing, upload, settings, auth, SEO, RSS, Open Graph, legal pages, age gate.
- **Server-side route guards**: `+layout.server.ts` and `+page.server.ts` load functions read the JWT from a cookie and `throw redirect(302)` before the page component renders. The `(app)` group requires authentication; the `upload` page requires `uploader`/`admin` role. See the cookie-based auth section below.

### Cookie-based Auth

The auth system uses an **HttpOnly session cookie** (`token`, kept for
compatibility) holding an opaque session id. The session id is validated by the
Encore auth handler against the `sessions` table (expiry + revocation). Because
the cookie is HttpOnly, browser JS cannot read it:

| Consumer | How it authenticates |
|----------|----------------------|
| SvelteKit server (SSR) | `resolveUser(cookies)` in `+layout.server.ts` → calls Encore `/auth/me` with the session as `Authorization: Bearer` |
| Browser API client | Routes through the same-origin `/api/[...path]` proxy, which reads the HttpOnly cookie and forwards the session to Encore |
| Auth store (`login`/`register`/`logout`) | POSTs to SvelteKit server endpoints (`/auth/login`, `/auth/register`, `/auth/logout`) which set/clear the HttpOnly cookie |

The cookie is set with `path=/`, `HttpOnly`, `SameSite=Lax`, and `max-age=2592000`
(30 days); `Secure` is added in production. Server-side guards validate the
session by asking Encore `/auth/me` rather than decoding a JWT payload. See
`docs/authentication.md` and ADR `0022-auth-methods.md`.

### Auth UX (modals + minimal login page)

- **Login / Register / Forgot-password** are global **modals** (`LoginModal`,
  `RegisterModal`, `ForgotPasswordModal`) mounted in the root layout and opened
  via the shared `modal` store (single-active, non-stacking). Nav "Login" /
  "Register" buttons (and home/pricing CTAs) open them instead of navigating.
- A standalone **`/login` page** still exists (minimal shell, no nav/footer) as
  the redirect target for auth-gated routes (`(app)` guard → `/login`).
- Registration requires a **username** handle — `3–20` lowercase alphanumerics
  with single `_`/`-` inside — validated live against `GET /auth/username-available`.
- Password fields use a visibility toggle; auth cards show the Comics Galore brand
  header.

### Data Loading — SvelteKit `load` Pattern

**Required for all pages that fetch external data.** Uses `+page.server.ts` load functions so data arrives at render time — no skeletons, SSR-friendly, search-engine readable.

```ts
// +page.server.ts — using the generated Encore client
import { getEncoreClient } from '$lib/server/encore';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
    const client = getEncoreClient(cookies.get('token'));
    const res = await client.comics.ListComics({ Page: 1, Limit: 4, Sort: 'newest' });
    return { comics: res.comics || [] };
};
```

```svelte
<!-- +page.svelte -->
<script lang="ts">
    let { data } = $props();
    // data.comics is available immediately — no onMount, no loading state, no skeleton
</script>
```

**Current state (v1):** All data-dependent pages use `+page.server.ts` load functions with the generated Encore TypeScript client (`getEncoreClient(token)`). The migration from `onMount` + client-side `api.get/post` is complete.

**For auth-required pages:** the `load` function reads the JWT cookie (same pattern as route guards) and passes it as `Authorization: Bearer <token>` to the Encore API.

**Generated client:** The Encore TypeScript client (`$lib/api/encore-client.ts`) is generated via `encore gen client --lang typescript` and wrapped in `$lib/api/encore.ts` (browser) and `$lib/server/encore.ts` (server). All page load functions and client-side API calls use this typed client — no raw `fetch()` or manual URL building remains. See `.agents/skills/sveltekit-encore-client/SKILL.md` for setup and usage.

### Admin (`frontend-admin/` — admin.comics-galore.com)

- Separate SvelteKit application for admin-only features.
- Fixed sidebar layout (`w-60`, `bg-slate-900`) with active page highlighting.
- Login required (no register — admins are created by the system).
- Uses `packages/ui` for shared components.
- Handles dashboard, moderation (pending comics queue), user management, subscriptions, comics management, recycle bin, settings.
- **Role-scoped access**: `admin` sees the full sidebar; `moderator` logs in but is restricted to the `/moderation` page only (server-side guard in `+layout.server.ts` redirects any other route back to `/moderation`).
- Settings page has Form/JSON toggle for raw editing of the `app_settings` JSON blob.
- Shows red plan matrix banner when incomplete (NowPayments link wizard resolves this) — banner/wizard are admin-only.

## Desktop Client (Wails)

- Wails Go shell + embedded webview.
- Frontend is a Vite + Svelte (or SvelteKit static) app that also consumes `packages/ui`.
- Reuses the same pages/flows for:
  - Browsing, reader, library
  - New Comic 3-tab workspace
  - Messaging, support, admin (if the user has the role)
- Desktop-specific additions (only in `desktop/`):
  - Native window & menu bar
  - Native file open dialogs (still upload via presigned URLs)
  - System tray / background behaviour (optional)
  - Deeper offline cache later (future)
- Auth tokens and API calls still go to the Encore backend (same as web).

## Backend (Encore) – unchanged responsibilities

- Auth, users, roles, bootstrap admin check
- Comics, upload sessions, presigned URLs, final `POST /comics`
- Pending review + AI / human moderation
- Tiers, intervals, plan matrix, NowPayments (wallets, deposits, subscriptions, webhooks)
- Social, messaging, support, settings, admin KPIs
- All business rules stay on the server

### Service communication rules

Each Encore service owns its own PostgreSQL database. Services **never** read or write another service's tables directly. Cross-service communication happens only through:

- **Typed private API calls** (`//encore:api private`) — e.g. `auth.EnsureSubPartnerID`, `auth.SetUserTier`, `tiers.GetPlan`.
- **Pub/Sub events** for asynchronous work.

Shared Go libraries that are not services (e.g. `backend/nowpayments/`) may be imported freely by multiple services. See ADR `0016-service-communication.md`.

### `sub_partner_id` lifecycle

- Stored on `users`, **owned by the `auth` service**.
- Provisioned **eagerly** on email verification (`VerifyEmail` → `ensureSubPartnerID`, synchronous, non-fatal on failure).
- Provisioned **lazily** via `auth.EnsureSubPartnerID` (called by `billing`) at subscription/deposit/balance-check time for users who never verified email.
- Saved atomically (`UPDATE ... WHERE sub_partner_id IS NULL`) and enforced unique via a partial index.

## Planned LATER architecture (not yet implemented)

The full vision adds these domains. Detailed specs are in the ADRs; this section summarises the service/table/endpoint shape so implementation can start from here.

### Messaging & Support (`social` service) — ADR 0017
- **Implemented.** `socialdb` with `conversations`, `messages`, `support_tickets`, `support_messages`, `broadcasts`.
- Endpoints: messaging (`/messages/*` + SSE), support (`/support/*`, admin `/admin/support/*`), broadcasts (`/admin/broadcasts`).
- Email fan-out via `auth.NotifySupportReply` (private), honouring preferences.

### AI Moderation (shared `aiprovider` package) — ADR 0018
- **Implemented.** OpenAI-compatible client package; config in `AppSettings` (`ai_moderation_enabled`, `ai_model`, `ai_endpoint`, `ai_prompt`, thresholds) + `AIModeratorAPIKey` secret.
- Encore Pub/Sub topic `ai-moderation` (published on comic/comment creation) + subscription worker that writes `ai_decisions`/`ai_review_queue`, auto-approve/reject by threshold, else queue for human.

### Admin Power Tools — ADR 0019
- **Implemented (partial).** Audited impersonation (scoped `impersonated_by` JWT claim), `saved_views`, `staff_picks`, CSV export (`/admin/export/users`). `job_runs` dead-letter dashboard deferred (needs worker instrumentation).

### Billing Growth — ADR 0020
- **Implemented.** `coupons` table + admin CRUD; manual grant/revoke; past-due list. Force-sync and `revenue_by_tier_interval` deferred; second provider = new adapter behind `PaymentsProvider` (interface already in place).

### Social Engagement — ADR 0021
- **Implemented.** `reading_lists` / `reading_list_items` (public shelves) in comics; `GET /comics/:id/related` (co-occurrence). Full shelf management shipped: rename / toggle-public / delete (`PATCH`/`DELETE /reading-lists/:id`), owner view of private shelves (`GET /reading-lists/:id/mine`), and an "Add to list" modal on the comic detail page.

### Series association — ADR 0023
- **Implemented.** Comics attach to a series at creation time (`series_id` or inline `series_title`); `series_order` auto-increments; series carry `cover_key`/`genre`/`category`/`schedule_day`/engagement aggregates. Series are discoverable/filterable via `GET /series-search` + `GET /series-categories`, and manageable in admin via `GET /admin/series`.

### Comic archive build — ADR 0024
- **Implemented.** Manual creation builds a self-describing `.cbz` (`metadata.json` + page images) on the client, then both Manual and Archive tabs converge on one shared extract → upload → publish pipeline with verbose step reporting. Storage split: cover/preview on Cloudflare Images, pages/original on S3.

## Upload & Creation Flow (identical on web and desktop)

1. Choose a tab: **Manual** (metadata + cover + previews + page images) or **Archive** (drop a `.cbz`/`.cbr`/…).
2. **Manual** builds a `.cbz` containing `metadata.json` + the page images client-side with **fflate** (ADR `0024`); the archive filename is derived from `author - title - volume - issue` (`#` → `no-`).
3. The shared pipeline uploads the archive (original), extracts pages client-side with fflate, uploads each page to S3, and POSTs `/comics` — with verbose per-step progress.
   - `upload_mode` (`direct` | `backend`, default `backend`) chooses presigned-URL uploads vs backend-streamed uploads.
   - Archives larger than `upload_part_size_mb` are split into parts and uploaded with bounded concurrency (`upload_concurrency`), then merged server-side via `FinalizeMultipart`.
4. Comic created as `pending_review`; user returns to the "My Comics" tab.

Desktop may use a native file picker, but the rest of the pipeline is shared code from `packages/ui`.

### Download delivery

- `GET /download/*key` (auth) serves the archive. Files under `download_stream_threshold_mb` are streamed with a `Content-Disposition` filename; larger files 302-redirect to a presigned URL — the object key basename carries the friendly filename (Encore's object API can't sign a `response-content-disposition`).
- The download button records the download (quota check + counter) via `POST /comics/:id/download`, then navigates to the download URL.

### Background-job observability

- A dedicated `jobs` service owns a `job_runs` table and private `RecordJobStart` / `RecordJobFinish` endpoints; cron handlers and pubsub handlers record their runs (best-effort — recording failures never block the job).
- Pubsub subscriptions (`archive-extract`, `ai-moderation`) configure `RetryPolicy{MaxRetries: 5}`; Encore dead-letters messages that exceed retries (the DLQ itself is internal/Cloud-only, so `failed` `job_runs` rows are the app-visible signal).
- The admin "Background Jobs" page reads `GET /admin/jobs` and can manually re-trigger the two cron sweeps.

### Storage usage

- `GET /admin/storage` (upload service) enumerates the `comic-files` bucket via `ComicBucket.List` (object count + bytes, broken down by key prefix) and merges the Cloudflare Images count; the admin "Storage" page renders it. The older `storage_bytes` KPI is the DB `SUM(file_size_bytes)` — the bucket enumeration is the accurate source.

### Admin dashboard (aggregate + SSE)

- A dedicated `dashboard` service aggregates KPI stats from the other services and exposes `GET /admin/dashboard` (snapshot) plus `GET /admin/dashboard-stream` (SSE) — the stream pushes the aggregate on connect and every 15s, recomputing only while a client is connected.
- The per-service stats endpoints (`auth.AdminDashboardStats`/`GetSignupTrend`, `comics.GetComicsStats`, `billing.GetBillingStats`, `reading.GetReadingStats`/`GetDownloadTrend`, `upload.GetStorageStats`) are `private` (service-internal); the admin frontend talks only to the `dashboard` service.
- The admin dashboard keeps an SSE connection (toggleable "Realtime: On/Off", default on) via `EventSource('/api/admin/dashboard-stream')` through the same-origin `/api` proxy (cookie → Bearer).

## Why this split

| Concern                    | Web (SvelteKit)     | Desktop (Wails)          | Shared (`packages/ui`) |
|---------------------------|---------------------|--------------------------|------------------------|
| SEO / public marketing    | Yes                 | No                       | —                      |
| Native OS integration     | Limited             | Yes                      | —                      |
| Comic reader & cards      | Yes                 | Yes                      | Yes                    |
| Uploader 3-tab workspace  | Yes                 | Yes                      | Yes                    |
| Admin control panel       | Yes                 | Yes (same UI)            | Yes                    |
| Business logic & data     | via Encore          | via Encore               | client helpers only    |


## Desktop Capabilities (Wails only)

### Offline library
- User-selected folder; comics stored as `.cbz`.
- Per-comic offline toggle; bulk series / Continue Reading download.
- Optional auto-download next issue; multiple library profiles.
- Local index + file I/O on the **Go side**; UI in Svelte.
- Reader prefers local CBZ when present.

### Native shell
- System tray + minimize-to-tray + optional start with OS.
- Native notifications (downloads, follows, support, payments).
- Global hotkeys.
- Native file / folder dialogs; “Open with” for CBZ/CBR.
- Jump List / dock menu.

### Import & reader
- Drag-and-drop archives or image folders onto the window.
- Fullscreen distraction-free reader (optional dual-page).
- Touch / pen / gamepad page turns.
- Quick Look preview.
- Local reading stats; export/backup of offline library.

Web has none of the above offline/native features.



## Images & media storage

### Asset kinds
Every image asset is classified with a kind flag:
- `cover` – primary cover art
- `preview` – preview/sample pages shown before download or in marketing
- `page` – full comic page (reader)
- `original` – source archive or other non-display binary (S3 only)

### Where assets are stored
| Kind | Storage |
|------|---------|
| `cover` | **Cloudflare Images only** |
| `preview` | **Cloudflare Images only** |
| `page` | **S3-compatible object storage only** |
| archives / originals | **S3 only** |

### How images are served (configurable)
Admin setting `image_serving_mode` (default: `direct`):

| Mode | Behaviour |
|------|-----------|
| `direct` | Use the storage URL as-is (Cloudflare Images URL or S3/signed URL) |
| `imgproxy` | Route through a configured imgproxy (or compatible) base URL |
| `cloudflare_images` | Prefer Cloudflare Images delivery URLs / variants where applicable |

- Default for new installs: **`direct`**.
- Mode is global and changeable in the admin control panel (General / Media settings).
- Page images on S3 may still be passed through imgproxy when mode is `imgproxy` (resize/webp); covers/previews already live on Cloudflare Images.
- Reader and public grids always resolve URLs via a small media helper so switching mode does not require rewriting content.

### imgproxy mode (implemented, deployment optional)

- Admin settings: `imgproxy_base_url`, `imgproxy_key`, `imgproxy_salt` (hex-encoded key/salt, imgproxy `IMGPROXY_KEY`/`IMGPROXY_SALT` convention).
- When `image_serving_mode == "imgproxy"` and a base URL is configured, `ServeMedia` (`GET /media/*key`) 302-redirects to a **signed** imgproxy URL (`buildImgproxyURL`: URL-safe base64 HMAC-SHA256 over `salt || path`, source = the S3 signed URL, processing `rs:fit:2000:2000`). The browser then fetches the CDN directly.
- No live imgproxy deployment is required to ship this — the code + admin config are ready to be wired when a deployment is provisioned.

### Upload pipeline implications
- Manual / archive creation: after presigned flow (or CF Images upload API for cover/preview), the creation payload includes keys **and** kind.
- Cover & preview uploads target Cloudflare Images APIs (or CF-presigned flows), not the general S3 comic-page bucket.
- Page files and the main archive remain on S3 via existing presigned upload sessions.
- `upload_mode` (admin setting, default `backend`) controls the transport:
  - `backend` — cover/preview streamed to Cloudflare Images via `POST /upload/image`; archive/pages streamed to S3 via `POST /upload/file` (multipart).
  - `direct` — cover/preview via `POST /media/cloudflare/upload-url`; archive/pages via presigned upload sessions.
- Large archives are split into parts (see Upload & Creation Flow) and merged server-side with `FinalizeMultipart`.


## Internationalization

- **UI locale** (chrome) vs **content language** (comic) are separate concepts.
- Frontend uses an i18n message catalog (SvelteKit-compatible). Default locale `en`.
- Priority locales for comics engagement: en, ja, es, ko, fr, pt-BR, zh-CN, de, it, id.
- Backend stores `comics.content_language` and `users.ui_locale`; list endpoints support language filters.
- Enabled locales are configurable in admin settings.

### v1.1 i18n foundation

Located in `frontend-public/src/lib/i18n/`:
- `locales.ts` (registry: `Locale`, `ENABLED_LOCALES`, `PRIORITY_LOCALES`, `LOCALE_META`)
- `messages/en.ts` (English catalog — source of truth for `MessageKey`)
- `detect.ts` (pure `detectLocale()`: user → Accept-Language → `en`)
- `index.svelte.ts` (runes store: `state.locale`, `t(key, params)`, `initializeLocale`, `setLocale`)

Locale is resolved server-side in `+layout.server.ts` (cookie / Accept-Language), passed through `data.locale`, and applied via `<html lang>` + `initializeLocale()`. Components translate with `{t('key')}`. English ships in v1.1; new locale packs add a `messages/<code>.ts` catalog + `registerCatalog()`.

## Payments: NowPayments (v1)

### Payment Flow (all screens render in a modal, closable via X button or Esc only)

**Step 1 — Plan Selection**
User browses tier × interval grid on `/pricing`. Each card shows tier name, interval, price, and cumulative features. User picks a plan.

**Step 2 — Crypto Currency Selection**
A grid of crypto currency options (BTC, ETH, USDT, LTC, etc.) is displayed with icons. User selects a currency — selecting one deactivates the others; unselecting clears the estimated price. Backend calls NowPayments live price estimate API for the selected plan + crypto, and the estimated amount is displayed.

**Step 3 — Balance Check**
Backend checks the user's NowPayments balance using `sub_partner_id` (stored on the `users` table).

**Step 4A — User Has Enough Balance**
1. Backend calls NowPayments to create the subscription.
2. **Atomic**: only if NowPayments succeeds, save the subscription locally (rollback on failure).
3. Frontend transitions to "Processing subscription..." screen.
4. **Poll for webhook**: frontend polls the local subscription by ID every few seconds, waiting for the subscription webhook to update status.
5. When subscription is confirmed → re-sign JWT cookie with the new tier and refresh the page.
6. Polling timeout: 5 minutes, then show error with retry.

**Step 4B — User Has NOT Enough Balance**
1. Backend calls NowPayments `deposit with payment`:
   - `ipn_callback_url` = deposit webhook URL with local deposit ID as query parameter.
2. Frontend shows QR code + deposit address (from NowPayments response), plus the amount to send in the selected crypto.
3. **Poll for webhook**: frontend polls the local deposit by ID every few seconds, waiting for the deposit webhook.
4. When deposit is confirmed → transition to Step 4A screen (Processing subscription).
5. Polling timeout: 30 minutes for crypto confirmations, then "Transaction not detected yet" with retry.

### Webhooks

**Subscription Webhook** (`POST /webhooks/nowpayments/subscription`)
- Verify signature (`x-nowpayments-sig` header).
- Store raw payload in `webhook_events`.
- If `payment_status == "finished"`:
  - Set `subscriptions.active = true` for the matching subscription.
  - Update `users.tier` to the tier from the subscription's plan.

**Deposit Webhook** (`POST /webhooks/nowpayments/deposit`)
- Verify signature.
- Store raw payload in `webhook_events`.
- If `payment_status == "finished"`:
  - Update `deposits` table with status = `completed`.

### Provider Abstraction
- `PaymentsProvider` interface lives in the shared `backend/nowpayments` package, with a single v1 implementation: `NowPaymentsProvider`.
- Methods: `EstimatePrice`, `CheckBalance`, `CreateCustomer`, `CreateSubscription`, `CreateDeposit`, `CreatePlan`.
- `auth` and `tiers` import the shared package directly (customer creation + plan auto-link); `billing` obtains plan details via `tiers.GetPlan` and `sub_partner_id` via `auth.EnsureSubPartnerID` rather than reading those services' tables.
- `provider` field on relevant tables (`nowpayments`).

### JWT Authentication (NowPayments API)

Some NowPayments endpoints require a JWT in addition to the `x-api-key`. The JWT is obtained via `POST /v1/auth` (email + password) and expires in 5 minutes.

**Token caching**: The `NowPaymentsProvider` caches the JWT for 4 minutes with a mutex lock. All JWT-requiring methods call `getAuthToken()` internally — no manual token management needed. On expiry, the next call auto-refreshes the token transparently.

| Endpoint | Auth method | Go method |
|----------|:---:|------|
| `GET /v1/estimate` | API key only | `doRequest` |
| `GET /v1/sub-partner/balance/{id}` | API key only | `doRequest` |
| `POST /v1/sub-partner/balance` | JWT | `doJWTRequest` → `CreateCustomer` |
| `POST /v1/subscriptions` | JWT + API key | `doJWTRequest` → `CreateSubscription` |
| `POST /v1/sub-partner/payment` | JWT + API key | `doJWTRequest` → `CreateDeposit` |

**Secrets required:**
- `NowPaymentsAPIKey` — for all requests
- `NowPaymentsIPNKey` — for webhook signature verification
- `NowPaymentsEmail` — dashboard login for JWT auth
- `NowPaymentsPassword` — dashboard password for JWT auth

### NowPayments API Reference

Full API spec: `backend/billing/nowpayments-openapi.yaml`

Key endpoints used in v1:

| Endpoint | Method | Purpose | Auth |
|----------|--------|---------|------|
| `/v1/auth` | POST | Obtain JWT (email+password, 5min expiry) | — |
| `/v1/status` | GET | API availability check | — |
| `/v1/estimate` | GET | Live crypto price estimate | API key |
| `/v1/sub-partner/balance` | POST | Create customer account → `sub_partner_id` | JWT |
| `/v1/sub-partner/balance/{id}` | GET | Check customer balance | API key |
| `/v1/sub-partner/payment` | POST | Deposit with payment (top-up) | JWT + API key |
| `/v1/subscriptions` | POST | Create recurring payment (subscription) | JWT + API key |
| `/v1/subscriptions/plans` | POST/GET | Create/list subscription plans | JWT / API key |

### All polling has timeouts
- Subscription polling: 5 minutes max.
- Deposit polling: 30 minutes max.

## Security & Consistency

- Same Encore Auth for both clients.
- Same role checks.
- Same plan-matrix and pending-review rules.
- Desktop does not embed a second backend; it is a client.
