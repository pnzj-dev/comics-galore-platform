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
├── backend/                 # Encore Go API (single source of truth)
├── frontend/                # SvelteKit web application
├── desktop/                 # Wails v2 application
│   ├── main.go / app.go     # Go side (bindings, native menus, etc.)
│   └── frontend/            # Vite + Svelte app that uses packages/ui
└── packages/
    └── ui/                  # Shared Svelte code
        ├── components/
        ├── forms/
        ├── lib/             # Zod, stores, upload-session, API helpers
        └── ...
```

## Shared Layer (`packages/ui`)

Must contain (at minimum):

- shadcn-svelte based primitives and domain components (ComicCard, Reader, PlanModal, etc.)
- Superforms + Zod schemas for comic creation, login, settings…
- Upload Session client logic (presign → upload → collect file keys → build payload)
- Comic creation payload builder (used by both Manual and Archive tabs)
- Reader chrome (keyboard + swipe handlers)
- Theme / dark-mode stores
- Thin wrappers around the generated Encore client

Web and desktop both import from this package.  
Duplication of UI or validation logic is considered a bug.

## Web Client (SvelteKit)

- Full SSR / progressive enhancement for public pages.
- Route groups: `(public)`, `(app)`, `(uploader)`, `(moderator)`, `(admin)`.
- Uses `packages/ui` for almost all visual and form code.
- Handles SEO, RSS, Open Graph, legal pages, age gate, tag pages, series follow, notification preferences.

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
- `PaymentsProvider` interface with single v1 implementation: `NowPaymentsProvider`.
- Methods: `EstimatePrice`, `CheckBalance`, `CreateSubscription`, `CreateDeposit`.
- `provider` field on relevant tables (`nowpayments`).

### NowPayments API Reference

Full API spec: `backend/billing/nowpayments-openapi.yaml`

Key endpoints used in v1:

| Endpoint | Method | Purpose | Auth |
|----------|--------|---------|------|
| `/v1/status` | GET | API availability check | — |
| `/v1/estimate` | GET | Live crypto price estimate | API key |
| `/v1/payment` | POST | Create payment (returns pay_address + pay_amount) | API key |
| `/v1/payment/{id}` | GET | Get payment status (polling) | API key |
| `/v1/invoice` | POST | Create invoice (returns invoice_url) | API key |
| `/v1/sub-partner/balance` | POST | Create customer account → `sub_partner_id` | JWT |
| `/v1/sub-partner/balance/{id}` | GET | Check customer balance | API key |
| `/v1/sub-partner/payment` | POST | Deposit with payment (top-up) | JWT + API key |
| `/v1/subscriptions` | POST | Create recurring payment (subscription) | JWT + API key |
| `/v1/subscriptions/plans` | POST/GET | Create/list subscription plans | JWT / API key |
| `/v1/subscriptions/{id}` | GET/DELETE | Get/delete recurring payment | API key / JWT |

### All polling has timeouts
- Subscription polling: 5 minutes max.
- Deposit polling: 30 minutes max.

## Security & Consistency

- Same Encore Auth for both clients.
- Same role checks.
- Same plan-matrix and pending-review rules.
- Desktop does not embed a second backend; it is a client.
