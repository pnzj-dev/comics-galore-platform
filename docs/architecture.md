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
- Route groups: `(public)`, `(app)`, `(auth)`.
- Uses `packages/ui` for almost all visual and form code.
- Handles home, browse, detail, reader, series, tags, pricing, upload, settings, auth, SEO, RSS, Open Graph, legal pages, age gate.
- **Server-side route guards**: `+layout.server.ts` and `+page.server.ts` load functions read the JWT from a cookie and `throw redirect(302)` before the page component renders. The `(app)` group requires authentication; the `upload` page requires `uploader`/`admin` role. See the cookie-based auth section below.

### Cookie-based Auth

The auth system uses a **non-HttpOnly cookie** (`token`) for JWT storage, enabling both:

| Consumer | How it reads the token |
|----------|----------------------|
| SvelteKit server (SSR) | `cookies.get('token')` in `+layout.server.ts` / `+page.server.ts` — validates before rendering |
| Browser API client | `document.cookie.match('token=...')` — sends as `Authorization: Bearer` to Encore |
| Auth store (`login`/`logout`) | `document.cookie = 'token=...'` — sets/clears on login/logout |

The cookie is set with `path=/`, `SameSite=Lax`, and `max-age=2592000` (30 days). Server-side guards decode the JWT payload to check roles before page rendering. Client-side `onMount` → `goto()` redirects have been removed from all protected pages.

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
- `reading_lists` / `reading_list_items` (public shelves) in comics; `GET /comics/:id/related` (co-occurrence, no new table).

## Upload & Creation Flow (identical on web and desktop)

1. Create Upload Session.
2. Obtain presigned URLs.
3. Upload assets (browser or Wails webview → S3).
4. Receive file keys.
5. Build the **same form payload** (manual UI or archive+libarchive.js).
6. `POST /comics` → comic created as `pending_review`.
7. Redirect / navigate to “My Comics” tab.

Desktop may use a native file picker to choose the archive, but the rest of the pipeline is shared code from `packages/ui`.

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

### Upload pipeline implications
- Manual / archive creation: after presigned flow (or CF Images upload API for cover/preview), the creation payload includes keys **and** kind.
- Cover & preview uploads target Cloudflare Images APIs (or CF-presigned flows), not the general S3 comic-page bucket.
- Page files and the main archive remain on S3 via existing presigned upload sessions.


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
