# Product Specification – Comics Galore

## Vision

Comics Galore is a modern, subscription-based platform for discovering, reading, discussing, and downloading digital comics.  
It is monetised exclusively through cryptocurrency subscriptions via NowPayments and includes human + optional AI moderation, internal messaging, a full support system, and a feature-rich admin control panel.

The product is delivered as:
- a **web application** (SvelteKit) optimised for discovery, SEO and reading;
- a **desktop application** (Wails + Svelte) that reuses the same UI components and flows.

Both clients share the majority of Svelte code via an internal `packages/ui` library and talk to the same Encore backend.

## Roles

| Role        | Key Capabilities |
|-------------|------------------|
| `user`      | Browse, read, like, favorite, rate, comment, download (quota), message, open support tickets, flag content |
| `uploader`  | Everything a user can do + **New Comic** workspace (3 tabs), manage series, analytics, schedule, tier-gate |
| `moderator` | Everything a user can do + moderate comments, handle flags, review pending comics (accesses the admin app but is restricted to the Moderation page only) |
| `admin`     | Full control panel, settings, plan matrix, AI config, users, comics, webhooks, support, etc. |

## Uploader: “New Comic” Workspace (3 tabs)

When the user has the `uploader` role, a dedicated **New Comic** area is available with exactly three tabs. **Tab state is driven by query parameters** (`?tab=list`, `?tab=manual`, `?tab=archive`) for bookmarkability and browser back/forward support. Active upload sessions are loaded server-side in `+page.server.ts` alongside the comics list.

### Tab 1 – My Comics (default after success)
- Grid of **simplified comic cards** belonging to the current uploader.
- Ordered by creation date, **newest → oldest**.
- **Pagination** (or infinite scroll with clear page controls).
- Shows status badge (`pending_review`, `published`, `rejected`, etc.).
- After a successful creation (manual or archive) the user is automatically returned to this tab (`/upload?tab=list`) so they can see the new comic.

### Tab 2 – Manual Creation
- **Full-width 3-section layout**: Row 1 — metadata (left, `1fr`) + 3:4 cover dropzone (right, 320px); Row 2 — preview image grid (Cloudflare upload, min 2, up to 10); Row 3 — archive file grid (S3 upload, min 1, up to 10).
- Rich form fields: title, author, description, synopsis, category, tags, language, age rating.
- **All file inputs upload directly to S3 via presigned URLs**. The frontend receives **file keys** back.
- Those keys are written into the form state.
- When the form is complete and valid, the frontend sends the **same creation payload** to the backend.
- Supports crash recovery via an **Upload Session**.
- On success the comic is created with status `pending_review` and the user is sent to Tab 1.

### Tab 3 – Archive + Metadata JSON
- User selects a single archive (`.zip`, `.cbz`, `.cbr`, `.rar`…).
- **Frontend** uses **libarchive.js** to parse the archive in the browser.
- Frontend locates and validates `comic.json` / `metadata.json`.
- Individual assets are uploaded via presigned URLs; the frontend receives **file keys**.
- The archive path **automatically builds the exact same form payload** that the manual form would produce (title, description, series, cover_key, file_key, page_keys, tags, …).
- No user intervention is required after the archive is chosen (unless validation fails).
- The frontend then calls the **same comic-creation API** as the manual path.
- Same crash-recovery mechanism and same redirect to Tab 1 on success.

### Upload Session / Crash Recovery
- Long-running uploads (especially large archives) must survive browser crashes, network drops, or accidental page closes.
- Flow:
  1. Frontend asks backend to create an `upload_session` (or `comic_draft`).
  2. Backend returns session ID + presigned URLs (or multipart upload IDs).
  3. Frontend uploads parts, periodically reports progress / completed parts.
  4. On page reload the frontend can resume the session.
  5. When all parts are confirmed, frontend finalises → comic moves to `pending_review`.
- After successful finalisation the user is redirected to **Tab 1**.

### Pre-publication Validation
- Every newly created comic (both modes) starts in **`pending_review`**.
- It is **not** publicly visible.
- Validation path:
  1. Optional AI/LLM moderation (if enabled in settings).
  2. Human administrator or moderator review.
- Only after explicit approval does the status become `published` and the comic appears on the public site.
- Rejection is possible with a reason visible to the uploader.

