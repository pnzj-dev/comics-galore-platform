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

## Payments: single provider (v1), multi-provider-ready

- **v1 production**: NowPayments only (crypto).
- Subscription / deposit / webhook flows go through a **PaymentsProvider** interface.
- Sole implementation in v1: `NowPaymentsProvider`.
- DB stores `provider` + provider-specific external IDs; `webhook_events` keeps full raw payloads.
- Plan matrix (tiers × intervals) is owned by the app; the provider only executes payment operations.
- Admin can show a provider setting (one option today) without redesign later.
- User-facing copy: “Pay with crypto” rather than hard-wiring a single brand in every screen.

## Security & Consistency

- Same Encore Auth for both clients.
- Same role checks.
- Same plan-matrix and pending-review rules.
- Desktop does not embed a second backend; it is a client.
