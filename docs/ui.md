# UI / UX Specification – Comics Galore

## Design Principles

- Public site: fast, content-first, minimal JS (SvelteKit SSR).
- Uploader / admin areas: powerful but clean.
- Full dark mode + WCAG compliance.
- Excellent progress feedback for long uploads.

## Technology

- **Web**: SvelteKit + shadcn-svelte + Tailwind CSS
- **Desktop**: Wails + Svelte + shadcn-svelte (same design system)
- **Shared**: `packages/ui` (components, forms, reader, upload logic, Zod schemas)
- Superforms + Zod, TanStack Table (Svelte), ECharts/Chart.js, libarchive.js
- Superforms + Zod
- TanStack Table (Svelte)
- Apache ECharts or Chart.js
- libarchive.js (client-side archive parsing)

## Uploader – New Comic Workspace (3 tabs)

### Tab 1 – My Comics
- Paginated grid of simplified comic cards (cover, title, status badge, date).
- Newest first.
- Clicking a card can open edit / view status.
- Empty state encourages going to Tab 2 or 3.
- After successful creation the user lands here.

### Tab 2 – Manual Creation
- Superforms form with clear sections.
- Cover + archive file inputs that:
  - Request presigned URL(s) from the backend
  - Upload **directly from browser to S3**
  - Show progress, allow cancel, support resume via Upload Session
- “Save draft / Resume” indicator if an incomplete session exists.
- Submit → creates comic in `pending_review` → redirect to Tab 1.

### Tab 3 – Archive + JSON
- Large drop zone.
- On file select:
  1. libarchive.js parses the archive in the browser.
  2. Looks for `comic.json` / `metadata.json`.
  3. Validates against the public JSON Schema (client-side first, server-side authoritative).
  4. Shows a preview of extracted metadata + file list.
  5. Uploads each required object via presigned URLs (multipart for large files).
- Progress UI for both parsing and uploading.
- Crash recovery via the same Upload Session mechanism.
- Builds the same form payload as the manual path (metadata + file keys).
- Calls the same creation API.
- On success → redirect to Tab 1.

### Shared Upload UX Rules
- Never upload the file body through the Encore backend.
- Clear error messages for invalid JSON, missing files, size limits, network failures.
- Ability to resume an interrupted session after page reload.
- Visual status of the session (uploading, paused, finalising, failed).

## Other Screens (summary)

**Public**: Home, Comic Detail, Series, Reader (keyboard + swipe), Search, Profiles, Pricing, RSS.  
**Authenticated**: Library, Messaging, Support, Account.  
**Moderator**: Moderation queue + pending comic reviews.  
**Admin**: Full control panel with red banner, datalists, settings, plan matrix, charts, etc.

## Component Guidelines

- Prefer shadcn-svelte.
- Comic cards reusable.
- Skeletons + optimistic UI where appropriate.
- Reader: ← → Space Esc + touch swipe.
- All uploads show determinate progress when possible.


## Desktop-only UI (Wails)

### Offline
- Per-comic **“Available offline”** toggle + progress.
- Bulk: “Download series”, “Download Continue Reading”.
- Settings: offline folder picker, multiple library profiles, auto-download next issue, storage warnings.
- Offline library manager (list, size, remove, export/backup).

### Native behaviour
- System tray menu (Continue Reading, active downloads, unread).
- Native notification toasts (download complete, new issues, support replies…).
- Global hotkey settings.
- Drag-and-drop zones on main window and New Comic tab.
- “Open with” / file association feedback.

### Reader polish
- True fullscreen (F11), auto-hide chrome, optional dual-page spread.
- Touch / pen swipe; optional gamepad hints.
- Quick Look overlay (cover + first pages).
- Local session stats visible in a discreet panel.

### Admin Control Panel
- Persistent **red banner** when plan matrix is incomplete
- Dashboard: KPI cards, charts, activity feeds, quick actions
- Datalists with search, filter, **saved views**, **bulk actions**, **CSV export**
- User detail drawer (subs, downloads, flags, tickets); ban/suspend; audited impersonation
- Comics/series/tags management; recycle bin; scheduled list; Staff Picks ordering
- Plan matrix editor; manual subscription grant/extend; coupons; past-due list
- AI settings + AI review queue + decision log
- System health: background jobs, dead-letter retry, storage usage
- Broadcast announcements; email template preview
- Support tickets; audit log; general settings

## Public experience (completed)

- Home rails: Staff Picks, New this week, Popular this month, Comic of the Day, Continue Reading
- Tag pages as public routes
- Series page: progress %, missing issues, series-follow button
- Reader: thumbnails/scrubber, fit modes (width/height/original), optional dual-page
- Age rating badge on cards/detail; age-gate modal when required
- Static routes: Terms, Privacy, DMCA
- Cookie consent banner (if analytics enabled)
- Auth: verify-email, reset-password flows; notification preferences screen
- Empty states + first-time guidance; quota-blocked upgrade state

- Admin media settings: image serving mode (default direct), imgproxy base URL.


## Language & i18n UI

- Locale switcher in header/user menu (enabled locales only).
- Comic forms: required **content language** select (defaults from settings).
- Browse/search: language facet / filter chips.
- Comic cards/detail show content language badge when useful (e.g. mixed-language grids).
- All chrome strings go through i18n catalogs (SvelteKit-friendly i18n library).
- Default locale `en`; fallback chain user → browser → en.

## Tier-gated image gallery

- Main image, dots, thumbnails, fullscreen lightbox (see skill `tier-gated-gallery` and reference TS gallery).
- Indices beyond the tier preview limit are blurred with upgrade CTA (plans modal).
- Locked images must not open sharp fullscreen; prefer withheld or placeholder URLs from the API.