## Public-Facing Experience (feature-complete)

### Discovery
- Fast comic grids with progressive / responsive covers
- Staff Picks, Comic of the Day / Random (displayed with `HeroComicCard` — landscape layout, cover + info side by side)
- “New this week” / “Popular this month” rails
- Faceted search; **tag pages** as first-class public URLs
- Browse page search bar (title/author/description) + popular tag pills
- “People also liked” related comics

### Series
- Series pages with reading order, **progress %**, missing-issue gaps
- **Series follow** (in addition to uploader follow)
- “Next issue” / Continue series CTAs

### Reader
- Keyboard + swipe; progress saving; Continue Reading shelf
- **Page thumbnails / scrubber**
- **Fit modes** (width / height / original)
- Optional simple dual-page on large screens
- Clear reading direction (LTR default, documented)

### Social & library
- Like, favorite, rate, comments, follow uploaders + series
- Dedicated **Favorites** page (authenticated) listing all favorited comics with inline unfavorite
- Shareable public reading lists / shelves
- Flagging / report flow

### Trust & compliance
- **Content rating / age rating** on comics (e.g. all-ages, teen, mature)
- **Age gate** when mature content is present (or site-wide policy)
- **Forbid mature for free users** (admin setting) — when enabled, free-tier and anonymous users cannot access mature/explicit comics at all (lists, detail, reader, download). Detail shows a blurred cover + upgrade CTA; paid tiers and staff are unaffected, and the age gate is skipped for blocked users.
- Static pages: **Terms**, **Privacy**, **DMCA / copyright** complaint
- Cookie / consent banner if non-essential analytics are used

### Account hygiene
- Email verification on signup
- Password reset / magic link (per Encore Auth capabilities)
- **Notification preferences** (which email / in-app events the user wants)
- Polished empty states + first-time guidance
- Clear quota-blocked / upgrade states opening the plans modal

### SEO & share
- RSS (site + per series)
- Strong Open Graph / social cards

## Subscription Model (Critical)

**Payment provider (v1):** NowPayments only. The system is designed multi-provider-ready (provider interface + stored `provider` field + raw webhooks) so a second crypto processor can be added later without changing the upgrade UX or plan matrix.


- Global Intervals × Tiers = strict NowPayments plan matrix.
- Incomplete matrix → permanent red banner in admin.
- Multi-step upgrade modal (plan → currency → balance/deposit QR → create subscription → status webhook + timeout).
- Full raw webhook audit trail.

## AI Moderation

- Optional, configurable model + prompt, or fully disabled.
- Used both for comments and for new comic review.
- Human moderators / admins always have final authority.

## Admin Control Panel (feature-complete)

### Dashboard
- KPI cards: users, new users, active subscribers, MRR, comics, downloads, views, open tickets, pending flags, online users, comments
- Charts: revenue, engagement, user growth, tier distribution, top content, support volume
- Live activity feeds + quick action panels (moderation, support, pending reviews)

### Content & operations
- Searchable/filterable datalists: users, comics, series, tags, webhooks, subscription attempts
- **Bulk actions** on comics/users (publish, reject, feature, change tier, ban…)
- Soft-deleted **recycle bin** + restore
- Scheduled comics list / calendar
- Tag & series management (CRUD, merge tags)
- Staff Picks / homepage featured ordering

### Users & risk
- User detail drawer: subscriptions, downloads, flags, tickets, messages
- Ban / suspend with reason + duration
- **Impersonate user** (audited)
- Multi-flag / abuse score indicator

### Billing
- Plan matrix editor + red banner until complete
- Manual grant / extend / revoke subscription (comp, support)
- Simple coupon / promo codes
- Failed / past-due payments list + retry
- Revenue by tier × interval breakdown
- Force-sync subscription from NowPayments

### AI & automation
- AI moderation enable/model/prompt settings
- AI review queue (uncertain decisions)
- AI decision log (why auto-approved/rejected)
- Toggle AI per content type (comics vs comments)

### System health
- Background job status (extraction, webhooks, AI, quota reset)
- Failed jobs / dead-letter + retry
- Storage usage approximation (S3)
- Recent error signals when available

### Communications
- Broadcast / announcement to all users or by tier
- Email template preview (Resend)
- Notification delivery log

### Other
- General settings (maintenance, registration, contact email, hide mature default, enable comments, default meta description, quotas, rate limit, image serving — editable via Form/JSON toggle)
- Full audit log
- CSV export on main datalists
- Saved filters / views on datalists

## Bootstrap Rule

System refuses to start if no admin user exists (clear log message).


## Desktop Application Features (Wails only)

### Offline library
- User-chosen local folder for offline comics.
- Per-comic **“Make available offline”** / remove.
- Stored as local **`.cbz`** files; reader opens them directly (no network).
- Offline library manager: list, total size, change folder, free-space warnings.
- Bulk actions: download entire series, download Continue Reading list.
- Optional auto-download of next issue when finishing a comic.
- Multiple offline libraries / profiles (e.g. SSD vs archive drive).

### Native desktop behaviour
- System tray: minimize to tray, tray menu (Continue Reading, downloads in progress, unread items).
- Optional “Start with OS”.
- Native OS notifications (new followed uploads, offline download complete, support replies, payment events).
- Global / media-style hotkeys (open app, Continue Reading, next offline comic).
- “Open with Comics Galore” for `.cbz` / `.cbr` files.
- Jump List / dock menu (recent comics, Continue Reading).

### Import & reader polish
- Drag-and-drop: drop archive onto window → open in reader or start New Comic archive flow; drop folder of images → quick local CBZ for offline.
- True fullscreen distraction-free reader (auto-hide cursor, borderless, optional dual-page spread).
- Touch / pen / convertible gestures; optional gamepad page turns.
- Quick Look / preview (cover + first pages without full reader).
- Local reading stats (time, pages/session; optional later sync).
- Export / backup offline library + reading progress; export a series as CBZ folder.

Web client does **not** implement full-comic offline storage or native OS integrations.


## Media storage & delivery

- **Covers** and **previews** are stored on **Cloudflare Images only** (kind: `cover` | `preview`).
- **Comic pages** and source archives are stored on **S3-compatible storage only** (kind: `page` | `original`).
- **Image serving mode** is configurable in the admin panel:
  - `direct` (default) – use storage URLs as-is
  - `imgproxy` – deliver via configured image proxy
  - `cloudflare_images` – prefer Cloudflare Images delivery/variants
- Every image asset carries a **kind** flag so the system never confuses covers, previews, and pages.


## Preview gallery (tier-gated)

- Comic detail may show a multi-image gallery/lightbox.
- Number of sharp, viewable images is limited by the user’s tier (`max_preview_pages` / preview perk).
- Remaining images are shown blurred with an upgrade invitation (plans modal).
- Prefer not exposing full-resolution URLs for locked images.

## Language & internationalization

### Comic content language
- Every comic has a required **content language** (ISO 639-1 code, e.g. `en`, `ja`, `es`).
- Used for filtering, search facets, series browsing, and recommendations.
- Upload form (manual and archive metadata JSON) must set or default language.
- Public UI can filter by language; default browse may prefer the user’s UI locale then English.

### UI internationalization (i18n)
The product UI is internationalized to maximize reach for a **comics** audience.

**Priority locales (v1 set to implement as translations become available):**

| Priority | Locale | Rationale |
|----------|--------|-----------|
| 1 | `en` | Default; largest global digital comics / web audience |
| 2 | `ja` | Manga |
| 3 | `es` | Spain + Latin America |
| 4 | `ko` | Manhwa / webtoon |
| 5 | `fr` | Franco-Belgian BD tradition |
| 6 | `pt-BR` | Brazil |
| 7 | `zh-CN` | Manhua / Chinese web comics |
| 8 | `de` | Large EU comics market |
| 9 | `it` | Strong local comics culture |
| 10 | `id` | Large webtoon / digital comics readership |

- Default UI locale: **`en`**.
- Locale detection: user preference → Accept-Language → `en`.
- User profile stores preferred UI locale.
- Admin can enable/disable locales and set default.
- Legal pages and emails should follow the same locale strategy where practical.
- **Content language** (on the comic) is independent of **UI locale** (chrome, buttons, settings).

## Out of Scope (v1)

- Creator payouts, native apps, DRM, multi-language, real-time WebSocket presence
